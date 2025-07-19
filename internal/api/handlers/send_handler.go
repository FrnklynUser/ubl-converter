package handlers

import (
	"net/http"
	"ubl-converter/internal/core/services"
	"ubl-converter/internal/core/services/sunat"
	"ubl-converter/internal/pkg/pdfutil"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// SendHandler estructura para el manejador de envíos
type SendHandler struct {
	sunatService sunat.Service
}

// NewSendHandler crea una nueva instancia de SendHandler
func NewSendHandler(isProd bool) *SendHandler {
	return &SendHandler{
		sunatService: sunat.NewService(isProd),
	}
}

// Handle maneja el envío directo de la factura a SUNAT
func (h *SendHandler) Handle(c *gin.Context) {
	var req services.FacturaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Error en los datos de entrada",
			"error":   err.Error(),
		})
		return
	}

	// Convertir a XML
	xmlContent, err := services.ConvertirAUBL(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error convirtiendo a XML",
			"error":   err.Error(),
		})
		return
	}

	// Preparar y validar el documento
	invoiceID := req.Emisor.RUC + "-" + req.Comprobante.TipoComprobante + "-" + req.Comprobante.Serie + "-" + req.Comprobante.Numero
	result, err := h.sunatService.PrepareAndValidate(xmlContent, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error preparando documento",
			"error":   err.Error(),
		})
		return
	}

	// Generar PDF + QR
	xmlPath, ok := result["file"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "ruta XML no encontrada"})
		return
	}
	pdfPath := pdfutil.BuildPDFPath(invoiceID)
	if err := pdfutil.GenerateInvoicePDF(xmlPath, pdfPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	// Construir URL completa para el PDF
	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	// Asumiendo que la ruta para obtener el PDF es /document/{id}/pdf
	pdfURL := fmt.Sprintf("%s://%s/document/%s/pdf", scheme, host, invoiceID)

	// Extraer datos del resultado de forma segura
	estado, _ := result["estado"].(string)
	hash, _ := result["hash"].(string)
	cdrZip, _ := result["cdr_zip"].(string)
	xmlFirmado, _ := result["xml_firmado"].(string)

	// Guardar XML firmado en disco para futura recuperación
	xmlFirmadoPath := filepath.Join("temp", invoiceID+".xml")
	if err := os.WriteFile(xmlFirmadoPath, []byte(xmlFirmado), 0644); err != nil {
		// Loggear el error pero no interrumpir el flujo
		fmt.Printf("Error al guardar el XML firmado en disco: %v\n", err)
	}

	// Guardar en memoria
	docData := services.DocumentData{
		Status:     estado,
		XMLContent: xmlFirmado,
		PDFURL:     pdfURL,
		CDRZip:     cdrZip,
	}
	services.SaveDocument(invoiceID, docData)

	// Construir respuesta acorde a especificación
	response := gin.H{
		"estado":      estado,
		"hash":        hash,
		"cdr_zip":     cdrZip,
		"xml_firmado": xmlFirmado,
		"pdf_url":     pdfURL,
		"document_id": invoiceID,
	}

	c.JSON(http.StatusOK, response)
}
