package src

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type MenuModel struct {
	Opciones []string
	Cursor   int
	Salir    bool
}

func NewMenuModel(opciones []string) MenuModel {
	return MenuModel{Opciones: opciones, Cursor: 0}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) SelectedIndex() int {
	if m.Salir {
		return -1
	}
	return m.Cursor
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Salir = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Opciones)-1 {
				m.Cursor++
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var b strings.Builder
	b.WriteString("¿Qué desea modificar?\n\n")
	for i, op := range m.Opciones {
		cursor := "  "
		if i == m.Cursor {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, op))
	}
	b.WriteString("\nEnter = elegir   q = salir")
	return b.String()
}
