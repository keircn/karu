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
		useHistory, _ := cmd.Flags().GetBool("history")

		var selection *workflow.AnimeSelection
		var err error

		if useHistory {
			selection, err = workflow.GetAnimeSelectionFromHistory()
		} else {
			selection, err = workflow.GetAnimeSelection(query)
		}

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

			cfg, _ := config.Load()

			if autoQuality {
				fmt.Printf("Getting video source for episode %s...\n", *episode)
				videoURL, err := scraper.GetVideoURLWithQuality(selection.ShowID, *episode, cfg.Quality)
				if err != nil {
					fmt.Printf("Error getting video URL: %v\n", err)
					return
				}

				fmt.Printf("Video source found! Starting playback...\n")
				if cfg.AutoPlayNext {
					fmt.Printf("Auto-play next episode: %s\n", getAutoPlayStatus(cfg.AutoPlayNext))

					playbackInfo := &player.PlaybackInfo{
						ShowID:   selection.ShowID,
						Episodes: selection.Episodes,
						Current:  *episode,
						VideoURL: videoURL,
					}

					getVideoURLFunc := func(showID, ep string) (string, error) {
						fmt.Printf("Getting next episode source...\n")
						return scraper.GetVideoURLWithQuality(showID, ep, cfg.Quality)
					}

					if err := player.PlayWithAutoNext(playbackInfo, getVideoURLFunc); err != nil {
						fmt.Printf("Error playing video: %v\n", err)
					}
				} else {
					if err := player.Play(videoURL); err != nil {
						fmt.Printf("Error playing video: %v\n", err)
					}
				}
				return
			}

			fmt.Printf("Loading available qualities for episode %s...\n", *episode)
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
			fmt.Printf("Starting playback...\n")

			if cfg.AutoPlayNext {
				fmt.Printf("Auto-play next episode: %s\n", getAutoPlayStatus(cfg.AutoPlayNext))

				playbackInfo := &player.PlaybackInfo{
					ShowID:   selection.ShowID,
					Episodes: selection.Episodes,
					Current:  *episode,
					VideoURL: selectedQuality.URL,
				}

				getVideoURLFunc := func(showID, ep string) (string, error) {
					fmt.Printf("Getting next episode source...\n")
					qualities, err := scraper.GetAvailableQualities(showID, ep)
					if err != nil {
						return "", err
					}
					for _, q := range qualities.Options {
						if q.Quality == selectedQuality.Quality {
							return q.URL, nil
						}
					}
					if len(qualities.Options) > 0 {
						return qualities.Options[0].URL, nil
					}
					return "", fmt.Errorf("no video URL found")
				}

				if err := player.PlayWithAutoNext(playbackInfo, getVideoURLFunc); err != nil {
					fmt.Printf("Error playing video: %v\n", err)
				}
			} else {
				if err := player.Play(selectedQuality.URL); err != nil {
					fmt.Printf("Error playing video: %v\n", err)
				}
			}
		}
	},
}

func getAutoPlayStatus(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolP("auto-quality", "a", false, "Automatically select quality based on config")
	searchCmd.Flags().BoolP("history", "H", false, "Browse search history instead of searching")
}
