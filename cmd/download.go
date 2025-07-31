package cmd

import (
	"fmt"

	"github.com/keircn/karu/internal/ui"
	"github.com/keircn/karu/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	downloadAll   bool
	downloadRange string
)

var downloadCmd = &cobra.Command{
	Use:   "download [query]",
	Short: "Download anime episodes",
	Long:  `Download anime episodes with options for single episodes, ranges, or entire series.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var query string
		if len(args) > 0 {
			query = args[0]
		}

		selection, err := workflow.GetAnimeSelection(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("You chose: %s\n", selection.Anime.Title)

		if downloadAll {
			fmt.Printf("Downloading all %d episodes of %s\n", len(selection.Episodes), selection.Anime.Title)
			opts := workflow.DownloadOptions{All: true}
			result, err := workflow.DownloadEpisodes(selection, opts)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			workflow.PrintDownloadSummary(result)
		} else if downloadRange != "" {
			fmt.Printf("Downloading episodes %s of %s\n", downloadRange, selection.Anime.Title)
			opts := workflow.DownloadOptions{Range: downloadRange}
			result, err := workflow.DownloadEpisodes(selection, opts)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			workflow.PrintDownloadSummary(result)
		} else {
			episode, err := ui.SelectEpisode(selection.Episodes, selection.Anime.Title)
			if err != nil {
				fmt.Printf("Error selecting episode: %v\n", err)
				return
			}

			if episode != nil {
				opts := workflow.DownloadOptions{Range: *episode}
				result, err := workflow.DownloadEpisodes(selection, opts)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
				workflow.PrintDownloadSummary(result)
			}
		}
	},
}

var downloadListCmd = &cobra.Command{
	Use:   "list",
	Short: "List downloaded episodes",
	Long:  `List all downloaded episodes with file sizes and dates.`,
	Run: func(cmd *cobra.Command, args []string) {
		fm, err := workflow.NewFileManager()
		if err != nil {
			fmt.Printf("Error initializing file manager: %v\n", err)
			return
		}

		downloads, err := fm.ListDownloads()
		if err != nil {
			fmt.Printf("Error listing downloads: %v\n", err)
			return
		}

		if len(downloads) == 0 {
			fmt.Println("No downloaded episodes found.")
			return
		}

		fmt.Println("Downloaded episodes:")
		for _, download := range downloads {
			fmt.Printf("  %s (%.2f MB) - %s\n",
				download.Name,
				float64(download.Size)/(1024*1024),
				download.Modified.Format("2006-01-02 15:04"))
		}
	},
}

var downloadCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all downloaded episodes",
	Long:  `Remove all downloaded episodes from your download directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fm, err := workflow.NewFileManager()
		if err != nil {
			fmt.Printf("Error initializing file manager: %v\n", err)
			return
		}

		downloads, err := fm.ListDownloads()
		if err != nil {
			fmt.Printf("Error listing downloads: %v\n", err)
			return
		}

		if len(downloads) == 0 {
			fmt.Println("No downloaded episodes found.")
			return
		}

		fmt.Printf("This will remove %d downloaded episodes. Continue? (y/N): ", len(downloads))
		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return
		}

		removed, err := fm.CleanDownloads()
		if err != nil {
			fmt.Printf("Error cleaning downloads: %v\n", err)
			return
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
