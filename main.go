package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"v2/internal/handler"
	"v2/internal/src"
)

func main() {
	// Elegir entorno (Producción o Desarrollo)
	menuEntorno := src.NewMenuModel([]string{"Producción", "Desarrollo"})
	envFinal, err := tea.NewProgram(menuEntorno).Run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	envMenu := envFinal.(src.MenuModel)
	envIdx := envMenu.SelectedIndex()
	if envIdx < 0 {
		return
	}
	var env string
	if envIdx == 0 {
		env = handler.EnvProduccion
	} else {
		env = handler.EnvDesarrollo
	}

	opcionesMenu := []string{
		"Activar/Desactivar MS en bancos",
		"Cambiar configuracion del endpoint",
		"Cambiar configuracion de la ip",
	}
	menuModel := src.NewMenuModel(opcionesMenu)
	menuFinal, err := tea.NewProgram(menuModel).Run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	menu := menuFinal.(src.MenuModel)
	idx := menu.SelectedIndex()
	if idx < 0 {
		return
	}

	cfg, err := handler.LoadConfig(env)
	if err != nil {
		fmt.Println("Error al cargar configuración:", err)
		os.Exit(1)
	}

	// Casos 1 y 2: un solo banco caso 0: varios
	singleSelect := idx == 1 || idx == 2
	selectedRows, err := runBankTable(cfg, singleSelect)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if len(selectedRows) == 0 {
		fmt.Println("No se seleccionó ningún banco. La configuración no se modifica.")
		return
	}

	switch idx {

	// Activar/Desactivar MS
	case 0:
		var editItems []handler.EditActivarItem
		for _, row := range selectedRows {
			if len(row) < 3 {
				continue
			}
			code, name, ip := row[0], row[1], row[2]
			editItems = append(editItems, handler.EditActivarItem{
				Code: code, Name: name, IP: ip,
			})
		}
		editModel := handler.NewEditActivarModel(editItems)
		editFinal, err := tea.NewProgram(editModel).Run()
		if err != nil {
			fmt.Println("Error en pantalla de edición:", err)
			os.Exit(1)
		}
		editModel = editFinal.(handler.EditActivarModel)
		activarTodos, cancelled := editModel.GetActivarTodos()
		if cancelled {
			fmt.Println("Cancelado. La configuración no se modifica.")
			return
		}
		for _, it := range editItems {
			entry, exists := cfg[it.Code]
			if !exists {
				entry = handler.Bancos{
					Nombre:   it.Name,
					Endpoint: "http://" + it.IP + ":8080/api",
					IP:       it.IP,
				}
			}
			entry.ActivarMS = activarTodos
			cfg[it.Code] = entry
		}

	// Cambiar endpoint
	case 1:

		endpointModel := handler.NewEditEndpointModel("http://192.168.1.1:8080/api")
		epFinal, err := tea.NewProgram(endpointModel).Run()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		epModel := epFinal.(handler.EditEndpointModel)
		newEndpoint, cancelled := epModel.GetEndpoint()
		if cancelled || newEndpoint == "" {
			fmt.Println("Cancelado o endpoint vacío. La configuración no se modifica.")
			return
		}
		for _, row := range selectedRows {
			if len(row) < 1 {
				continue
			}
			code := row[0]
			entry, exists := cfg[code]
			if !exists {
				entry = handler.Bancos{}
				if len(row) >= 3 {
					entry.Nombre, entry.IP = row[1], row[2]
				}
			}
			entry.Endpoint = newEndpoint
			cfg[code] = entry
		}

	// Cambiar IP
	case 2:

		IPModel := handler.NewEditIPModel("192.168.1.1")
		IPFinal, err := tea.NewProgram(IPModel).Run()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		ipModel := IPFinal.(handler.EditIPModel)
		newIP, cancelled := ipModel.GetIP()
		if cancelled || newIP == "" {
			fmt.Println("Cancelado o IP vacío. La configuración no se modifica.")
			return
		}
		for _, row := range selectedRows {
			if len(row) < 1 {
				continue
			}
			code := row[0]
			entry, exists := cfg[code]
			if !exists {
				entry = handler.Bancos{}
				if len(row) >= 3 {
					entry.Nombre, entry.IP = row[1], row[2]
				}
			}
			entry.IP = newIP
			cfg[code] = entry
		}
	default:
		return
	}

	if err := handler.SaveConfig(env, cfg); err != nil {
		fmt.Println("Error al guardar configuración:", err)
		os.Exit(1)
	}

	configPath, _ := handler.GetConfigPath(env)
	fmt.Println("\n\nConfiguración guardada en:", configPath)

	loading := src.NewLoadingModel("Cargando Configuracion...")
	if _, err := tea.NewProgram(loading).Run(); err != nil {
		fmt.Println("Se presento un error al cargar la configuracion:", err)
		os.Exit(1)
	}

	fmt.Println("Listo.")
}

func runBankTable(cfg handler.Config, singleSelect bool) ([]table.Row, error) {
	if len(cfg) == 0 {
		return nil, fmt.Errorf("no hay bancos en la configuración del ambiente seleccionado")
	}
	columns := []table.Column{
		{Title: " ", Width: 4},
		{Title: "Codigo", Width: 10},
		{Title: "Nombre de Banco", Width: 30},
		{Title: "ip del servidor", Width: 20},
	}
	codes := make([]string, 0, len(cfg))
	for code := range cfg {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	rows := make([]table.Row, 0, len(codes))
	for _, code := range codes {
		b := cfg[code]
		rows = append(rows, table.Row{"[ ]", code, b.Nombre, b.IP})
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
	var m src.TableModel
	if singleSelect {
		m = src.NewTableModelSingleSelect(t)
	} else {
		m = src.NewTableModel(t)
	}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}
	tbl, ok := finalModel.(src.TableModel)
	if !ok {
		return nil, fmt.Errorf("no se pudo obtener el modelo de la tabla")
	}
	return tbl.SelectedRows(), nil
}
