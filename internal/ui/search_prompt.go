package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type searchModel struct {
	textInput textinput.Model
	err       error
	query     string
	quitting  bool
}

func initialSearchModel() searchModel {
	ti := textinput.New()
	ti.Placeholder = "Enter anime name to search..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return searchModel{
		textInput: ti,
		err:       nil,
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.query = strings.TrimSpace(m.textInput.Value())
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m searchModel) View() string {
	if m.quitting {
		return ""
	}

	return fmt.Sprintf(
		"\n%s\n\n%s\n\n%s\n",
		focusedStyle.Render("What anime would you like to search for?"),
		m.textInput.View(),
		blurredStyle.Render("Press Enter to search or Esc to quit"),
	) + "\n"
}

func PromptForSearch() (string, error) {
	p := tea.NewProgram(initialSearchModel(), tea.WithOutput(os.Stderr))
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if model, ok := m.(searchModel); ok {
		return model.query, nil
	}

	return "", nil
}
