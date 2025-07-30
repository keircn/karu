package cmd

import (
	"fmt"
	"strings"

	"github.com/keircn/karu/internal/player"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/internal/ui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for an anime",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
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
				fmt.Printf("You chose episode: %s\n", *episode)
				videoURL, err := scraper.GetVideoURL(showID, *episode)
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
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
