package src

import (
	"v2/internal/handler"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var tableStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
var headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))

type TableModel struct {
	Table        table.Model
	Selectmap    map[int]struct{}
	BaseRows     []table.Row
	SingleSelect bool
	Header       string
	// Confirmed: true si el usuario pulsó K (confirmar), false si pulsó Q (volver)
	Confirmed bool
}

func NewTableModel(t table.Model) TableModel {
	return newTableModel(t, false)
}

func NewTableModelSingleSelect(t table.Model) TableModel {
	return newTableModel(t, true)
}

func NewTableModelWithHeader(t table.Model, header string) TableModel {
	m := newTableModel(t, false)
	m.Header = header
	return m
}

func NewTableModelSingleSelectWithHeader(t table.Model, header string) TableModel {
	m := newTableModel(t, true)
	m.Header = header
	return m
}

func newTableModel(t table.Model, singleSelect bool) TableModel {
	selectmap := make(map[int]struct{})
	base := t.Rows()
	baseRows := make([]table.Row, len(base))
	for i, row := range base {
		baseRows[i] = make(table.Row, len(row))
		copy(baseRows[i], row)
	}
	return TableModel{Table: t, Selectmap: selectmap, BaseRows: baseRows, SingleSelect: singleSelect, Confirmed: false}
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
	handler.ResetInactivity()
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			// Única forma de retroceder: Q (ninguna otra tecla hace nada para volver)
			m.Confirmed = false
			return m, tea.Quit
		case "k":
			// Única forma de avanzar/confirmar: K (ninguna otra tecla hace nada para confirmar)
			m.Confirmed = true
			return m, tea.Quit

		case "esc":
			if m.Table.Focused() {
				m.Table.Blur()
			} else {
				m.Table.Focus()
			}
		case "a":
			if !m.SingleSelect {
				if len(m.Selectmap) == len(m.BaseRows) {
					m.Selectmap = make(map[int]struct{})
				} else {
					for i := range m.BaseRows {
						m.Selectmap[i] = struct{}{}
					}
				}
				(&m.Table).SetRows(m.ApplySelectionMarkers())
				return m, nil
			}
		case " ", "enter":
			idx := m.Table.Cursor()
			if m.SingleSelect {
				m.Selectmap = map[int]struct{}{idx: {}}
				(&m.Table).SetRows(m.ApplySelectionMarkers())
				return m, nil
			}
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
	help := "Espacio/Enter = marcar   A = todos    K = confirmar    Q = volver"
	if m.SingleSelect {
		help = "Enter/Espacio = marcar banco    K = confirmar    Q = volver"
	}
	if m.Header != "" {
		return headerStyle.Render(m.Header) + "\n" + tableStyle.Render(m.Table.View()) + "\n" + help
	}
	return tableStyle.Render(m.Table.View()) + "\n" + help
}

func NewTableViewModel(t table.Model) TableViewModel {
	return TableViewModel{Table: t}
}

type TableViewModel struct {
	Table  table.Model
	Header string
}

func NewTableViewModelWithHeader(t table.Model, header string) TableViewModel {
	return TableViewModel{Table: t, Header: header}
}

func (m TableViewModel) Init() tea.Cmd { return nil }

func (m TableViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	handler.ResetInactivity()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "q", "k", " ":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m TableViewModel) View() string {
	if m.Header != "" {
		return headerStyle.Render(m.Header) + "\n" + tableStyle.Render(m.Table.View()) + "\nEnter = continuar"
	}
	return tableStyle.Render(m.Table.View()) + "\nEnter = continuar"
}
