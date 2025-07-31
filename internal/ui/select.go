package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/pkg/ui"
)

func SelectAnime(animes []scraper.Anime) (*scraper.Anime, error) {
	items := make([]list.Item, len(animes))
	for i, anime := range animes {
		items[i] = ui.NewGenericItem(anime.Title, anime.URL, anime)
	}

	model := ui.NewListModel(items, "Select an anime")
	result, err := ui.RunSelection(model)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	anime := result.(scraper.Anime)
	return &anime, nil
}
