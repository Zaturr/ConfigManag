package handler

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EditIPModel struct {
	Input     textinput.Model
	Cancelled bool
}

func NewEditIPModel(placeholder string) EditIPModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 50
	ti.Focus()
	return EditIPModel{Input: ti}
}

func (m EditIPModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditIPModel) GetIP() (string, bool) {
	return m.Input.Value(), m.Cancelled
}

func (m EditIPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m EditIPModel) View() string {
	return "Nuevo IP: " + m.Input.View() + "\nEnter = guardar   q = cancelar"
}
