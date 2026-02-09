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

const (
	OpcionSi = 0
	OpcionNo = 1
)

type EditActivarModel struct {
	Items     []EditActivarItem
	Cursor    int
	Cancelled bool
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
			m.Cancelled = true
			return m, tea.Quit
		case "k", "enter":
			m.Cancelled = false
			return m, tea.Quit
		case "left", "h":
			m.Cursor = OpcionSi
			return m, nil
		case "right", "l":
			m.Cursor = OpcionNo
			return m, nil
		}
	}
	return m, nil
}

func (m EditActivarModel) View() string {
	title := "  ¿Activar microservicio (activar_ms) para los bancos seleccionados?\n\n"
	title += "  Bancos seleccionados:\n"

	if len(m.Items) == 0 {
		return editStyle.Render(title + "  (ninguno)\n\n" + editHelpStyle.Render("  K = salir"))
	}

	for _, it := range m.Items {
		title += "    " + it.Code + "  " + it.Name + "\n"
	}
	title += "  Desea activar o desactivar el microservicio para estos bancos\n"
	si := "  Activar "
	no := "  Desactivar "
	if m.Cursor == OpcionSi {
		si = "  > Activar "
	} else {
		no = "  > Desactivar "
	}

	title += "\n  " + si + "    " + no + "\n\n"
	title += editHelpStyle.Render("  ←/→ o h/l = elegir   K o Enter = confirmar   q = cancelar")

	return editStyle.Render(title)
}

func (m EditActivarModel) GetActivarTodos() (activar bool, cancelled bool) {
	return m.Cursor == OpcionSi, m.Cancelled
}
