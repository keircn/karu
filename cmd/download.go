package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/internal/ui"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [query]",
	Short: "Download anime episodes",
	Args:  cobra.MaximumNArgs(1),
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

			episode, err := ui.SelectEpisode(episodes)
			if err != nil {
				fmt.Printf("Error selecting episode: %v\n", err)
				return
			}

			if episode != nil {
				fmt.Printf("Downloading episode: %s\n", *episode)

				cfg, err := config.Load()
				if err != nil {
					cfg = &config.DefaultConfig
				}

				filename := fmt.Sprintf("%s_episode_%s.mp4",
					strings.ReplaceAll(choice.Title, " ", "_"), *episode)
				outputPath := filepath.Join(cfg.DownloadDir, filename)

				if err := scraper.DownloadEpisodeWithProgress(showID, *episode, outputPath); err != nil {
					fmt.Printf("Error downloading episode: %v\n", err)
					return
				}
			}
		}
	},
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
	downloadCmd.AddCommand(downloadListCmd)
	downloadCmd.AddCommand(downloadCleanCmd)
	rootCmd.AddCommand(downloadCmd)
}
