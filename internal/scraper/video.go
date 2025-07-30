package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type clockResponse struct {
	Links []struct {
		Link string `json:"link"`
		Mp4  bool   `json:"mp4"`
	} `json:"links"`
	EpisodeIframe string `json:"episodeIframe"`
}

func fetchClockURL(clockURL string) (string, error) {
	req, err := http.NewRequest("GET", clockURL, nil)
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
	req, err := http.NewRequest("GET", iframeURL, nil)
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
		// this is a really stupid way of doing this but it fucking works
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
