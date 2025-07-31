package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/pkg/ui"
)

type historyItem struct {
	entry config.HistoryEntry
}

func (i historyItem) FilterValue() string {
	return i.entry.Title + " " + i.entry.Query
}

func (i historyItem) Title() string {
	progress := ""
	if i.entry.TotalEps > 0 {
		progress = fmt.Sprintf(" (%d/%d)", i.entry.LastWatched, i.entry.TotalEps)
	}
	return i.entry.Title + progress
}

func (i historyItem) Description() string {
	timeAgo := formatTimeAgo(i.entry.Timestamp)
	accessInfo := fmt.Sprintf("Watched %d times • %s", i.entry.AccessCount, timeAgo)
	return accessInfo
}

type HistoryModel struct {
	list     list.Model
	selected config.HistoryEntry
	quitting bool
}

func NewHistoryModel(entries []config.HistoryEntry) HistoryModel {
	items := make([]list.Item, len(entries))
	for i, entry := range entries {
		items[i] = historyItem{entry: entry}
	}

	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Search History"
	l.SetShowStatusBar(true)
	l.Styles.Title = ui.TitleStyle
	l.Styles.PaginationStyle = ui.PaginationStyle
	l.Styles.HelpStyle = ui.HelpStyle

	return HistoryModel{list: l}
}

func (m HistoryModel) Init() tea.Cmd {
	return nil
}

func (m HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(historyItem); ok {
				m.selected = i.entry
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m HistoryModel) View() string {
	if m.quitting {
		return ""
	}

	return docStyle.Render(m.list.View())
}

func (m HistoryModel) GetSelected() config.HistoryEntry {
	return m.selected
}

type HistoryOptionsModel struct {
	cursor   int
	options  []string
	selected string
	quitting bool
}

func NewHistoryOptionsModel() HistoryOptionsModel {
	return HistoryOptionsModel{
		options: []string{
			"Recent searches",
			"Most watched",
			"Search history",
			"New search",
		},
	}
}

func (m HistoryOptionsModel) Init() tea.Cmd {
	return nil
}

func (m HistoryOptionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = m.options[m.cursor]
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m HistoryOptionsModel) View() string {
	if m.quitting {
		return ""
	}

	s := "What would you like to do?\n\n"

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			option = selectedStyle.Render(option)
		} else {
			option = unselectedStyle.Render(option)
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}

	s += "\n" + ui.HelpStyle.Render("j/k: move • enter: select • q: quit")

	return docStyle.Render(s)
}

func (m HistoryOptionsModel) GetSelected() string {
	return m.selected
}

func SelectFromHistory(viewType string) (config.HistoryEntry, error) {
	history, err := config.LoadHistory()
	if err != nil {
		return config.HistoryEntry{}, err
	}

	var entries []config.HistoryEntry

	switch viewType {
	case "Recent searches":
		entries = history.GetRecent(20)
	case "Most watched":
		entries = history.GetMostWatched(20)
	case "Search history":
		entries = history.GetRecent(0)
	default:
		return config.HistoryEntry{}, fmt.Errorf("invalid view type")
	}

	if len(entries) == 0 {
		return config.HistoryEntry{}, fmt.Errorf("no history entries found")
	}

	m := NewHistoryModel(entries)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return config.HistoryEntry{}, err
	}

	if historyModel, ok := finalModel.(HistoryModel); ok {
		return historyModel.GetSelected(), nil
	}

	return config.HistoryEntry{}, fmt.Errorf("no selection made")
}

func ShowHistoryOptions() (string, error) {
	m := NewHistoryOptionsModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if optionsModel, ok := finalModel.(HistoryOptionsModel); ok {
		return optionsModel.GetSelected(), nil
	}

	return "", fmt.Errorf("no selection made")
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)
