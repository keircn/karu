package cmd

import (
	"fmt"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/internal/player"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/internal/ui"
	"github.com/keircn/karu/internal/workflow"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for anime by title",
	Long:  `Search for anime by title and interactively select episodes to watch.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var query string
		if len(args) > 0 {
			query = args[0]
		}

		autoQuality, _ := cmd.Flags().GetBool("auto-quality")

		selection, err := workflow.GetAnimeSelection(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("You chose: %s\n", selection.Anime.Title)

		episode, err := ui.SelectEpisode(selection.Episodes)
		if err != nil {
			fmt.Printf("Error selecting episode: %v\n", err)
			return
		}

		if episode != nil {
			fmt.Printf("You chose episode: %s\n", *episode)

			scraper.PreloadAdjacentEpisodes(selection.ShowID, selection.Episodes, *episode)

			if autoQuality {
				cfg, _ := config.Load()
				videoURL, err := scraper.GetVideoURLWithQuality(selection.ShowID, *episode, cfg.Quality)
				if err != nil {
					fmt.Printf("Error getting video URL: %v\n", err)
					return
				}

				if err := player.Play(videoURL); err != nil {
					fmt.Printf("Error playing video: %v\n", err)
				}
				return
			}

			qualities, err := scraper.GetAvailableQualities(selection.ShowID, *episode)
			if err != nil {
				fmt.Printf("Error getting video qualities: %v\n", err)
				return
			}

			selectedQuality, err := ui.SelectQuality(qualities)
			if err != nil {
				fmt.Printf("Error selecting quality: %v\n", err)
				return
			}

			if selectedQuality == nil {
				fmt.Println("No quality selected.")
				return
			}

			fmt.Printf("Selected quality: %s\n", selectedQuality.Quality)

			if err := player.Play(selectedQuality.URL); err != nil {
				fmt.Printf("Error playing video: %v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolP("auto-quality", "a", false, "Automatically select quality based on config")
}
