package handler

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EditEndpointModel struct {
	Input     textinput.Model
	Cancelled bool
}

func NewEditEndpointModel(placeholder string) EditEndpointModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 50
	ti.Focus()
	return EditEndpointModel{Input: ti}
}

func (m EditEndpointModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditEndpointModel) GetEndpoint() (string, bool) {
	return m.Input.Value(), m.Cancelled
}

func (m EditEndpointModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Cancelled = true
			return m, tea.Quit
		case "enter":
			m.Cancelled = false
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m EditEndpointModel) View() string {
	return "Nuevo endpoint: " + m.Input.View() + "\nEnter = guardar   q = cancelar"
}
