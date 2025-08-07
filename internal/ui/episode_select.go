package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/pkg/ui"
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

func (i episodeItem) GetValue() interface{} { return i.title }

type episodeModel struct {
	ui.ListModel
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

		items[i] = episodeItem{
			title:      ep,
			watched:    watched,
			episodeNum: episodeNum,
		}
	}

	baseModel := ui.NewListModel(items, "Select an episode")

	return episodeModel{
		ListModel: baseModel,
		showTitle: showTitle,
		hasResume: hasResume,
		resumeEp:  resumeEp,
	}
}

func (m episodeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "r":
			if m.hasResume {
				return m, tea.Quit
			}
		}
	}

	baseModel, cmd := m.ListModel.Update(msg)
	m.ListModel = baseModel.(ui.ListModel)
	return m, cmd
}

func (m episodeModel) View() string {
	baseView := m.ListModel.View()
	if baseView == "" {
		return ""
	}

	if m.hasResume {
		baseView += fmt.Sprintf("\n\nPress 'r' to resume from episode %d", m.resumeEp)
	}

	return baseView
}

func SelectEpisode(episodes []string, showTitle string) (*string, error) {
	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes available")
	}

	for i, j := 0, len(episodes)-1; i < j; i, j = i+1, j-1 {
		episodes[i], episodes[j] = episodes[j], episodes[i]
	}

	m := NewEpisodeModel(episodes, showTitle)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("episode selection failed: %v", err)
	}

	episodeModel, ok := finalModel.(episodeModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type in episode selection")
	}

	result := episodeModel.GetChoice()
	if result == nil {
		return nil, nil
	}

	episode, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid episode selection")
	}

	if episode == "" {
		return nil, fmt.Errorf("empty episode selected")
	}

	return &episode, nil
}
