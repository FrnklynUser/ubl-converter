package soap

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
)

type Client struct {
	URL      string
	Username string
	Password string
}

const (
	TestURL       = "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService"
	ProductionURL = "https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService"
)

// NewClient crea un nuevo cliente SOAP para SUNAT
func NewClient(ruc, username, password string, isTest bool) *Client {
	url := ProductionURL
	if isTest {
		url = TestURL
	}

	return &Client{
		URL:      url,
		Username: fmt.Sprintf("%s%s", ruc, username),
		Password: password,
	}
}

// SendBill envía una factura a SUNAT
func (c *Client) SendBill(zipPath string) ([]byte, string, error) {
	// Leer el archivo ZIP
	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		return nil, "", fmt.Errorf("error leyendo ZIP: %v", err)
	}

	// Codificar el ZIP en base64
	encodedZip := base64.StdEncoding.EncodeToString(zipContent)

	// Preparar el request SOAP
	soapEnv := `
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" 
						 xmlns:ser="http://service.sunat.gob.pe" 
						 xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
			<soapenv:Header>
				<wsse:Security>
					<wsse:UsernameToken>
						<wsse:Username>{{.Username}}</wsse:Username>
						<wsse:Password>{{.Password}}</wsse:Password>
					</wsse:UsernameToken>
				</wsse:Security>
			</soapenv:Header>
			<soapenv:Body>
				<ser:sendBill>
					<fileName>{{.FileName}}</fileName>
					<contentFile>{{.Content}}</contentFile>
				</ser:sendBill>
			</soapenv:Body>
		</soapenv:Envelope>`

	// Crear el template
	tmpl, err := template.New("soap").Parse(soapEnv)
	if err != nil {
		return nil, "", fmt.Errorf("error creando template SOAP: %v", err)
	}

	// Preparar los datos
	data := struct {
		Username string
		Password string
		FileName string
		Content  string
	}{
		Username: c.Username,
		Password: c.Password,
		FileName: "20123456789-01-F001-1.zip",
		Content:  encodedZip,
	}

	// Ejecutar el template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, "", fmt.Errorf("error ejecutando template SOAP: %v", err)
	}

	// Crear el request HTTP
	req, err := http.NewRequest("POST", c.URL, &buf)
	if err != nil {
		return nil, "", fmt.Errorf("error creando request HTTP: %v", err)
	}

	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "urn:sendBill")

	// Enviar el request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error enviando request: %v", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error leyendo respuesta: %v", err)
	}

	// TODO: Parsear la respuesta SOAP y extraer el CDR y ticket
	// Por ahora devolvemos la respuesta como está
	return respBody, "TICKET-123", nil
}
