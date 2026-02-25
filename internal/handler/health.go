package handler

import (
	"encoding/json"
	"net/http"
)

func CheckHealth(w http.ResponseWriter, r *http.Request, config ConfigFile){
	result := NetworkModel{
		CheckHealth: make(map[string]string),
		Status: "UP",
	}

	for codigo, banco := range config.Desarrollo{
		direccion, err := ObtenerIp(banco)
		if err != nil {
			result.CheckHealth[banco.Nombre] = err.Error()
			continue
	}
	if err := realizaTelnet(direccion); err != nil {
		result.CheckHealth[codigo] = "OFFLINE: " + err.Error()
		result.Status = "Error"
		continue
	}else{
		result.CheckHealth[banco.Nombre] = "ONLINE"
	}

  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(result)


}		