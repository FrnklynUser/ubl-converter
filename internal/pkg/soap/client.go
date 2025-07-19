package soap

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SOAPClient interface para el cliente SOAP
type SOAPClient interface {
	Call(endpoint, soapAction string, request interface{}, response interface{}) error
}

type soapClient struct {
	isProd     bool
	httpClient *http.Client
	username   string
	password   string
}

// SOAPFault estructura para decodificar errores SOAP de SUNAT
type SOAPFault struct {
	XMLName    xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`
	FaultCode  string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
}

// NewSOAPClient crea una nueva instancia del cliente SOAP
func NewSOAPClient(isProd bool) SOAPClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	return &soapClient{
		isProd:     isProd,
		httpClient: client,
		username:   "20123456789MODDATOS", // Credenciales de prueba
		password:   "moddatos",            // Credenciales de prueba
	}
}

// Call realiza una llamada SOAP
func (c *soapClient) Call(endpoint, soapAction string, request interface{}, response interface{}) error {
	// Marshal del request a XML
	reqBody, err := xml.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshal request: %v", err)
	}

	// Limpiar el body del request
	cleanBody := strings.Replace(string(reqBody), `<?xml version="1.0" encoding="UTF-8"?>`, "", 1)
	cleanBody = strings.TrimSpace(cleanBody)

	// Construir envelope SOAP
	soapEnvelope := fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ser="http://service.sunat.gob.pe">
   <soapenv:Header>
      <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
         <wsse:UsernameToken>
            <wsse:Username>%s</wsse:Username>
            <wsse:Password>%s</wsse:Password>
         </wsse:UsernameToken>
      </wsse:Security>
   </soapenv:Header>
   <soapenv:Body>
      %s
   </soapenv:Body>
</soapenv:Envelope>`, c.username, c.password, cleanBody)

	// Crear request HTTP
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader([]byte(soapEnvelope)))
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	// Headers
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", soapAction)

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando request: %v", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %v", err)
	}

	// Log para depuración
	fmt.Printf("--- SUNAT Raw Response ---\n%s\n--------------------------\n", string(body))

	// Verificar código de respuesta
	if resp.StatusCode != http.StatusOK {
		// Intentar decodificar como SOAP Fault
		var soapFault SOAPFault
		if err := xml.Unmarshal(body, &soapFault); err == nil {
			return fmt.Errorf("error de SUNAT: %s - %s", soapFault.FaultCode, soapFault.FaultString)
		}
		return fmt.Errorf("error de SUNAT: %s", string(body))
	}

	// Extraer body del envelope
	bodyStr := string(body)
	startTag := ""
	endTag := ""

	if strings.Contains(bodyStr, "<soap:Body>") {
		startTag = "<soap:Body>"
		endTag = "</soap:Body>"
	} else if strings.Contains(bodyStr, "<soapenv:Body>") {
		startTag = "<soapenv:Body>"
		endTag = "</soapenv:Body>"
	} else if strings.Contains(bodyStr, "<env:Body>") {
		startTag = "<env:Body>"
		endTag = "</env:Body>"
	}

	bodyStart := strings.Index(bodyStr, startTag)
	bodyEnd := strings.Index(bodyStr, endTag)

	if bodyStart < 0 || bodyEnd < 0 {
		return fmt.Errorf("respuesta SOAP inválida: no se encontró el body")
	}

	respBody := bodyStr[bodyStart+len(startTag) : bodyEnd]
	if err := xml.Unmarshal([]byte(respBody), response); err != nil {
		return fmt.Errorf("error unmarshal response: %v", err)
	}

	return nil
}