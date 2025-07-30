package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
)

type BrowseMode string

const (
	BrowseModeSearch   BrowseMode = "search"
	BrowseModePopular  BrowseMode = "popular"
	BrowseModeTrending BrowseMode = "trending"
)

type browseItem struct {
	title string
	mode  BrowseMode
}

func (i browseItem) Title() string       { return i.title }
func (i browseItem) Description() string { return string(i.mode) }
func (i browseItem) FilterValue() string { return i.title }

type browseModel struct {
	list     list.Model
	choice   *BrowseMode
	quitting bool
}

func NewBrowseModel() browseModel {
	items := []list.Item{
		browseItem{title: "🔍 Search for anime", mode: BrowseModeSearch},
		browseItem{title: "📈 Browse trending anime", mode: BrowseModeTrending},
		browseItem{title: "⭐ Browse popular anime", mode: BrowseModePopular},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Browse Anime"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return browseModel{list: l}
}

func (m browseModel) Init() tea.Cmd {
	return nil
}

func (m browseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(browseItem)
			if ok {
				m.choice = &i.mode
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m browseModel) View() string {
	if m.choice != nil {
		return ""
	}
	if m.quitting {
		return ""
	}
	return appStyle.Render(m.list.View())
}

func SelectBrowseMode() (*BrowseMode, error) {
	m := NewBrowseModel()
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	return finalModel.(browseModel).choice, nil
}
