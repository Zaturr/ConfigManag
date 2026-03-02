package handler

type ConfigFile struct {
	Produccion Config `json:"Produccion"`
	Desarrollo Config `json:"Desarrollo"`
}

type Bancos struct {
	Nombre   string `json:"nombre"`
	Endpoint string `json:"endpoint"`
	IP       string `json:"ip"`
	Envio    Envio  `json:"envio"`
}

type Envio struct {
	Activar                               bool `json:"Activar"`
	TiempoDeEsperaEntreBucles             int  `json:"TiempoDeEsperaEntreBucles"`
	NumeroDeSolicitudesPorBucle           int  `json:"NumeroDeSolicitudesPorBucle"`
	NumeroMaximoDeSolicitudesPorOperacion int  `json:"NumeroMaximoDeSolicitudesPorOperacion"`
}

type Config map[string]Bancos

type Credentials struct {
	Password string `json:"password"`
}
