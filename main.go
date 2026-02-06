package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"v2/internal/src"
)

func main() {
	columns := []table.Column{
		{Title: "Codigo", Width: 10},
		{Title: "Nombre de Banco", Width: 30},
		{Title: "ip del servidor", Width: 20},
	}

	rows := []table.Row{
		{"0102", "Banco de Venezuela", "192.168.120.109"},
		{"0191", "BNC", "192.168.120.109"},
		{"0151", "BFC", "192.168.120.109"},
		{"0172", "Bancamiga", "192.168.120.109"},
		{"0105", "Banco Mercantil", "192.168.120.109"},
		{"0108", "Provincial", "192.168.120.109"},
		{"0134", "Banesco", "192.168.120.109"},
		{"0114", "Bancaribe", "192.168.120.109"},
		{"0169", "R4", "192.168.120.109"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("20")).
		Bold(false)
	t.SetStyles(s)

	m := src.NewTableModel(t)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	loading := src.NewLoadingModel("Cargando Configuracion...")
	if _, err := tea.NewProgram(loading).Run(); err != nil {
		fmt.Println("Se presento un error al cargar la configuracion:", err)
		os.Exit(1)
	}
}
