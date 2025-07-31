package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingTop(1)
	HelpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type SelectableItem interface {
	list.Item
	GetValue() interface{}
}

type GenericItem struct {
	title       string
	description string
	value       interface{}
}

func NewGenericItem(title, description string, value interface{}) GenericItem {
	return GenericItem{
		title:       title,
		description: description,
		value:       value,
	}
}

func (i GenericItem) Title() string         { return i.title }
func (i GenericItem) Description() string   { return i.description }
func (i GenericItem) FilterValue() string   { return i.title }
func (i GenericItem) GetValue() interface{} { return i.value }

type ListModel struct {
	list     list.Model
	choice   interface{}
	quitting bool
}

func NewListModel(items []list.Item, title string) ListModel {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.Styles.Title = TitleStyle
	l.Styles.PaginationStyle = PaginationStyle
	l.Styles.HelpStyle = HelpStyle

	return ListModel{list: l}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := AppStyle.GetFrameSize()
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
			if item, ok := m.list.SelectedItem().(SelectableItem); ok {
				m.choice = item.GetValue()
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	if m.choice != nil || m.quitting {
		return ""
	}
	return AppStyle.Render(m.list.View())
}

func (m ListModel) GetChoice() interface{} {
	return m.choice
}

func RunSelection(model ListModel) (interface{}, error) {
	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	return finalModel.(ListModel).GetChoice(), nil
}
