package src

import (
	"fmt"
	"strings"

	"v2/internal/handler"

	tea "github.com/charmbracelet/bubbletea"
)

type MenuModel struct {
	Titulo   string
	Opciones []string
	Cursor   int
	Salir    bool
}

func NewMenuModel(opciones []string) MenuModel {
	return MenuModel{Opciones: opciones, Cursor: 0}
}

// NewMenuModelWithTitle crea un menú con título personalizado.
func NewMenuModelWithTitle(titulo string, opciones []string) MenuModel {
	return MenuModel{Titulo: titulo, Opciones: opciones, Cursor: 0}
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
	handler.ResetInactivity()
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
	titulo := m.Titulo
	if titulo == "" {
		titulo = "¿En que ambiente quiere realizar modificaciones?"
	}
	b.WriteString(titulo + "\n\n")
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
