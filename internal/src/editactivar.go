package src

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditActivarItem struct {
	Code      string
	Name      string
	IP        string
	ActivarMS bool
}

type EditActivarModel struct {
	Items  []EditActivarItem
	Cursor int
}

var editStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
var editHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

func NewEditActivarModel(items []EditActivarItem) EditActivarModel {
	return EditActivarModel{Items: items, Cursor: 0}
}

func (m EditActivarModel) Init() tea.Cmd {
	return nil
}

func (m EditActivarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "k":
			return m, tea.Quit
		case " ", "enter":
			if len(m.Items) == 0 {
				return m, nil
			}
			m.Items[m.Cursor].ActivarMS = !m.Items[m.Cursor].ActivarMS
			return m, nil
		case "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
			return m, nil
		}
	}
	return m, nil
}

func (m EditActivarModel) View() string {
	title := "  Seleccione el banco que desea activar el microservicio de msNotification\n\n\n"

	if len(m.Items) == 0 {
		return editStyle.Render(title + "  No hay bancos seleccionados.\n\n" + editHelpStyle.Render("  K = salir"))
	}

	var list string
	for i, it := range m.Items {
		mark := "[ ]"
		if it.ActivarMS {
			mark = "[x]"
		}
		line := "  " + mark + "  " + it.Code + "  " + it.Name + "  activar_ms: " + mark + "\n"
		if i == m.Cursor {
			line = "  > " + mark + "  " + it.Code + "  " + it.Name + "  activar_ms: " + mark + "\n"
		}
		list += line
	}

	return editStyle.Render(title + list + "\n" + editHelpStyle.Render("  Espacio = alternar   K = guardar y salir"))
}

func (m EditActivarModel) GetItems() []EditActivarItem {
	return m.Items
}
