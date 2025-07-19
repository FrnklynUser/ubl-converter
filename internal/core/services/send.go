package services

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	// SunatEndpoint URL del servicio web de SUNAT
	SunatEndpoint = "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService"
)

// SendBillRequest estructura para el envío del comprobante a SUNAT
type SendBillRequest struct {
	XMLFilename string
	RUCEmisor   string
	UserSOL     string
	PassSOL     string
}

// SendBillResponse estructura de respuesta del envío
type SendBillResponse struct {
	Success bool
	Message string
	CDR     []byte
}

// EnviarComprobante envía el comprobante a SUNAT
func EnviarComprobante(req *SendBillRequest) (*SendBillResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validar parámetros requeridos
	if err := validateSendRequest(req); err != nil {
		return nil, err
	}

	// Leer archivo XML
	xmlContent, err := os.ReadFile(req.XMLFilename)
	if err != nil {
		return nil, fmt.Errorf("error reading XML file: %v", err)
	}

	// Convertir XML a base64
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlContent)

	// Construir envelope SOAP
	soapEnvelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
						 xmlns:ser="http://service.sunat.gob.pe"
						 xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
			<soapenv:Header>
				<wsse:Security>
					<wsse:UsernameToken>
						<wsse:Username>%s%s</wsse:Username>
						<wsse:Password>%s</wsse:Password>
					</wsse:UsernameToken>
				</wsse:Security>
			</soapenv:Header>
			<soapenv:Body>
				<ser:sendBill>
					<fileName>%s.zip</fileName>
					<contentFile>%s</contentFile>
				</ser:sendBill>
			</soapenv:Body>
		</soapenv:Envelope>`, req.RUCEmisor, req.UserSOL, req.PassSOL,
		req.XMLFilename, xmlBase64)

	// Crear request HTTP
	httpReq, err := http.NewRequest("POST", SunatEndpoint, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "text/xml")
	httpReq.Header.Set("SOAPAction", "")

	// Enviar request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parsear respuesta SOAP
	var soapResponse struct {
		XMLName xml.Name
		Body    struct {
			SendBillResponse struct {
				ApplicationResponse string
			}
		}
	}

	if err := xml.Unmarshal(body, &soapResponse); err != nil {
		return nil, fmt.Errorf("error parsing SOAP response: %v", err)
	}

	// Decodificar CDR
	cdr, err := base64.StdEncoding.DecodeString(soapResponse.Body.SendBillResponse.ApplicationResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding CDR: %v", err)
	}

	// Guardar CDR en archivo
	cdrFilename := fmt.Sprintf("temp/%s_CDR.zip", req.XMLFilename)
	if err := os.WriteFile(cdrFilename, cdr, 0644); err != nil {
		return nil, fmt.Errorf("error saving CDR: %v", err)
	}

	return &SendBillResponse{
		Success: true,
		Message: "Comprobante enviado exitosamente",
		CDR:     cdr,
	}, nil
}

func validateSendRequest(req *SendBillRequest) error {
	if req.XMLFilename == "" {
		return fmt.Errorf("nombre de archivo XML requerido")
	}
	if req.RUCEmisor == "" || len(req.RUCEmisor) != 11 {
		return fmt.Errorf("RUC del emisor inválido")
	}
	if req.UserSOL == "" {
		return fmt.Errorf("usuario SOL requerido")
	}
	if req.PassSOL == "" {
		return fmt.Errorf("clave SOL requerida")
	}
	return nil
}
