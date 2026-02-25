package handler

import (
	"net/http"
	//"github.com/gorilla/mux"
)

func Routes(mux *http.ServeMux, config Config) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		CheckHealth(w, r, ConfigFile{
			Desarrollo: config,
			Produccion: config,
		})
	})
}
