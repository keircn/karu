package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/internal/ui"
	"github.com/spf13/cobra"
)

var (
	downloadAll   bool
	downloadRange string
)

var downloadCmd = &cobra.Command{
	Use:   "download [query]",
	Short: "Download anime episodes",
	Long: `Download anime episodes with options for single episodes, ranges, or entire series.

Examples:
  karu download                        # Interactive single episode download
  karu download "bocchi"               # Search and download single episode
  karu download --all "bocchi"         # Download all episodes
  karu download --range "1-5" "bocchi" # Download episodes 1-5
  karu download --range "1,3,5" "bocchi" # Download specific episodes`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var query string
		if len(args) == 0 {
			var err error
			query, err = ui.PromptForSearch()
			if err != nil {
				fmt.Printf("Error getting search query: %v\n", err)
				return
			}
			if query == "" {
				fmt.Println("No search query provided.")
				return
			}
		} else {
			query = args[0]
		}

		animes, err := scraper.Search(query)
		if err != nil {
			fmt.Printf("Error searching for anime: %v\n", err)
			return
		}

		if len(animes) == 0 {
			fmt.Println("No anime found.")
			return
		}

		choice, err := ui.SelectAnime(animes)
		if err != nil {
			fmt.Printf("Error selecting anime: %v\n", err)
			return
		}

		if choice != nil {
			fmt.Printf("You chose: %s\n", choice.Title)
			showID := choice.URL[strings.LastIndex(choice.URL, "/")+1:]
			episodes, err := scraper.GetEpisodes(showID)
			if err != nil {
				fmt.Printf("Error getting episodes: %v\n", err)
				return
			}

			if len(episodes) == 0 {
				fmt.Println("No episodes found for this anime.")
				return
			}

			cfg, err := config.Load()
			if err != nil {
				cfg = &config.DefaultConfig
			}

			var episodesToDownload []string

			if downloadAll {
				episodesToDownload = episodes
				fmt.Printf("Downloading all %d episodes of %s\n", len(episodes), choice.Title)
			} else if downloadRange != "" {
				episodesToDownload, err = parseEpisodeRange(downloadRange, episodes)
				if err != nil {
					fmt.Printf("Error parsing episode range: %v\n", err)
					return
				}
				fmt.Printf("Downloading episodes %s of %s\n", downloadRange, choice.Title)
			} else {
				episode, err := ui.SelectEpisode(episodes)
				if err != nil {
					fmt.Printf("Error selecting episode: %v\n", err)
					return
				}
				if episode != nil {
					episodesToDownload = []string{*episode}
				}
			}

			if len(episodesToDownload) == 0 {
				return
			}

			failed := 0
			for i, episode := range episodesToDownload {
				if len(episodesToDownload) > 1 {
					fmt.Printf("\n[%d/%d] Downloading episode: %s\n", i+1, len(episodesToDownload), episode)
				} else {
					fmt.Printf("Downloading episode: %s\n", episode)
				}

				filename := fmt.Sprintf("%s_episode_%s.mp4",
					strings.ReplaceAll(choice.Title, " ", "_"), episode)
				outputPath := filepath.Join(cfg.DownloadDir, filename)

				if err := scraper.DownloadEpisodeWithProgress(showID, episode, outputPath); err != nil {
					fmt.Printf("Error downloading episode %s: %v\n", episode, err)
					failed++
					continue
				}
			}

			if len(episodesToDownload) > 1 {
				fmt.Printf("\nDownload summary: %d/%d episodes downloaded successfully",
					len(episodesToDownload)-failed, len(episodesToDownload))
				if failed > 0 {
					fmt.Printf(" (%d failed)", failed)
				}
				fmt.Println()
			}
		}
	},
}

func parseEpisodeRange(rangeStr string, availableEpisodes []string) ([]string, error) {
	if strings.Contains(rangeStr, "-") {
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format, use start-end (e.g., 1-5)")
		}

		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start episode number: %v", err)
		}

		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end episode number: %v", err)
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

	episodes := strings.Split(rangeStr, ",")
	var result []string
	for _, ep := range episodes {
		ep = strings.TrimSpace(ep)
		for _, availableEp := range availableEpisodes {
			if availableEp == ep {
				result = append(result, ep)
				break
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid episodes found in list: %s", rangeStr)
	}

	return result, nil
}

var downloadListCmd = &cobra.Command{
	Use:   "list",
	Short: "List downloaded episodes",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			cfg = &config.DefaultConfig
		}

		files, err := filepath.Glob(filepath.Join(cfg.DownloadDir, "*.mp4"))
		if err != nil {
			fmt.Printf("Error listing downloads: %v\n", err)
			return
		}

		if len(files) == 0 {
			fmt.Println("No downloaded episodes found.")
			return
		}

		fmt.Println("Downloaded episodes:")
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			fmt.Printf("  %s (%.2f MB) - %s\n",
				filepath.Base(file),
				float64(info.Size())/(1024*1024),
				info.ModTime().Format("2006-01-02 15:04"))
		}
	},
}

var downloadCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all downloaded episodes",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			cfg = &config.DefaultConfig
		}

		files, err := filepath.Glob(filepath.Join(cfg.DownloadDir, "*.mp4"))
		if err != nil {
			fmt.Printf("Error listing downloads: %v\n", err)
			return
		}

		if len(files) == 0 {
			fmt.Println("No downloaded episodes found.")
			return
		}

		fmt.Printf("This will remove %d downloaded episodes. Continue? (y/N): ", len(files))
		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return
		}

		removed := 0
		for _, file := range files {
			if err := os.Remove(file); err == nil {
				removed++
			}
		}

		fmt.Printf("Removed %d downloaded episodes.\n", removed)
	},
}

func init() {
	downloadCmd.Flags().BoolVarP(&downloadAll, "all", "a", false, "Download all episodes")
	downloadCmd.Flags().StringVarP(&downloadRange, "range", "r", "", "Download episode range (e.g., 1-5 or 1,3,5)")

	downloadCmd.AddCommand(downloadListCmd)
	downloadCmd.AddCommand(downloadCleanCmd)
	rootCmd.AddCommand(downloadCmd)
}
