package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/keircn/karu/internal/config"
)

type episodeItem struct {
	title      string
	watched    bool
	episodeNum int
}

func (i episodeItem) Title() string {
	if i.watched {
		return fmt.Sprintf("✓ %s", i.title)
	}
	return i.title
}

func (i episodeItem) Description() string {
	if i.watched {
		return "Watched"
	}
	return ""
}

func (i episodeItem) FilterValue() string { return i.title }

type episodeModel struct {
	list      list.Model
	choice    *string
	quitting  bool
	showTitle string
	hasResume bool
	resumeEp  int
}

func NewEpisodeModel(episodes []string, showTitle string) episodeModel {
	history, _ := config.LoadHistory()

	items := make([]list.Item, len(episodes))
	var hasResume bool
	var resumeEp int

	for i, ep := range episodes {
		episodeNum := len(episodes) - i
		watched := history.IsWatched(showTitle, episodeNum)

		if !hasResume {
			if lastWatched, exists := history.GetProgress(showTitle); exists {
				if episodeNum == lastWatched+1 {
					hasResume = true
					resumeEp = episodeNum
				}
			}
		}

		items[i] = list.Item(episodeItem{
			title:      ep,
			watched:    watched,
			episodeNum: episodeNum,
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select an episode"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return episodeModel{
		list:      l,
		showTitle: showTitle,
		hasResume: hasResume,
		resumeEp:  resumeEp,
	}
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

		case "r":
			if m.hasResume {
				resumeEpisode := fmt.Sprintf("%d", m.resumeEp)
				m.choice = &resumeEpisode
				return m, tea.Quit
			}

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

	view := appStyle.Render(m.list.View())

	if m.hasResume {
		view += fmt.Sprintf("\n\nPress 'r' to resume from episode %d", m.resumeEp)
	}

	return view
}

func SelectEpisode(episodes []string, showTitle string) (*string, error) {
	for i, j := 0, len(episodes)-1; i < j; i, j = i+1, j-1 {
		episodes[i], episodes[j] = episodes[j], episodes[i]
	}

	m := NewEpisodeModel(episodes, showTitle)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	return finalModel.(episodeModel).choice, nil
}
