package config

// SUNATCredentials contiene las credenciales para el servicio de SUNAT
type SUNATCredentials struct {
	RUC      string
	Username string
	Password string
	URLBeta  string
	URLProd  string
}

// GetSUNATCredentials retorna las credenciales configuradas para SUNAT
func GetSUNATCredentials() SUNATCredentials {
	return SUNATCredentials{
		RUC:      "20103129061MODDATOS",
		Username: "MODDATOS",
		Password: "MODDATOS",
		URLBeta:  "https://e-beta.sunat.gob.pe:443/ol-ti-itcpfegem-beta/billService",
		URLProd:  "https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService",
	}
}
