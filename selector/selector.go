package selector

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list   list.Model
	choice list.Item
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.choice = m.list.SelectedItem()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Selector(items []list.Item, prompt string) (*list.Item, error) {
	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = prompt

	p := tea.NewProgram(model{list: list}, tea.WithAltScreen())

	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("unable to show selector: %w", err)
	}

	if m, ok := m.(model); ok && m.choice != nil {
		return &m.choice, nil
	}

	return nil, nil
}
