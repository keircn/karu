package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/keircn/karu/internal/scraper"
	"github.com/keircn/karu/pkg/ui"
)

type qualityItem struct {
	option scraper.QualityOption
}

func (i qualityItem) Title() string {
	title := i.option.Quality
	if title == "" {
		title = "Auto"
	}
	if i.option.Source != "" {
		title += " (" + i.option.Source + ")"
	}
	return title
}

func (i qualityItem) Description() string {
	desc := ""
	if i.option.IsHLS {
		desc += "HLS"
	} else {
		desc += "Direct"
	}
	return desc
}

func (i qualityItem) FilterValue() string { return i.Title() }

func (i qualityItem) GetValue() interface{} { return i.option }

func SelectQuality(qualities *scraper.QualityChoice) (*scraper.QualityOption, error) {
	if qualities == nil || len(qualities.Options) == 0 {
		return nil, nil
	}

	if len(qualities.Options) == 1 {
		return &qualities.Options[0], nil
	}

	items := make([]list.Item, len(qualities.Options))
	for i, option := range qualities.Options {
		items[i] = qualityItem{option: option}
	}

	model := ui.NewListModel(items, "Select video quality")
	result, err := ui.RunSelection(model)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	option := result.(scraper.QualityOption)
	return &option, nil
}
