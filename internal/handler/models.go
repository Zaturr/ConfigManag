package handler

type Bancos struct {
	Nombre    string `json:"nombre"`
	ActivarMS bool   `json:"activar_ms"`
	Endpoint  string `json:"endpoint"`
	IP        string `json:"ip"`
}

type Config map[string]Bancos
