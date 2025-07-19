package sunat

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"ubl-converter/internal/pkg/soap"
	"ubl-converter/internal/pkg/ziputil"
)

// Service interfaz para el servicio SUNAT
type Service interface {
	SendInvoice(filename string) (string, error)
	SendCreditNote(filename string) (string, error)
	SendDebitNote(filename string) (string, error)
	ConsultaCDR(ruc, tipo, serie, numero string) (string, error)
	ConsultaEstado(ruc, tipo, serie, numero string) (string, error)
	ConsultaTicket(ticket string) (string, error)
	PrepareAndValidate(xmlContent, invoiceID string) (map[string]interface{}, error)
}

type service struct {
	isProd     bool            // true para producción, false para pruebas
	XMLPath    string          // ruta de los archivos XML
	CertPath   string          // ruta de los certificados
	TempPath   string          // ruta de archivos temporales
	soapClient soap.SOAPClient // cliente SOAP para comunicación con SUNAT
}

// NewService crea una nueva instancia del servicio SUNAT
func NewService(isProd bool) Service {
	// Crear directorios si no existen
	paths := []string{"xml", "certificados", "temp"}
	for _, path := range paths {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(fmt.Sprintf("error creando directorio %s: %v", path, err))
		}
	}

	return &service{
		isProd:     isProd,
		XMLPath:    "xml",
		CertPath:   "certificados",
		TempPath:   "temp",
		soapClient: soap.NewSOAPClient(isProd),
	}
}

// SendInvoice envía una factura a SUNAT
func (s *service) SendInvoice(filename string) (string, error) {
	return s.sendBill(filename)
}

// SendCreditNote envía una nota de crédito a SUNAT
func (s *service) SendCreditNote(filename string) (string, error) {
	return s.sendBill(filename)
}

// SendDebitNote envía una nota de débito a SUNAT
func (s *service) SendDebitNote(filename string) (string, error) {
	return s.sendBill(filename)
}

func (s *service) sendBill(filename string) (string, error) {
	zipFile := filepath.Join(s.TempPath, filepath.Base(filename)+".zip")
	if err := ziputil.CreateZIP(filename, zipFile); err != nil {
		return "", fmt.Errorf("error creando ZIP: %v", err)
	}

	zipContent, err := ioutil.ReadFile(zipFile)
	if err != nil {
		return "", fmt.Errorf("error leyendo ZIP: %v", err)
	}

	encodedZip := base64.StdEncoding.EncodeToString(zipContent)

	request := &struct {
		XMLName     xml.Name `xml:"ser:sendBill"`
		FileName    string   `xml:"fileName"`
		ContentFile string   `xml:"contentFile"`
	}{
		FileName:    filepath.Base(zipFile),
		ContentFile: encodedZip,
	}

	response := &struct {
		XMLName xml.Name `xml:"sendBillResponse"`
		Ticket  string   `xml:"ticket"`
	}{}

	endpoint := s.getBillServiceEndpoint()
	if err := s.soapClient.Call(endpoint, "urn:sendBill", request, response); err != nil {
		return "", fmt.Errorf("error enviando a SUNAT: %v", err)
	}

	return response.Ticket, nil
}

// ConsultaCDR consulta el CDR de un comprobante
func (s *service) ConsultaCDR(ruc, tipo, serie, numero string) (string, error) {
	endpoint := s.getConsultServiceEndpoint()
	request := &struct {
		XMLName   xml.Name `xml:"getStatus"`
		RucEmisor string   `xml:"rucComprobante"`
		TipoComp  string   `xml:"tipoComprobante"`
		Serie     string   `xml:"serieComprobante"`
		Numero    string   `xml:"numeroComprobante"`
	}{
		RucEmisor: ruc,
		TipoComp:  tipo,
		Serie:     serie,
		Numero:    numero,
	}

	response := &struct {
		XMLName xml.Name `xml:"getStatusResponse"`
		Return  string   `xml:"statusResponse"`
	}{}

	if err := s.soapClient.Call(endpoint, "urn:getStatus", request, response); err != nil {
		return "", fmt.Errorf("error consultando CDR: %v", err)
	}

	return response.Return, nil
}

