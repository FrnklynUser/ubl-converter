package handlers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"ubl-converter/internal/core/services"
	"ubl-converter/internal/core/services/sunat"

	"github.com/gin-gonic/gin"
)

// DocumentHandler estructura para el manejador de documentos
type DocumentHandler struct {
	sunatService sunat.Service
}

// NewDocumentHandler crea una nueva instancia de DocumentHandler
func NewDocumentHandler(isProd bool) *DocumentHandler {
	return &DocumentHandler{
		sunatService: sunat.NewService(isProd),
	}
}

// GetStatus maneja la consulta de estado de un documento
func (h *DocumentHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")

	// Intentar obtener de la memoria primero
	if docData, found := services.GetDocument(id); found {
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"document_id": id,
			"estado":      docData.Status,
			"xml_url":     fmt.Sprintf("/document/%s/xml", id),
			"pdf_url":     docData.PDFURL,
			"cdr_zip_url": fmt.Sprintf("/document/%s/cdr", id), // Asumiendo una ruta para el CDR
		})
		return
	}

	parts := strings.Split(id, "-")
	if len(parts) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de documento inválido. Formato esperado: RUC-TIPO-SERIE-NUMERO"})
		return
	}

	ruc := parts[0]
	tipo := parts[1]
	serie := parts[2]
	numero := parts[3]

	status, err := h.sunatService.ConsultaEstado(ruc, tipo, serie, numero)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "status": status})
}

// GetXML maneja la obtención del XML de un documento
func (h *DocumentHandler) GetXML(c *gin.Context) {
	id := c.Param("id")

	var xmlContent []byte
	var err error

	// Intentar obtener de la memoria primero
	if docData, found := services.GetDocument(id); found {
		xmlContent = []byte(docData.XMLContent)
	} else {
		// Si no está en memoria, leer del archivo de respaldo
		filePath := filepath.Join("temp", id+".xml")
		xmlContent, err = ioutil.ReadFile(filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("XML no encontrado para el ID: %s", id)})
			return
		}
	}

	// Decodificar de Base64 si es necesario (el contenido del archivo también está codificado)
	decodedXML, err := base64.StdEncoding.DecodeString(string(xmlContent))
	if err != nil {
		// Si falla la decodificación, podría ser que no estaba en Base64. Servirlo directamente.
		c.Data(http.StatusOK, "application/xml", xmlContent)
		return
	}

	c.Data(http.StatusOK, "application/xml", decodedXML)
}

// GetPDF maneja la obtención del PDF de un documento
func (h *DocumentHandler) GetPDF(c *gin.Context) {
	id := c.Param("id")

	// El PDF se genera y guarda en 'temp' durante el envío.
	// El nombre del archivo es {id}.pdf
	filePath := filepath.Join("temp", id+".pdf")

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("PDF no encontrado para el ID: %s", id)})
		return
	}

	c.Data(http.StatusOK, "application/pdf", content)
}
