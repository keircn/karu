package workflow

import (
	"fmt"
	"strings"

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
	episodes, err := scraper.GetEpisodes(showID)
	if err != nil {
		return nil, fmt.Errorf("getting episodes: %w", err)
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes found for this anime")
	}

	return &AnimeSelection{
		Anime:    choice,
		ShowID:   showID,
		Episodes: episodes,
	}, nil
}