// ConsultaEstado consulta el estado de un comprobante
func (s *service) ConsultaEstado(ruc, tipo, serie, numero string) (string, error) {
	endpoint := s.getConsultServiceEndpoint()
	request := &struct {
		XMLName   xml.Name `xml:"getStatus"`
		RucEmisor string   `xml:"rucComprobante"`
		TipoComp  string   `xml:"tipoComprobante"`
		Serie     string   `xml:"serieComprobante"`
		Numero    string   `xml:"numeroComprobante"`
	}{
		RucEmisor: ruc,
		TipoComp:  tipo,
		Serie:     serie,
		Numero:    numero,
	}

	response := &struct {
		XMLName xml.Name `xml:"getStatusResponse"`
		Return  string   `xml:"statusResponse"`
	}{}

	if err := s.soapClient.Call(endpoint, "urn:getStatus", request, response); err != nil {
		return "", fmt.Errorf("error consultando estado: %v", err)
	}

	return response.Return, nil
}

// ConsultaTicket consulta el estado de un ticket
func (s *service) ConsultaTicket(ticket string) (string, error) {
	endpoint := s.getConsultServiceEndpoint()
	request := &struct {
		XMLName xml.Name `xml:"getStatus"`
		Ticket  string   `xml:"ticket"`
	}{
		Ticket: ticket,
	}

	response := &struct {
		XMLName xml.Name `xml:"getStatusResponse"`
		Return  string   `xml:"statusResponse"`
	}{}

	if err := s.soapClient.Call(endpoint, "urn:getStatus", request, response); err != nil {
		return "", fmt.Errorf("error consultando ticket: %v", err)
	}

	return response.Return, nil
}

func (s *service) getConsultServiceEndpoint() string {
	if s.isProd {
		return "https://e-factura.sunat.gob.pe/ol-it-wsconscpegem/billConsultService"
	}
	return "https://e-beta.sunat.gob.pe/ol-it-wsconscpegem-beta/billConsultService"
}

func (s *service) getBillServiceEndpoint() string {
	if s.isProd {
		return "https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService"
	}
	return "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService"
}

// PrepareAndValidate prepara y valida los archivos necesarios para el envío
func (s *service) PrepareAndValidate(xmlContent, invoiceID string) (map[string]interface{}, error) {
	// Validar que existan los certificados
	certFiles := []string{
		filepath.Join(s.CertPath, "C23022479065.crt"),
		filepath.Join(s.CertPath, "C23022479065.key"),
	}
	for _, file := range certFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, fmt.Errorf("archivo %s no encontrado", file)
		}
	}

	// Escribir el XML en un archivo temporal
	xmlFile := filepath.Join(s.TempPath, invoiceID+".xml")
	if err := ioutil.WriteFile(xmlFile, []byte(xmlContent), 0644); err != nil {
		return nil, fmt.Errorf("error escribiendo XML: %v", err)
	}

	// Calcular hash SHA256 del XML
	xmlBytes, _ := ioutil.ReadFile(xmlFile)
	h := sha256.Sum256(xmlBytes)
	hash := fmt.Sprintf("%x", h[:])

	// Base64 del XML firmado (mock)
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlBytes)

	// Generar ZIP mock como CDR
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	zf, _ := zw.Create(invoiceID + ".xml")
	zf.Write(xmlBytes)
	zw.Close()
	cdrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return map[string]interface{}{
		"estado":      "aceptado", // mock
		"hash":        hash,
		"cdr_zip":     cdrBase64,
		"xml_firmado": xmlBase64,
		"file":        xmlFile,
	}, nil
}
