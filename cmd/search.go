package cmd

import (
	"fmt"

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
			videoURL, err := scraper.GetVideoURL(selection.ShowID, *episode)
			if err != nil {
				fmt.Printf("Error getting video URL: %v\n", err)
				return
			}

			if videoURL == "" {
				fmt.Println("No video URL found for this episode.")
				return
			}

			if err := player.Play(videoURL); err != nil {
				fmt.Printf("Error playing video: %v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
