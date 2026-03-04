package handler

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EditNumMaxPorBucle struct {
	Input     textinput.Model
	Cancelled bool
}

func NewEditNumMaxPorBucle(placeholder string) EditNumMaxPorBucle {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 50
	ti.Focus()
	return EditNumMaxPorBucle{Input: ti}
}

func (m EditNumMaxPorBucle) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditNumMaxPorBucle) GetNumMaxPorBucle() (string, bool) {
	return m.Input.Value(), m.Cancelled
}

func (m EditNumMaxPorBucle) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Cancelled = true
			return m, tea.Quit
		case "enter", " ":
			m.Cancelled = false
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m EditNumMaxPorBucle) View() string {
	return "Nuevo numero maximo de solicitudes por bucle(Maximo 5): " + m.Input.View() + "\nEnter o Espacio = guardar   q = volver"
}
