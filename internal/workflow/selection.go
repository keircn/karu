package workflow

import (
	"fmt"
	"strings"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/internal/ui"
)

type AnimeSelection struct {
	Anime    *scraper.Anime
	ShowID   string
	Episodes []string
}

func GetAnimeSelection(query string) (*AnimeSelection, error) {
	if query == "" {
		var err error
		query, err = ui.PromptForSearch()
		if err != nil {
			return nil, fmt.Errorf("getting search query: %w", err)
		}
		if query == "" {
			return nil, fmt.Errorf("no search query provided")
		}
	}

	fmt.Printf("Searching for: %s...\n", query)
	animes, err := scraper.Search(query)
	if err != nil {
		return nil, fmt.Errorf("searching for anime: %w", err)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("no anime found")
	}

	choice, err := ui.SelectAnime(animes)
	if err != nil {
		return nil, fmt.Errorf("selecting anime: %w", err)
	}

	if choice == nil {
		return nil, fmt.Errorf("no anime selected")
	}

	showID := choice.URL[strings.LastIndex(choice.URL, "/")+1:]
	fmt.Printf("Loading episodes for %s...\n", choice.Title)
	episodes, err := scraper.GetEpisodes(showID)
	if err != nil {
		return nil, fmt.Errorf("getting episodes: %w", err)
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes found for this anime")
	}

	history, _ := config.LoadHistory()
	if history != nil {
		history.AddEntry(query, choice.Title, choice.URL, len(episodes))
	}

	return &AnimeSelection{
		Anime:    choice,
		ShowID:   showID,
		Episodes: episodes,
	}, nil
}

func GetAnimeSelectionFromHistory() (*AnimeSelection, error) {
	option, err := ui.ShowHistoryOptions()
	if err != nil {
		return nil, fmt.Errorf("showing history options: %w", err)
	}

	if option == "New search" {
		return GetAnimeSelection("")
	}

	entry, err := ui.SelectFromHistory(option)
	if err != nil {
		return nil, fmt.Errorf("selecting from history: %w", err)
	}

	if entry.Title == "" {
		return GetAnimeSelection("")
	}

	showID := entry.URL[strings.LastIndex(entry.URL, "/")+1:]
	fmt.Printf("Loading episodes for %s...\n", entry.Title)
	episodes, err := scraper.GetEpisodes(showID)
	if err != nil {
		return nil, fmt.Errorf("getting episodes: %w", err)
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes found for this anime")
	}

	history, _ := config.LoadHistory()
	if history != nil {
		history.UpdateProgress(entry.Title, entry.LastWatched)
	}

	return &AnimeSelection{
		Anime:    &scraper.Anime{Title: entry.Title, URL: entry.URL},
		ShowID:   showID,
		Episodes: episodes,
	}, nil
}
