package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
)

type episodeItem struct {
	title string
}

func (i episodeItem) Title() string       { return i.title }
func (i episodeItem) Description() string { return "" }
func (i episodeItem) FilterValue() string { return i.title }

type episodeModel struct {
	list     list.Model
	choice   *string
	quitting bool
}

func NewEpisodeModel(episodes []string) episodeModel {
	items := make([]list.Item, len(episodes))
	for i, ep := range episodes {
		items[i] = list.Item(episodeItem{title: ep})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select an episode"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return episodeModel{list: l}
}

func (m episodeModel) Init() tea.Cmd {
	return nil
}

func (m episodeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(episodeItem)
			if ok {
				m.choice = &i.title
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m episodeModel) View() string {
	if m.choice != nil {
		return ""
	}
	if m.quitting {
		return ""
	}
	return appStyle.Render(m.list.View())
}

func SelectEpisode(episodes []string) (*string, error) {
	for i, j := 0, len(episodes)-1; i < j; i, j = i+1, j-1 {
		episodes[i], episodes[j] = episodes[j], episodes[i]
	}

	m := NewEpisodeModel(episodes)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	return finalModel.(episodeModel).choice, nil
}
