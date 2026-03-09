package handler

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EditTiempoBucle struct {
	Input     textinput.Model
	Cancelled bool
}

func NewEditTiempoBucle(placeholder string) EditTiempoBucle {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 50
	ti.Focus()
	return EditTiempoBucle{Input: ti}
}

func (m EditTiempoBucle) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditTiempoBucle) GetTiempoBucle() (string, bool) {
	return m.Input.Value(), m.Cancelled
}

func (m EditTiempoBucle) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ResetInactivity()
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

func (m EditTiempoBucle) View() string {
	return "Nuevo tiempo de espera entre bucles(Maximo 10): " + m.Input.View() + "\nEnter = guardar   q = cancelar"
}
