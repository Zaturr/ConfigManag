package handler

import (
	"net"
	"net/url"
	"time"
)

type NetworkModel struct {
	Status      string            `json:"status"`
	CheckHealth map[string]string `json:"check_health"`
}

func ObtenerIp(banco Bancos) (string, error) {
	u, err := url.Parse(banco.Endpoint)
	if err != nil {
		return net.JoinHostPort(banco.IP, "8080"), nil
	}
	host, puerto, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		if host == "" {
			host = banco.IP
		}
		switch u.Scheme {
		case "https":
			puerto = "443"
		case "http":
			puerto = "80"
		default:
			puerto = "8080"
		}
		return net.JoinHostPort(host, puerto), nil
	}
	return net.JoinHostPort(host, puerto), nil
}

func realizaTelnet(direccion string) error {
	conn, err := net.DialTimeout("tcp", direccion, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
