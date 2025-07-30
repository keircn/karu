package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keircn/karu/internal/scraper"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingTop(1)

	helpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type item struct {
	title, url string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.url }
func (i item) FilterValue() string { return i.title }

type model struct {
	list     list.Model
	choice   *scraper.Anime
	quitting bool
}

func NewModel(animes []scraper.Anime) model {
	items := make([]list.Item, len(animes))
	for i, anime := range animes {
		items[i] = list.Item(item{title: anime.Title, url: anime.URL})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select an anime"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{list: l}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &scraper.Anime{Title: i.title, URL: i.url}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != nil {
		return ""
	}
	if m.quitting {
		return ""
	}
	return appStyle.Render(m.list.View())
}

func SelectAnime(animes []scraper.Anime) (*scraper.Anime, error) {
	m := NewModel(animes)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	return finalModel.(model).choice, nil
}
