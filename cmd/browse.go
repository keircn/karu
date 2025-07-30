package cmd

import (
	"fmt"

	"github.com/keircn/karu/internal/player"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/internal/ui"
	"github.com/keircn/karu/internal/workflow"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse anime by category",
	Long:  `Interactive browsing interface for discovering anime by different categories.`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, err := ui.SelectBrowseMode()
		if err != nil {
			fmt.Printf("Error selecting browse mode: %v\n", err)
			return
		}

		if mode == nil {
			return // User cancelled
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
		Episodes: nil, // Will be loaded when needed
	}
}

func extractShowID(url string) string {
	// Extract show ID from URL: get everything after the last slash
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			return url[i+1:]
		}
	}
	return url // fallback
}

func handleEpisodeSelection(selection *workflow.AnimeSelection) {
	// Load episodes if not already loaded
	if selection.Episodes == nil {
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
}

func init() {
	rootCmd.AddCommand(browseCmd)
}
