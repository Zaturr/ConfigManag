package src

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var tableStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))

type TableModel struct {
	Table     table.Model
	Selectmap map[int]struct{}
	BaseRows  []table.Row
}

func NewTableModel(t table.Model) TableModel {
	selectmap := make(map[int]struct{})
	base := t.Rows()
	baseRows := make([]table.Row, len(base))
	for i, row := range base {
		baseRows[i] = make(table.Row, len(row))
		copy(baseRows[i], row)
	}
	return TableModel{Table: t, Selectmap: selectmap, BaseRows: baseRows}
}

func (m TableModel) ApplySelectionMarkers() []table.Row {
	out := make([]table.Row, len(m.BaseRows))
	for i, row := range m.BaseRows {
		out[i] = make(table.Row, len(row))
		copy(out[i], row)
		if _, ok := m.Selectmap[i]; ok {
			out[i][0] = "[x]"
		} else {
			out[i][0] = "[ ]"
		}
	}
	return out
}
func (m TableModel) SelectedRows() []table.Row {
	var out []table.Row
	for idx := range m.Selectmap {
		if idx >= 0 && idx < len(m.BaseRows) {
			row := m.BaseRows[idx]
			if len(row) > 1 {
				out = append(out, row[1:])
			} else {
				out = append(out, row)
			}
		}
	}
	return out
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Table.Focused() {
				m.Table.Blur()
			} else {
				m.Table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "k":
			return m, tea.Quit

		case " ", "enter":
			idx := m.Table.Cursor()
			if _, ok := m.Selectmap[idx]; ok {
				delete(m.Selectmap, idx)
			} else {
				m.Selectmap[idx] = struct{}{}
			}
			(&m.Table).SetRows(m.ApplySelectionMarkers())
			return m, nil
		}

	}
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m TableModel) View() string {
	return tableStyle.Render(m.Table.View()) + "\npresione Espacio/Enter = marcar, K = avanzar"
}
