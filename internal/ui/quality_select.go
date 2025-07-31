package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/keircn/karu/internal/scraper"
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

type qualityModel struct {
	list     list.Model
	choice   *scraper.QualityOption
	quitting bool
}

func NewQualityModel(qualities *scraper.QualityChoice) qualityModel {
	items := make([]list.Item, len(qualities.Options))
	for i, option := range qualities.Options {
		items[i] = list.Item(qualityItem{option: option})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select video quality"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	if qualities.Default < len(items) {
		l.Select(qualities.Default)
	}

	return qualityModel{list: l}
}

func (m qualityModel) Init() tea.Cmd {
	return nil
}

func (m qualityModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(qualityItem)
			if ok {
				m.choice = &i.option
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m qualityModel) View() string {
	if m.choice != nil {
		return ""
	}
	if m.quitting {
		return ""
	}
	return appStyle.Render(m.list.View())
}

func SelectQuality(qualities *scraper.QualityChoice) (*scraper.QualityOption, error) {
	if qualities == nil || len(qualities.Options) == 0 {
		return nil, nil
	}

	if len(qualities.Options) == 1 {
		return &qualities.Options[0], nil
	}

	m := NewQualityModel(qualities)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	return finalModel.(qualityModel).choice, nil
}
