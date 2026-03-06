package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"v2/internal/api"
	"v2/internal/handler"
	"v2/internal/server"
	"v2/internal/src"

	"github.com/SOLUCIONESSYCOM/scribe"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {

	fmt.Println("Contraseña correcta")

	cfg, _ := handler.LoadConfig(handler.EnvDesarrollo)
	srv := server.NewServer(":8080", cfg)
	go func() {
		if err := srv.Run(); err != nil {
			fmt.Println("Servidor API:", err)
		}
	}()
	fmt.Println("API escuchando en http://localhost:8080 (GET /api/v1/health)")

	if err := handler.Logs(); err != nil {
		fmt.Println("Logs:", err)
		os.Exit(1)
	}
	scribe.Info().Msg("Loggers inicializados, iniciando aplicación")
	_, err := handler.GetPassword()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	opcionesMenu := []string{

		"Activar/Desactivar MS en bancos",
		"Cambiar configuracion del endpoint",
		"Cambiar configuracion de la ip",
		"Cambiar configuracion del tiempo de espera entre bucles",
		"Cambiar configuracion del numero de solicitudes por bucle",
		"Cambiar configuracion del numero maximo de solicitudes por operacion",
	}

	for {
		menuEntorno := src.NewMenuModelWithTitle("¿En qué ambiente quiere realizar modificaciones?", []string{"Producción", "Desarrollo"})
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

		configPath, err := handler.GetConfigPath(env)
		if err != nil {
			fmt.Println("Error al obtener ruta de configuración:", err)
			continue
		}
		if !handler.ConfigExists(configPath) {
			fmt.Printf("No se encontró el archivo de configuración en la ruta %s. Verifique que se haya creado correctamente en esa ruta.\n", configPath)
		}

		for menuLoop := true; menuLoop; {
			menuModel := src.NewMenuModelWithTitle("¿Qué desea hacer?", opcionesMenu)
			menuFinal, err := tea.NewProgram(menuModel).Run()
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			menu := menuFinal.(src.MenuModel)
			idx := menu.SelectedIndex()
			if idx < 0 {
				menuLoop = false
				continue
			}

			cfg, err := handler.LoadConfig(env)
			if err != nil {
				fmt.Println("Error al cargar configuración:", err)
				fmt.Print("Presione Enter para continuar...")
				var discard string
				fmt.Scanln(&discard)
				os.Exit(1)
			}

			// Casos 1 y 2: un solo banco; caso 0 y 3: varios
			singleSelect := idx == 1 || idx == 2
			selectedRows, err := runBankTable(cfg, singleSelect)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			if len(selectedRows) == 0 {
				continue
			}

			var accionNombre string
			var accionDetalles map[string]interface{}
			switch idx {

			case 0:
				primeraVez := true
				var rowsActivar []table.Row
				var activarTodos bool
				for {
					if primeraVez {
						rowsActivar = selectedRows
						primeraVez = false
					} else {
						rowsActivar, err = runBankTable(cfg, false)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						if len(rowsActivar) == 0 {
							break
						}
					}
					var editItems []handler.EditActivarItem
					for _, row := range rowsActivar {
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
					var cancelled bool
					activarTodos, cancelled = editModel.GetActivarTodos()
					if cancelled {
						// q: volver a selección de bancos, sin contraseña ni actualizar
						continue
					}

					_, err = handler.GetPassword()
					fmt.Println("Contraseña correcta")

					for _, it := range editItems {
						entry, exists := cfg[it.Code]
						if !exists {
							entry = handler.Bancos{
								Nombre:   it.Name,
								Endpoint: "http://" + it.IP + ":8080/api",
								IP:       it.IP,
							}
						}
						entry.Envio.Activar = activarTodos
						cfg[it.Code] = entry
					}
					accionNombre = "activar_desactivar_ms"
					bancosActivar := make([]string, 0, len(rowsActivar))
					for _, r := range rowsActivar {
						if len(r) >= 1 {
							bancosActivar = append(bancosActivar, r[0])
						}
					}
					accionDetalles = map[string]interface{}{"bancos": bancosActivar, "activar": activarTodos}
					break
				}
				if len(rowsActivar) == 0 {
					continue
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
					continue
				}

				_, err = handler.GetPassword()
				fmt.Println("Contraseña correcta")
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
				accionNombre = "cambiar_endpoint"
				bancosEp := make([]string, 0, len(selectedRows))
				for _, row := range selectedRows {
					if len(row) >= 1 {
						bancosEp = append(bancosEp, row[0])
					}
				}
				accionDetalles = map[string]interface{}{"bancos": bancosEp, "endpoint": newEndpoint}

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
					continue
				}

				_, err = handler.GetPassword()
				fmt.Println("Contraseña correcta")

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
				accionNombre = "cambiar_ip"
				bancosIP := make([]string, 0, len(selectedRows))
				for _, row := range selectedRows {
					if len(row) >= 1 {
						bancosIP = append(bancosIP, row[0])
					}
				}
				accionDetalles = map[string]interface{}{"bancos": bancosIP, "ip": newIP}
			// Cambiar tiempo de espera entre bucles
			case 3:
				primeraVezTiempo := true
				var rowsTiempoBucle []table.Row
				var valorTiempoBucle int
				for {
					if primeraVezTiempo {
						rowsTiempoBucle = selectedRows
						primeraVezTiempo = false
					} else {
						rowsTiempoBucle, err = runBankTable(cfg, false)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						if len(rowsTiempoBucle) == 0 {
							break
						}
					}
					valorConfirmado := false
					for {
						TiempoBucle := handler.NewEditTiempoBucle("10")
						TiempoBucleFinal, err := tea.NewProgram(TiempoBucle).Run()
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						TiempoBucleModel := TiempoBucleFinal.(handler.EditTiempoBucle)
						newTiempoBucle, cancelled := TiempoBucleModel.GetTiempoBucle()
						if cancelled || newTiempoBucle == "" {
							break
						}
						valorTiempoBucle, err = strconv.Atoi(newTiempoBucle)
						if err != nil {
							fmt.Println("Solo se permiten números. Intente de nuevo o cancele con q.")
							fmt.Print("Presione Enter para continuar...")
							var discard string
							fmt.Scanln(&discard)
							continue
						}
						if valorTiempoBucle < 1 || valorTiempoBucle > 10 {
							fmt.Println("El valor debe estar entre 1 y 10. Intente de nuevo o cancele con q.")
							fmt.Print("Presione Enter para continuar...")
							var discard string
							fmt.Scanln(&discard)
							continue
						}
						valorConfirmado = true
						break
					}
					if !valorConfirmado {
						continue
					}
					break
				}
				if len(rowsTiempoBucle) == 0 {
					continue
				}

				_, err = handler.GetPassword()
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}

				fmt.Println("Contraseña correcta")

				for _, row := range rowsTiempoBucle {
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
					entry.Envio.TiempoDeEsperaEntreBucles = valorTiempoBucle
					cfg[code] = entry
				}
				accionNombre = "cambiar_tiempo_espera_entre_bucles"
				bancosTiempo := make([]string, 0, len(rowsTiempoBucle))
				for _, row := range rowsTiempoBucle {
					if len(row) >= 1 {
						bancosTiempo = append(bancosTiempo, row[0])
					}
				}
				accionDetalles = map[string]interface{}{"bancos": bancosTiempo, "TiempoDeEsperaEntreBucles": valorTiempoBucle}

			// Cambiar numero de solicitudes por bucle
			case 4:
				primeraVezSolBucle := true
				var rowsNumeroSolBucle []table.Row
				var valorNumeroSolBucle int
				for {
					if primeraVezSolBucle {
						rowsNumeroSolBucle = selectedRows
						primeraVezSolBucle = false
					} else {
						rowsNumeroSolBucle, err = runBankTable(cfg, false)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						if len(rowsNumeroSolBucle) == 0 {
							break
						}
					}
					valorConfirmadoSolBucle := false
					for {
						NumeroSolBucle := handler.NewEditNumeroSolBucle("10")
						NumeroSolBucleFinal, err := tea.NewProgram(NumeroSolBucle).Run()
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						NumeroSolBucleModel := NumeroSolBucleFinal.(handler.EditNumeroSolBucle)
						newNumeroSolBucle, cancelled := NumeroSolBucleModel.GetNumeroSolBucle()
						if cancelled || newNumeroSolBucle == "" {
							break
						}
						valorNumeroSolBucle, err = strconv.Atoi(newNumeroSolBucle)
						if err != nil {
							fmt.Println("Solo se permiten números. Intente de nuevo o cancele con q.")
							fmt.Print("Presione Enter para continuar...")
							var discard string
							fmt.Scanln(&discard)
							continue
						}
						if valorNumeroSolBucle < 1 || valorNumeroSolBucle > 10 {
							fmt.Println("El valor debe estar entre 1 y 10. Intente de nuevo o cancele con q.")
							fmt.Print("Presione Enter para continuar...")
							var discard string
							fmt.Scanln(&discard)
							continue
						}
						valorConfirmadoSolBucle = true
						break
					}
					if !valorConfirmadoSolBucle {
						continue
					}
					break
				}
				if len(rowsNumeroSolBucle) == 0 {
					continue
				}

				_, err = handler.GetPassword()
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
				fmt.Println("Contraseña correcta")

				for _, row := range rowsNumeroSolBucle {
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
					entry.Envio.NumeroDeSolicitudesPorBucle = valorNumeroSolBucle
					cfg[code] = entry
				}
				accionNombre = "cambiar_tiempo_espera_entre_bucles"
				bancosTiempo := make([]string, 0, len(rowsNumeroSolBucle))
				for _, row := range rowsNumeroSolBucle {
					if len(row) >= 1 {
						bancosTiempo = append(bancosTiempo, row[0])
					}
				}
				accionDetalles = map[string]interface{}{"bancos": bancosTiempo, "NumeroDeSolicitudesPorBucle": valorNumeroSolBucle}

			// Cambiar numero maximo de solicitudes por operacion
			case 5:
				primeraVezMaxOp := true
				var rowsNumeroMaxOp []table.Row
				var valorNumeroMaxPorOperacion int
				for {
					if primeraVezMaxOp {
						rowsNumeroMaxOp = selectedRows
						primeraVezMaxOp = false
					} else {
						rowsNumeroMaxOp, err = runBankTable(cfg, false)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						if len(rowsNumeroMaxOp) == 0 {
							break
						}
					}
					valorConfirmadoMaxOp := false
					for {
						NumeroMaxPorOperacion := handler.NewEditNumMaxPorBucle("5")
						NumeroMaxPorOperacionFinal, err := tea.NewProgram(NumeroMaxPorOperacion).Run()
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						NumeroMaxPorOperacionModel := NumeroMaxPorOperacionFinal.(handler.EditNumMaxPorBucle)
						newNumeroMaxPorOperacion, cancelled := NumeroMaxPorOperacionModel.GetNumMaxPorBucle()
						if cancelled || newNumeroMaxPorOperacion == "" {
							break
						}
						valorNumeroMaxPorOperacion, err = strconv.Atoi(newNumeroMaxPorOperacion)
						if err != nil {
							fmt.Println("Solo se permiten números. Intente de nuevo o cancele con q.")
							fmt.Print("Presione Enter para continuar...")
							var discard string
							fmt.Scanln(&discard)
							continue
						}
						if valorNumeroMaxPorOperacion < 1 || valorNumeroMaxPorOperacion > 5 {
							fmt.Println("El valor debe estar entre 1 y 5. Intente de nuevo o cancele con q.")
							fmt.Print("Presione Enter para continuar...")
							var discard string
							fmt.Scanln(&discard)
							continue
						}
						valorConfirmadoMaxOp = true
						break
					}
					if !valorConfirmadoMaxOp {
						continue
					}
					break
				}
				if len(rowsNumeroMaxOp) == 0 {
					continue
				}

				_, err = handler.GetPassword()
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
				fmt.Println("Contraseña correcta")

				for _, row := range rowsNumeroMaxOp {
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
					entry.Envio.NumeroMaximoDeSolicitudesPorOperacion = valorNumeroMaxPorOperacion
					cfg[code] = entry
				}
				accionNombre = "cambiar_tiempo_espera_entre_bucles"
				bancosTiempo := make([]string, 0, len(rowsNumeroMaxOp))
				for _, row := range rowsNumeroMaxOp {
					if len(row) >= 1 {
						bancosTiempo = append(bancosTiempo, row[0])
					}
				}
				accionDetalles = map[string]interface{}{"bancos": bancosTiempo, "NumeroMaximoDeSolicitudesPorOperacion": valorNumeroMaxPorOperacion}

			case 6:
				// Volver al menú anterior = menú de ambiente (Producción/Desarrollo)
				menuLoop = false
				continue

			default:
				return
			}

			// Guardar configuración primero
			if err := handler.SaveConfig(env, cfg); err != nil {
				fmt.Println("Error al guardar configuración:", err)
				os.Exit(1)
			}

			// Enviar actualización solest a cada banco modificado (antes del health check)
			SolestStatus := make(map[string]int)
			if bancos, ok := accionDetalles["bancos"].([]string); ok {
				for _, codigoBanco := range bancos {
					url := "http://localhost:8080/simf/api/v1/config/solest?banco=" + codigoBanco + "&env=" + env
					resp, err := http.Post(url, "application/json", nil)
					if err != nil {
						fmt.Println("Error enviando config al banco", codigoBanco, ":", err)
						if b, ok := cfg[codigoBanco]; ok {
							SolestStatus[b.Nombre] = 0
						}
						continue
					}
					status := resp.StatusCode
					resp.Body.Close()
					if b, ok := cfg[codigoBanco]; ok {
						SolestStatus[b.Nombre] = status
					}
					if resp.StatusCode >= 400 {
						fmt.Println("Banco", codigoBanco, "respondió con status:", resp.StatusCode)
					}
				}
			}

			// Health check al final de la opción elegida (después de guardar y enviar a bancos)
			health := api.RunHealthCheck(cfg)
			tablaHealthCheck(health, SolestStatus)
			fmt.Print("Presione Enter para continuar...")
			var enter string
			fmt.Scanln(&enter)

			configPath, _ := handler.GetConfigPath(env)
			fmt.Println("\n\nConfiguración guardada en:", configPath)

			handler.LogAccion(accionNombre, env, accionDetalles)

			loading := src.NewLoadingModel("Cargando Configuracion...")
			if _, err := tea.NewProgram(loading).Run(); err != nil {
				fmt.Println("Se presento un error al cargar la configuracion:", err)
				os.Exit(1)
			}

			fmt.Println("Listo.")
			continue
		}
	}
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

// tabla health check

func tablaHealthCheck(health api.HealthResponse, SolestStatus map[string]int) {
	columns := []table.Column{
		{Title: "Banco", Width: 40},
		{Title: "Estado", Width: 30},
		{Title: "Status http", Width: 30},
	}
	names := make([]string, 0, len(health.Checks))
	for n := range health.Checks {
		names = append(names, n)
	}
	sort.Strings(names)
	rows := make([]table.Row, 0, len(names)+1)
	rows = append(rows, table.Row{"Status general", health.Status, "-"})
	for _, n := range names {
		statusStr := "-"
		if code, ok := SolestStatus[n]; ok {
			statusStr = strconv.Itoa(code)
		}
		rows = append(rows, table.Row{n, health.Checks[n], statusStr})
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(12, len(rows)+1)),
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
	m := src.NewTableViewModel(t)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error al mostrar health check:", err)
	}
}
