package handlers

import (
	"net/http"
	"ubl-converter/internal/core/services"
	"ubl-converter/internal/core/services/sunat"
	"ubl-converter/internal/pkg/pdfutil"

	"github.com/gin-gonic/gin"
)

// ConvertHandler estructura del handler de conversión
type ConvertHandler struct {
	sunatService sunat.Service
}

// NewConvertHandler crea una nueva instancia de ConvertHandler
func NewConvertHandler(isProd bool) *ConvertHandler {
	return &ConvertHandler{
		sunatService: sunat.NewService(isProd),
	}
}

// ConvertirAUBL maneja la conversión de facturas a formato UBL
func (h *ConvertHandler) ConvertirAUBL(c *gin.Context) {
	var request services.FacturaRequest

	// Parsear el JSON de entrada
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir a XML
	xmlContent, err := services.ConvertirAUBL(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preparar y validar el documento
	invoiceID := request.Comprobante.Serie + "-" + request.Comprobante.Numero
	result, err := h.sunatService.PrepareAndValidate(xmlContent, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generar PDF + QR
	xmlPath, ok := result["file"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ruta XML no encontrada en resultado"})
		return
	}
	pdfPath := pdfutil.BuildPDFPath(invoiceID)
	if err := pdfutil.GenerateInvoicePDF(xmlPath, pdfPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result["pdf"] = pdfPath

	// Devolver el resultado
	c.JSON(http.StatusOK, result)
}
