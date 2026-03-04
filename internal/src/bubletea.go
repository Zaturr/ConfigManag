package src

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func NewModel(choices []string) Model {
	return Model{
		choices:  choices,
		cursor:   0,
		selected: make(map[int]struct{}),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("K/j = subir/bajar   Enter o Espacio = marcar   q = volver\n\n")
	for i, c := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		mark := " "
		if _, ok := m.selected[i]; ok {
			mark = "x"
		}
		b.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, mark, c))
	}
	return b.String()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
			return m, nil
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			return m, nil
		}
		return m, nil
	}
	return m, nil
}
