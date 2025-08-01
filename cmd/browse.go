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

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse anime by category",
	Long:  `Interactive browsing interface for discovering anime by search, recent releases, or catalog.`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, err := ui.SelectBrowseMode()
		if err != nil {
			fmt.Printf("Error selecting browse mode: %v\n", err)
			return
		}

		if mode == nil {
			return
		}

		switch *mode {
		case ui.BrowseModeSearch:
			handleSearchMode()
		case ui.BrowseModeTrending:
			handleTrendingMode()
		case ui.BrowseModePopular:
			handlePopularMode()
		}
	},
}

func handleSearchMode() {
	selection, err := workflow.GetAnimeSelection("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("You chose: %s\n", selection.Anime.Title)
	handleEpisodeSelection(selection)
}

func handleTrendingMode() {
	fmt.Println("Loading recent anime...")

	animes, err := scraper.GetTrending()
	if err != nil {
		fmt.Printf("Error getting recent anime: %v\n", err)
		return
	}

	if len(animes) == 0 {
		fmt.Println("No recent anime found.")
		return
	}

	choice, err := ui.SelectAnime(animes)
	if err != nil {
		fmt.Printf("Error selecting anime: %v\n", err)
		return
	}

	if choice != nil {
		selection := createSelectionFromAnime(choice)
		fmt.Printf("You chose: %s\n", choice.Title)
		handleEpisodeSelection(selection)
	}
}

func handlePopularMode() {
	fmt.Println("Loading anime catalog...")

	animes, err := scraper.GetPopular()
	if err != nil {
		fmt.Printf("Error getting anime catalog: %v\n", err)
		return
	}

	if len(animes) == 0 {
		fmt.Println("No anime found in catalog.")
		return
	}

	choice, err := ui.SelectAnime(animes)
	if err != nil {
		fmt.Printf("Error selecting anime: %v\n", err)
		return
	}

	if choice != nil {
		selection := createSelectionFromAnime(choice)
		fmt.Printf("You chose: %s\n", choice.Title)
		handleEpisodeSelection(selection)
	}
}

func createSelectionFromAnime(choice *scraper.Anime) *workflow.AnimeSelection {
	showID := extractShowID(choice.URL)
	return &workflow.AnimeSelection{
		Anime:    choice,
		ShowID:   showID,
		Episodes: nil,
	}
}

func extractShowID(url string) string {
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			return url[i+1:]
		}
	}
	return url
}

func handleEpisodeSelection(selection *workflow.AnimeSelection) {
	if selection.Episodes == nil {
		fmt.Printf("Loading episodes for %s...\n", selection.Anime.Title)
		episodes, err := scraper.GetEpisodes(selection.ShowID)
		if err != nil {
			fmt.Printf("Error getting episodes: %v\n", err)
			return
		}

		if len(episodes) == 0 {
			fmt.Println("No episodes found for this anime.")
			return
		}

		selection.Episodes = episodes
	}

	episode, err := ui.SelectEpisode(selection.Episodes, selection.Anime.Title)
	if err != nil {
		fmt.Printf("Error selecting episode: %v\n", err)
		return
	}

	if episode != nil {
		fmt.Printf("You chose episode: %s\n", *episode)

		cfg, _ := config.Load()

		fmt.Printf("Getting video source for episode %s...\n", *episode)
		videoURL, err := scraper.GetVideoURL(selection.ShowID, *episode)
		if err != nil {
			fmt.Printf("Error getting video URL: %v\n", err)
			return
		}

		if videoURL == "" {
			fmt.Println("No video URL found for this episode.")
			return
		}

		fmt.Printf("Video source found! Starting playback...\n")
		if cfg.AutoPlayNext {
			fmt.Printf("Auto-play next episode: %s\n", getAutoPlayStatus(cfg.AutoPlayNext))

			playbackInfo := &player.PlaybackInfo{
				ShowID:    selection.ShowID,
				ShowTitle: selection.Anime.Title,
				Episodes:  selection.Episodes,
				Current:   *episode,
				VideoURL:  videoURL,
			}

			getVideoURLFunc := func(showID, ep string) (string, error) {
				fmt.Printf("Getting next episode source...\n")
				return scraper.GetVideoURL(showID, ep)
			}

			if err := player.PlayWithAutoNext(playbackInfo, getVideoURLFunc); err != nil {
				fmt.Printf("Error playing video: %v\n", err)
			}
		} else {
			if err := player.Play(videoURL); err != nil {
				fmt.Printf("Error playing video: %v\n", err)
			}
		}
	}
}
func init() {
	rootCmd.AddCommand(browseCmd)
}
