package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/keircn/karu/pkg/ui"
)

type BrowseMode string

const (
	BrowseModeSearch   BrowseMode = "search"
	BrowseModePopular  BrowseMode = "catalog"
	BrowseModeTrending BrowseMode = "recent"
)

func SelectBrowseMode() (*BrowseMode, error) {
	items := []list.Item{
		ui.NewGenericItem("Search for anime", string(BrowseModeSearch), BrowseModeSearch),
		ui.NewGenericItem("Browse recent anime", string(BrowseModeTrending), BrowseModeTrending),
		ui.NewGenericItem("Browse anime catalog", string(BrowseModePopular), BrowseModePopular),
	}

	model := ui.NewListModel(items, "Browse Anime")
	result, err := ui.RunSelection(model)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	mode := result.(BrowseMode)
	return &mode, nil
}
