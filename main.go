package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"v2/internal/api"
	"v2/internal/handler"
	"v2/internal/server"
	"v2/internal/src"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gin-gonic/gin"
)

// Títulos de las tablas interactivas (bubbletea); el texto impreso con fmt.Println encima puede mantenerse o alinearse con estos.
const (
	tituloTablaCfgAntes   = "Configuración de los bancos (antes de los cambios)"
	tituloTablaCfgDespues = "Configuración de los bancos (después de los cambios)"
	tituloTablaCfg        = "Configuración de los bancos"

	tituloHealthAntesActivarMS = "Conexión con los bancos (antes de Activar/Desactivar)"
	tituloHealthAntesEnviar    = "Conexión con los bancos (antes de enviar)"
	tituloHealthDespuesEnviar  = "Conexión con los bancos (después de enviar)"
	tituloHealth               = "Conexión con los bancos"
)

func main() {

	fmt.Print("\033[8;30;143t")

	fmt.Println("Contraseña correcta")

	// Suprimir logs de Gin en consola; los de scribe se mantienen
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	cfg, _ := handler.LoadConfig(handler.EnvDesarrollo)
	srv := server.NewServer(":8080", cfg)
	go func() {
		if err := srv.Run(); err != nil {
			//fmt.Println("Servidor API:", err)
		}
	}()

	usuario, dispositivo := handler.GetUsuarioDispositivo()
	defer func() {
		handler.LogSesionCierre(usuario, dispositivo, time.Now().Format(time.RFC3339))
	}()

	handler.StartInactivityTimer(handler.DefaultInactivityMinutes)
	//fmt.Println("Configuración actual (Desarrollo):")

	_, err := handler.GetPassword()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	opcionesMenu := []string{

		"Activar/Desactivar servicio de solicitud de estado en bancos",
		"Cambiar configuración del endpoint",
		"Cambiar configuración de la IP",
		"Cambiar configuración del tiempo de espera entre bucles",
		"Cambiar configuración del número de solicitudes por bucle",
		"Cambiar configuración del número máximo de solicitudes por operación",
		"Mostrar configuración actual",
		"Mostrar health check actual",
		"Finalizar sesión y cerrar aplicación",
	}

	var sessionStartLogged bool
	for {
		menuEntorno := src.NewMenuModelWithTitle("¿En qué ambiente quiere realizar modificaciones?", []string{"Producción", "Desarrollo"})
		envFinal, err := tea.NewProgram(menuEntorno).Run()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		handler.ResetInactivity()

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

		if err := handler.Logs(env); err != nil {
			fmt.Println("Logs:", err)
			continue
		}
		handler.SetSesionUsuario(usuario, dispositivo)
		if !sessionStartLogged {
			handler.LogSesionInicio(usuario, dispositivo, time.Now().Format(time.RFC3339))
			sessionStartLogged = true
		}

		configPath, err := handler.GetConfigPath(env)
		if err != nil {
			fmt.Println("Error al obtener ruta de configuración:", err)
			continue
		}
		if !handler.ConfigExists(configPath) {
			fmt.Printf("No se encontró el archivo de configuración en la ruta %s. Verifique que se haya creado correctamente en esa ruta.\n", configPath)
		}

		cfgAmbiente, err := handler.LoadConfig(env)
		if err != nil {
			fmt.Println("Error al cargar configuración:", err)
			continue
		}
		fmt.Println("Configuración actual:")
		mostrarTablaConfig(cfgAmbiente, tituloTablaCfg)

		for menuLoop := true; menuLoop; {
			menuModel := src.NewMenuModelWithTitle("¿Qué desea hacer?", opcionesMenu)
			menuFinal, err := tea.NewProgram(menuModel).Run()
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			handler.ResetInactivity()
			menu := menuFinal.(src.MenuModel)
			idx := menu.SelectedIndex()
			if idx < 0 {
				menuLoop = false
				continue
			}
			if idx == 8 {
				return
			}

			cfg, err := handler.LoadConfig(env)
			if err != nil {
				fmt.Println("Error al cargar configuración:", err)
				//fmt.Print("Presione Enter para continuar...")
				//var discard string
				//fmt.Scanln(&discard)
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
				// Health check ANTES de Activar/Desactivar
				fmt.Println("Verificando conexion, espere por favor:")
				healthAntes := api.RunHealthCheck(cfg)
				//fmt.Println("Health check antes de Activar/Desactivar servicio de solicitud de estado:")
				tablaHealthCheck(healthAntes, tituloHealthAntesActivarMS, healthFilterNamesFromCodes(cfg, selectedCodesFromRows(selectedRows))...)

				primeraVez := true
				var rowsActivar []table.Row
				var activarTodos bool
				// Opcionales solo cuando se elige Activar (se piden antes de la contraseña)
				var wantTiempoBucle bool
				var valorTiempoBucle int
				var wantNumeroSolBucle bool
				var valorNumeroSolBucle int
				var wantNumeroMaxOp bool
				var valorNumeroMaxOp int
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
					handler.ResetInactivity()
					editModel = editFinal.(handler.EditActivarModel)
					var cancelled bool
					activarTodos, cancelled = editModel.GetActivarTodos()
					if cancelled {
						continue
					}

					if activarTodos {
						// Submenú de configuraciones solo al activar (antes de la contraseña)
						wantTiempoBucle = false
						wantNumeroSolBucle = false
						wantNumeroMaxOp = false
						configAbort := false
						// // 1) Tiempo de espera entre bucles
						// menuTiempo := src.NewMenuModelWithTitle("¿Desea cambiar la configuración del tiempo de espera entre bucles?", []string{"Sí", "No"})
						// progTiempo, err := tea.NewProgram(menuTiempo).Run()
						// if err != nil {
						// 	fmt.Println("Error:", err)
						// 	os.Exit(1)
						// }
						// mTiempo := progTiempo.(src.MenuModel)
						// if mTiempo.SelectedIndex() < 0 {
						// 	configAbort = true
						// } else if mTiempo.SelectedIndex() == 0 {
						// 	valorConfirmado := false
						// 	for {
						// 		TiempoBucle := handler.NewEditTiempoBucle("10")
						// 		TiempoBucleFinal, err := tea.NewProgram(TiempoBucle).Run()
						// 		if err != nil {
						// 			fmt.Println("Error:", err)
						// 			os.Exit(1)
						// 		}
						// 		TiempoBucleModel := TiempoBucleFinal.(handler.EditTiempoBucle)
						// 		newTiempoBucle, cancelledT := TiempoBucleModel.GetTiempoBucle()
						// 		if cancelledT || newTiempoBucle == "" {
						// 			break
						// 		}
						// 		valorTiempoBucle, err = strconv.Atoi(newTiempoBucle)
						// 		if err != nil {
						// 			fmt.Println("Solo se permiten números. Intente de nuevo o cancele con q.")
						// 			fmt.Print("Presione Enter para continuar...")
						// 			var discard string
						// 			fmt.Scanln(&discard)
						// 			continue
						// 		}
						// 		if valorTiempoBucle < 1 || valorTiempoBucle > 10 {
						// 			fmt.Println("El valor debe estar entre 1 y 10. Intente de nuevo o cancele con q.")
						// 			fmt.Print("Presione Enter para continuar...")
						// 			var discard string
						// 			fmt.Scanln(&discard)
						// 			continue
						// 		}
						// 		valorConfirmado = true
						// 		wantTiempoBucle = true
						// 		break
						// 	}
						// 	if !valorConfirmado {
						// 		configAbort = true
						// 	}
						// }
						// if configAbort {
						// 	continue
						// }
						//
						// // 2) Número de solicitudes por bucle
						// menuSolBucle := src.NewMenuModelWithTitle("¿Desea cambiar la configuración del número de solicitudes por bucle?", []string{"Sí", "No"})
						// progSolBucle, err := tea.NewProgram(menuSolBucle).Run()
						// if err != nil {
						// 	fmt.Println("Error:", err)
						// 	os.Exit(1)
						// }
						// mSolBucle := progSolBucle.(src.MenuModel)
						// if mSolBucle.SelectedIndex() < 0 {
						// 	configAbort = true
						// } else if mSolBucle.SelectedIndex() == 0 {
						// 	valorConfirmado := false
						// 	for {
						// 		NumeroSolBucle := handler.NewEditNumeroSolBucle("10")
						// 		NumeroSolBucleFinal, err := tea.NewProgram(NumeroSolBucle).Run()
						// 		if err != nil {
						// 			fmt.Println("Error:", err)
						// 			os.Exit(1)
						// 		}
						// 		NumeroSolBucleModel := NumeroSolBucleFinal.(handler.EditNumeroSolBucle)
						// 		newNumeroSolBucle, cancelledN := NumeroSolBucleModel.GetNumeroSolBucle()
						// 		if cancelledN || newNumeroSolBucle == "" {
						// 			break
						// 		}
						// 		valorNumeroSolBucle, err = strconv.Atoi(newNumeroSolBucle)
						// 		if err != nil {
						// 			fmt.Println("Solo se permiten números. Intente de nuevo o cancele con q.")
						// 			fmt.Print("Presione Enter para continuar...")
						// 			var discard string
						// 			fmt.Scanln(&discard)
						// 			continue
						// 		}
						// 		if valorNumeroSolBucle < 1 || valorNumeroSolBucle > 10 {
						// 			fmt.Println("El valor debe estar entre 1 y 10. Intente de nuevo o cancele con q.")
						// 			fmt.Print("Presione Enter para continuar...")
						// 			var discard string
						// 			fmt.Scanln(&discard)
						// 			continue
						// 		}
						// 		valorConfirmado = true
						// 		wantNumeroSolBucle = true
						// 		break
						// 	}
						// 	if !valorConfirmado {
						// 		configAbort = true
						// 	}
						// }
						if configAbort {
							continue
						}
						//
						// // 3) Número max de solicitudes por operación
						// menuMaxOp := src.NewMenuModelWithTitle("¿Desea cambiar la configuración del número máximo de solicitudes por operación?", []string{"Sí", "No"})
						// progMaxOp, err := tea.NewProgram(menuMaxOp).Run()
						// if err != nil {
						// 	fmt.Println("Error:", err)
						// 	os.Exit(1)
						// }
						// mMaxOp := progMaxOp.(src.MenuModel)
						// if mMaxOp.SelectedIndex() < 0 {
						// 	configAbort = true
						// } else if mMaxOp.SelectedIndex() == 0 {
						// 	valorConfirmado := false
						// 	for {
						// 		NumeroMaxPorOperacion := handler.NewEditNumMaxPorBucle("5")
						// 		NumeroMaxPorOperacionFinal, err := tea.NewProgram(NumeroMaxPorOperacion).Run()
						// 		if err != nil {
						// 			fmt.Println("Error:", err)
						// 			os.Exit(1)
						// 		}
						// 		NumeroMaxPorOperacionModel := NumeroMaxPorOperacionFinal.(handler.EditNumMaxPorBucle)
						// 		newNumeroMaxPorOperacion, cancelledM := NumeroMaxPorOperacionModel.GetNumMaxPorBucle()
						// 		if cancelledM || newNumeroMaxPorOperacion == "" {
						// 			break
						// 		}
						// 		valorNumeroMaxOp, err = strconv.Atoi(newNumeroMaxPorOperacion)
						// 		if err != nil {
						// 			fmt.Println("Solo se permiten números. Intente de nuevo o cancele con q.")
						// 			fmt.Print("Presione Enter para continuar...")
						// 			var discard string
						// 			fmt.Scanln(&discard)
						// 			continue
						// 		}
						// 		if valorNumeroMaxOp < 1 || valorNumeroMaxOp > 5 {
						// 			fmt.Println("El valor debe estar entre 1 y 5. Intente de nuevo o cancele con q.")
						// 			fmt.Print("Presione Enter para continuar...")
						// 			var discard string
						// 			fmt.Scanln(&discard)
						// 			continue
						// 		}
						// 		valorConfirmado = true
						// 		wantNumeroMaxOp = true
						// 		break
						// 	}
						// 	if !valorConfirmado {
						// 		configAbort = true
						// 	}
						// }
						// if configAbort {
						// 	continue
						// }

						// Opción final: solo al elegir "Realizar cambios" se pide contraseña y se aplica
						menuAplicar := src.NewMenuModelWithTitle("¿Desea realizar los cambios? (se pedirá contraseña)", []string{"Realizar cambios", "Volver"})
						progAplicar, err := tea.NewProgram(menuAplicar).Run()
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						mAplicar := progAplicar.(src.MenuModel)
						if mAplicar.SelectedIndex() != 0 {
							// Volver o q: no aplicar, volver a selección de bancos
							continue
						}
					} else {
						// Desactivar: también debe elegir "Realizar cambios" para pedir contraseña
						menuAplicar := src.NewMenuModelWithTitle("¿Desea realizar los cambios? (se pedirá contraseña)", []string{"Realizar cambios", "Volver"})
						progAplicar, err := tea.NewProgram(menuAplicar).Run()
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						mAplicar := progAplicar.(src.MenuModel)
						if mAplicar.SelectedIndex() != 0 {
							continue
						}
					}

					//fmt.Println("Configuración actual (antes de aplicar cambios):")
					mostrarTablaConfig(cfg, tituloTablaCfgAntes, selectedCodesFromRows(rowsActivar)...)
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
						if activarTodos && wantTiempoBucle {
							entry.Envio.TiempoDeEsperaEntreBucles = valorTiempoBucle
						}
						if activarTodos && wantNumeroSolBucle {
							entry.Envio.NumeroDeSolicitudesPorBucle = valorNumeroSolBucle
						}
						if activarTodos && wantNumeroMaxOp {
							entry.Envio.NumeroMaximoDeSolicitudesPorOperacion = valorNumeroMaxOp
						}
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
				//	fmt.Println("Configuración después de aplicar cambios:")
				mostrarTablaConfig(cfg, tituloTablaCfgDespues, selectedCodesFromRows(rowsActivar)...)

			// Cambiar endpoint
			case 1:
				endpointModel := handler.NewEditEndpointModel("/api")
				epFinal, err := tea.NewProgram(endpointModel).Run()
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
				handler.ResetInactivity()
				epModel := epFinal.(handler.EditEndpointModel)
				newEndpoint, cancelled := epModel.GetEndpoint()
				if cancelled || newEndpoint == "" {
					fmt.Println("Cancelado o endpoint vacío. La configuración no se modifica.")
					continue
				}

				//fmt.Println("Configuración actual (antes de aplicar cambios):")
				mostrarTablaConfig(cfg, tituloTablaCfgAntes, selectedCodesFromRows(selectedRows)...)
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
				//fmt.Println("Configuración después de aplicar cambios:")
				mostrarTablaConfig(cfg, tituloTablaCfgDespues, selectedCodesFromRows(selectedRows)...)

			// Cambiar IP
			case 2:
				IPModel := handler.NewEditIPModel("192.168.1.1")
				IPFinal, err := tea.NewProgram(IPModel).Run()
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
				handler.ResetInactivity()
				ipModel := IPFinal.(handler.EditIPModel)
				newIP, cancelled := ipModel.GetIP()
				if cancelled || newIP == "" {
					fmt.Println("Cancelado o IP vacío. La configuración no se modifica.")
					continue
				}

				//fmt.Println("Configuración actual (antes de aplicar cambios):")
				mostrarTablaConfig(cfg, tituloTablaCfgAntes, selectedCodesFromRows(selectedRows)...)
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
				//fmt.Println("Configuración después de aplicar cambios:")
				mostrarTablaConfig(cfg, tituloTablaCfgDespues, selectedCodesFromRows(selectedRows)...)
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
						TiempoBucle := handler.NewEditTiempoBucle("20")
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
							//fmt.Print("Presione Enter para continuar...")
							//var discard string
							//fmt.Scanln(&discard)
							continue
						}
						if valorTiempoBucle < 1 || valorTiempoBucle > 20 {
							fmt.Println("El valor debe estar entre 1 y 20. Intente de nuevo o cancele con q.")
							//fmt.Print("Presione Enter para continuar...")
							//var discard string
							//fmt.Scanln(&discard)
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

				//fmt.Println("Configuración actual (antes de aplicar cambios):")
				mostrarTablaConfig(cfg, tituloTablaCfgAntes, selectedCodesFromRows(rowsTiempoBucle)...)
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
				//fmt.Println("Configuración después de aplicar cambios:")
				mostrarTablaConfig(cfg, tituloTablaCfgDespues, selectedCodesFromRows(rowsTiempoBucle)...)

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
						NumeroSolBucle := handler.NewEditNumeroSolBucle("20")
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
							//fmt.Print("Presione Enter para continuar...")
							//var discard string
							//fmt.Scanln(&discard)
							continue
						}
						if valorNumeroSolBucle < 1 || valorNumeroSolBucle > 20 {
							fmt.Println("El valor debe estar entre 1 y 20. Intente de nuevo o cancele con q.")
							//fmt.Print("Presione Enter para continuar...")
							//var discard string
							//fmt.Scanln(&discard)
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

				//fmt.Println("Configuración actual (antes de aplicar cambios):")
				mostrarTablaConfig(cfg, tituloTablaCfgAntes, selectedCodesFromRows(rowsNumeroSolBucle)...)
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
				//fmt.Println("Configuración después de aplicar cambios:")
				mostrarTablaConfig(cfg, tituloTablaCfgDespues, selectedCodesFromRows(rowsNumeroSolBucle)...)

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
						NumeroMaxPorOperacion := handler.NewEditNumMaxPorBucle("15")
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
							//fmt.Print("Presione Enter para continuar...")
							//var discard string
							//fmt.Scanln(&discard)
							continue
						}
						if valorNumeroMaxPorOperacion < 1 || valorNumeroMaxPorOperacion > 15 {
							fmt.Println("El valor debe estar entre 1 y 15. Intente de nuevo o cancele con q.")
							//fmt.Print("Presione Enter para continuar...")
							//var discard string
							//fmt.Scanln(&discard)
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

				//fmt.Println("Configuración actual (antes de aplicar cambios):")
				mostrarTablaConfig(cfg, tituloTablaCfgAntes, selectedCodesFromRows(rowsNumeroMaxOp)...)
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
				//fmt.Println("Configuración después de aplicar cambios:")
				mostrarTablaConfig(cfg, tituloTablaCfgDespues, selectedCodesFromRows(rowsNumeroMaxOp)...)
			case 6:
				//Mostrar Configuración actual
				fmt.Println("Configuración actual:")
				mostrarTablaConfig(cfg, tituloTablaCfg, selectedCodesFromRows(selectedRows)...)

				continue
			case 7:
				//Mostrar Health check actual
				fmt.Println("Verificando conexion, espere por favor:")
				health := api.RunHealthCheck(cfg)
				tablaHealthCheck(health, tituloHealth, healthFilterNamesFromCodes(cfg, selectedCodesFromRows(selectedRows))...)

				continue
			case 8:
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

			// Health check y tabla ANTES de enviar a los bancos, para poder revisar el estado
			if bancosPre, ok := accionDetalles["bancos"].([]string); ok && len(bancosPre) > 0 {
				//if accionNombre == "activar_desactivar_ms" {
				//	fmt.Println("Health check antes de enviar (Activar/Desactivar servicio de solicitud de estado):")
				//}
				// else {
				//	fmt.Println("Health check antes de enviar la configuración a los bancos:")
				//}
				fmt.Println("Verificando conexion, espere por favor:")
				healthPre := api.RunHealthCheck(cfg)
				tablaHealthCheck(healthPre, tituloHealthAntesEnviar, healthFilterNamesFromCodes(cfg, bancosPre)...)
			}

			// Enviar actualización solest a cada banco modificado
			SolestStatus := make(map[string]int)
			var mu sync.Mutex
			if bancos, ok := accionDetalles["bancos"].([]string); ok {
				var wg sync.WaitGroup
				for _, codigoBanco := range bancos {
					wg.Add(1)
					go func(codigoBanco string) {
						defer wg.Done()
						url := "http://localhost:8080/simf/api/v1/config/solest?banco=" + codigoBanco + "&env=" + env
						client := &http.Client{Timeout: 10 * time.Second}
						resp, err := client.Post(url, "application/json", nil)
						if err != nil {
							fmt.Println("Error al enviar la configuración al banco", codigoBanco, ":", err)
							mu.Lock()
							if b, ok := cfg[codigoBanco]; ok {
								SolestStatus[b.Nombre] = 0
							}
							mu.Unlock()
							return
						}
						status := resp.StatusCode
						resp.Body.Close()
						mu.Lock()
						if b, ok := cfg[codigoBanco]; ok {
							SolestStatus[b.Nombre] = status
						}
						mu.Unlock()
						if status >= 400 {
							fmt.Println("\nBanco", codigoBanco, "respondió con status:", status)
						}
					}(codigoBanco)
				}
				wg.Wait()
			}

			// Health check al final de la opción elegida (después de guardar y enviar a bancos).
			// Se muestra la tabla para todas las opciones, incluida Activar/Desactivar MS en bancos.

			//if accionNombre == "activar_desactivar_ms" {
			//	fmt.Println("Health check después de Activar/Desactivar servicio de solicitud de estado:")
			//}
			fmt.Println("Verificando conexion, espere por favor:")
			health := api.RunHealthCheck(cfg)
			bancosPost := make([]string, 0)
			if bancos, ok := accionDetalles["bancos"].([]string); ok {
				bancosPost = bancos
			}
			tablaHealthCheck(health, tituloHealthDespuesEnviar, healthFilterNamesFromCodes(cfg, bancosPost)...)

			configPath, _ := handler.GetConfigPath(env)
			fmt.Println("\n\nConfiguración guardada en:", configPath)

			handler.LogAccion(accionNombre, env, accionDetalles)

			loading := src.NewLoadingModel("Cargando configuración...")
			if _, err := tea.NewProgram(loading).Run(); err != nil {
				fmt.Println("Se presentó un error al cargar la configuración:", err)
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
		{Title: "Código", Width: 10},
		{Title: "Nombre de Banco", Width: 30},
		{Title: "IP del servidor", Width: 20},
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
		m = src.NewTableModelSingleSelectWithHeader(t, "Lista de banco")
	} else {
		m = src.NewTableModelWithHeader(t, "Lista de bancos")
	}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}
	handler.ResetInactivity()
	tbl, ok := finalModel.(src.TableModel)
	if !ok {
		return nil, fmt.Errorf("no se pudo obtener el modelo de la tabla")
	}
	if !tbl.Confirmed {
		// Usuario pulsó Q (volver): no devolver selección para que el menú vuelva atrás
		return nil, nil
	}
	return tbl.SelectedRows(), nil
}

// tabla health check

func selectedCodesFromRows(rows []table.Row) []string {
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		if len(row) < 1 || row[0] == "" {
			continue
		}
		codes = append(codes, row[0])
	}
	return codes
}

func healthFilterNamesFromCodes(cfg handler.Config, codes []string) []string {
	names := make([]string, 0, len(codes))
	for _, code := range codes {
		if b, ok := cfg[code]; ok {
			names = append(names, b.Nombre)
		}
	}
	return names
}

func mostrarTablaConfig(cfg handler.Config, title string, filterCodes ...string) {
	if title == "" {
		title = tituloTablaCfg
	}
	if len(cfg) == 0 {
		fmt.Println("No hay datos de configuración para mostrar.")
		return
	}
	columns := []table.Column{
		{Title: "Código", Width: 10},
		{Title: "Nombre", Width: 22},
		{Title: "Endpoint", Width: 36},
		{Title: "IP", Width: 16},
		{Title: "Activar", Width: 8},
		{Title: "Tiempo Bucle", Width: 12},
		{Title: "Sol/Bucle", Width: 10},
		{Title: "Max Sol/Op", Width: 11},
	}
	filterSet := make(map[string]struct{}, len(filterCodes))
	for _, code := range filterCodes {
		if code == "" {
			continue
		}
		filterSet[code] = struct{}{}
	}
	applyFilter := len(filterSet) > 0

	codes := make([]string, 0, len(cfg))
	for code := range cfg {
		if applyFilter {
			if _, ok := filterSet[code]; !ok {
				continue
			}
		}
		codes = append(codes, code)
	}
	sort.Strings(codes)
	if len(codes) == 0 {
		fmt.Println("No hay bancos seleccionados para mostrar.")
		return
	}
	rows := make([]table.Row, 0, len(codes))
	for _, code := range codes {
		b := cfg[code]
		activar := "No"
		if b.Envio.Activar {
			activar = "Sí"
		}
		rows = append(rows, table.Row{
			code,
			b.Nombre,
			b.Endpoint,
			b.IP,
			activar,
			strconv.Itoa(b.Envio.TiempoDeEsperaEntreBucles),
			strconv.Itoa(b.Envio.NumeroDeSolicitudesPorBucle),
			strconv.Itoa(b.Envio.NumeroMaximoDeSolicitudesPorOperacion),
		})
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
	m := src.NewTableViewModelWithHeader(t, title)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error al mostrar tabla de configuración:", err)
	}
}

// healthEstadoParaTabla traduce el texto del health check a un estado claro para la tabla.
func healthEstadoParaTabla(raw string) string {
	switch {
	case raw == "sin IP en configuración":
		return "Sin IP en configuración"
	case strings.HasPrefix(raw, "OFFLINE:"):
		return "Sin conexion"
	case raw == "Conectado":
		return "Conectado"
	case strings.HasPrefix(raw, "HTTP "):
		return "Conectado"
	default:
		return raw
	}
}

func tablaHealthCheck(health api.HealthResponse, title string, filterBanks ...string) {
	if title == "" {
		title = tituloHealth
	}
	columns := []table.Column{
		{Title: "Banco", Width: 40},
		{Title: "Estado", Width: 40},
	}
	filterSet := make(map[string]struct{}, len(filterBanks))
	for _, bank := range filterBanks {
		if bank == "" {
			continue
		}
		filterSet[bank] = struct{}{}
	}
	applyFilter := len(filterSet) > 0

	names := make([]string, 0, len(health.Checks))
	for n := range health.Checks {
		if applyFilter {
			if _, ok := filterSet[n]; !ok {
				continue
			}
		}
		names = append(names, n)
	}
	sort.Strings(names)
	rows := make([]table.Row, 0, len(names))
	for _, n := range names {
		rows = append(rows, table.Row{n, healthEstadoParaTabla(health.Checks[n])})
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
	m := src.NewTableViewModelWithHeader(t, title)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error al mostrar health check:", err)
	}
}
