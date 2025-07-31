package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/keircn/karu/internal/config"
)

type QualityOption struct {
	Quality string
	URL     string
	Source  string
	IsHLS   bool
}

type QualityChoice struct {
	Options []QualityOption
	Default int
}

type clockResponse struct {
	Links []struct {
		Link string `json:"link"`
		Mp4  bool   `json:"mp4"`
	} `json:"links"`
	EpisodeIframe string `json:"episodeIframe"`
}

func fetchClockURL(clockURL string) (string, error) {
	cfg, _ := config.Load()
	timeout := time.Duration(cfg.RequestTimeout) * time.Second
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", clockURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0")
	req.Header.Set("Referer", "https://allanime.to")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch clock url: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var cr clockResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("failed to unmarshal clock response: %w", err)
	}

	if cr.EpisodeIframe != "" {
		return cr.EpisodeIframe, nil
	}

	if len(cr.Links) > 0 {
		return cr.Links[0].Link, nil
	}

	return "", fmt.Errorf("no iframe URL or direct links found in clock response")
}

func fetchIframeAndExtractStreams(iframeURL string) ([]Stream, error) {
	cfg, _ := config.Load()
	timeout := time.Duration(cfg.RequestTimeout) * time.Second
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", iframeURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0")
	req.Header.Set("Referer", "https://allanime.to")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch iframe url: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`(?s)const\s+streams\s*=\s*(\[.*?\]);`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		re2 := regexp.MustCompile(`(?s)streams\s*=\s*(\[.*?\])`)
		matches2 := re2.FindSubmatch(body)
		if len(matches2) < 2 {
			return nil, fmt.Errorf("could not find streams in iframe response")
		}
		matches = matches2
	}

	var streams []Stream
	if err := json.Unmarshal(matches[1], &streams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal streams from iframe: %w", err)
	}

	return streams, nil
}

func GetVideoURL(showID, episode string) (string, error) {
	videoResult, err := getVideoSourceURLs(showID, episode)
	if err != nil {
		return "", err
	}

	for _, source := range videoResult.Data.Episode.SourceUrls {
		if !strings.HasPrefix(source.SourceUrl, "--") {
			continue
		}

		deobfuscatedPath, err := Deobfuscate(source.SourceUrl[2:])
		if err != nil {
			continue
		}

		var clockURL string
		if strings.HasPrefix(deobfuscatedPath, "https://") {
			clockURL = deobfuscatedPath
		} else {
			clockURL = "https://allanime.day" + deobfuscatedPath
		}

		iframeURL, err := fetchClockURL(clockURL)
		if err != nil {
			continue
		}

		if strings.Contains(iframeURL, ".mp4") || strings.Contains(iframeURL, "sharepoint.com") {
			return iframeURL, nil
		}

		streams, err := fetchIframeAndExtractStreams(iframeURL)
		if err != nil {
			continue
		}

		for _, stream := range streams {
			if !stream.Hls {
				return stream.Link, nil
			}
		}
		if len(streams) > 0 {
			return streams[0].Link, nil
		}
	}

	return "", fmt.Errorf("no playable video URL found")
}

func parseQualityFromResolution(resolutionStr string) int {
	if resolutionStr == "" {
		return 0
	}

	resolutionStr = strings.ToLower(resolutionStr)
	if strings.Contains(resolutionStr, "1080") {
		return 1080
	}
	if strings.Contains(resolutionStr, "720") {
		return 720
	}
	if strings.Contains(resolutionStr, "480") {
		return 480
	}
	if strings.Contains(resolutionStr, "360") {
		return 360
	}

	re := regexp.MustCompile(`(\d+)p?`)
	matches := re.FindStringSubmatch(resolutionStr)
	if len(matches) > 1 {
		if quality, err := strconv.Atoi(matches[1]); err == nil {
			return quality
		}
	}

	return 0
}

func GetAvailableQualities(showID, episode string) (*QualityChoice, error) {
	videoResult, err := getVideoSourceURLs(showID, episode)
	if err != nil {
		return nil, err
	}

	var allOptions []QualityOption
	qualityMap := make(map[string]QualityOption)

	for _, source := range videoResult.Data.Episode.SourceUrls {
		if !strings.HasPrefix(source.SourceUrl, "--") {
			continue
		}

		deobfuscatedPath, err := Deobfuscate(source.SourceUrl[2:])
		if err != nil {
			continue
		}

		var clockURL string
		if strings.HasPrefix(deobfuscatedPath, "https://") {
			clockURL = deobfuscatedPath
		} else {
			clockURL = "https://allanime.day" + deobfuscatedPath
		}

		iframeURL, err := fetchClockURL(clockURL)
		if err != nil {
			continue
		}

		if strings.Contains(iframeURL, ".mp4") || strings.Contains(iframeURL, "sharepoint.com") {
			option := QualityOption{
				Quality: "Auto",
				URL:     iframeURL,
				Source:  source.SourceName,
				IsHLS:   false,
			}
			qualityMap["auto"] = option
			continue
		}

		streams, err := fetchIframeAndExtractStreams(iframeURL)
		if err != nil {
			continue
		}

		for _, stream := range streams {
			quality := parseQualityFromResolution(stream.ResolutionStr)
			qualityKey := fmt.Sprintf("%dp", quality)
			if quality == 0 {
				qualityKey = "auto"
			}

			option := QualityOption{
				Quality: stream.ResolutionStr,
				URL:     stream.Link,
				Source:  stream.SourceName,
				IsHLS:   stream.Hls,
			}

			if existing, exists := qualityMap[qualityKey]; !exists || (!existing.IsHLS && stream.Hls) {
				qualityMap[qualityKey] = option
			}
		}
	}

	for _, option := range qualityMap {
		allOptions = append(allOptions, option)
	}

	if len(allOptions) == 0 {
		return nil, fmt.Errorf("no video sources found")
	}

	sort.Slice(allOptions, func(i, j int) bool {
		qualityI := parseQualityFromResolution(allOptions[i].Quality)
		qualityJ := parseQualityFromResolution(allOptions[j].Quality)

		if qualityI == 0 && qualityJ != 0 {
			return false
		}
		if qualityJ == 0 && qualityI != 0 {
			return true
		}

		return qualityI > qualityJ
	})

	cfg, _ := config.Load()
	preferredQuality := cfg.Quality
	defaultIndex := 0

	for i, option := range allOptions {
		if strings.Contains(strings.ToLower(option.Quality), strings.ToLower(preferredQuality)) {
			defaultIndex = i
			break
		}
	}

	return &QualityChoice{
		Options: allOptions,
		Default: defaultIndex,
	}, nil
}

func GetVideoURLWithQuality(showID, episode, preferredQuality string) (string, error) {
	qualities, err := GetAvailableQualities(showID, episode)
	if err != nil {
		return "", err
	}

	if len(qualities.Options) == 0 {
		return "", fmt.Errorf("no video sources available")
	}

	if preferredQuality == "" {
		return qualities.Options[qualities.Default].URL, nil
	}

	for _, option := range qualities.Options {
		if strings.Contains(strings.ToLower(option.Quality), strings.ToLower(preferredQuality)) {
			return option.URL, nil
		}
	}

	return qualities.Options[qualities.Default].URL, nil
}
