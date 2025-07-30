package workflow

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/internal/scraper"
)

type DownloadOptions struct {
	All       bool
	Range     string
	OutputDir string
}

type DownloadResult struct {
	Total      int
	Successful int
	Failed     int
	Episodes   []string
}

func ParseEpisodeRange(rangeStr string, availableEpisodes []string) ([]string, error) {
	if err := ValidateEpisodeRange(rangeStr, len(availableEpisodes)); err != nil {
		return nil, err
	}

	if strings.Contains(rangeStr, "-") {
		return parseRangeSelection(rangeStr, availableEpisodes)
	}
	return parseListSelection(rangeStr, availableEpisodes)
}

func parseRangeSelection(rangeStr string, availableEpisodes []string) ([]string, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format, use start-end (e.g., 1-5)")
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start episode number: %w", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end episode number: %w", err)
	}

	if start > end {
		return nil, fmt.Errorf("start episode must be less than or equal to end episode")
	}

	var result []string
	for _, ep := range availableEpisodes {
		epNum, err := strconv.Atoi(ep)
		if err != nil {
			continue
		}
		if epNum >= start && epNum <= end {
			result = append(result, ep)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no episodes found in range %d-%d", start, end)
	}

	return result, nil
}

func parseListSelection(rangeStr string, availableEpisodes []string) ([]string, error) {
	episodes := strings.Split(rangeStr, ",")
	var result []string
	for _, ep := range episodes {
		ep = strings.TrimSpace(ep)
		if slices.Contains(availableEpisodes, ep) {
			result = append(result, ep)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid episodes found in list: %s", rangeStr)
	}

	return result, nil
}

func DownloadEpisodes(selection *AnimeSelection, opts DownloadOptions) (*DownloadResult, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.DefaultConfig
	}

	if opts.OutputDir != "" {
		cfg.DownloadDir = opts.OutputDir
	}

	var episodesToDownload []string

	if opts.All {
		episodesToDownload = selection.Episodes
	} else if opts.Range != "" {
		episodesToDownload, err = ParseEpisodeRange(opts.Range, selection.Episodes)
		if err != nil {
			return nil, fmt.Errorf("parsing episode range: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no download options specified")
	}

	result := &DownloadResult{
		Total:    len(episodesToDownload),
		Episodes: episodesToDownload,
	}

	for i, episode := range episodesToDownload {
		if len(episodesToDownload) > 1 {
			fmt.Printf("\n[%d/%d] Downloading episode: %s\n", i+1, len(episodesToDownload), episode)
		} else {
			fmt.Printf("Downloading episode: %s\n", episode)
		}

		filename := fmt.Sprintf("%s_episode_%s.mp4",
			strings.ReplaceAll(selection.Anime.Title, " ", "_"), episode)
		outputPath := filepath.Join(cfg.DownloadDir, filename)

		if err := scraper.DownloadEpisodeWithProgress(selection.ShowID, episode, outputPath); err != nil {
			fmt.Printf("Error downloading episode %s: %v\n", episode, err)
			result.Failed++
			continue
		}
		result.Successful++
	}

	return result, nil
}

func PrintDownloadSummary(result *DownloadResult) {
	if result.Total > 1 {
		fmt.Printf("\nDownload summary: %d/%d episodes downloaded successfully",
			result.Successful, result.Total)
		if result.Failed > 0 {
			fmt.Printf(" (%d failed)", result.Failed)
		}
		fmt.Println()
	}
}
