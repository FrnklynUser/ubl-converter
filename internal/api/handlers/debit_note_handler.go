package handlers

import (
	"net/http"
	"ubl-converter/internal/core/services"
	"ubl-converter/internal/core/services/sunat"

	"github.com/gin-gonic/gin"
)

// DebitNoteHandler estructura para el manejador de notas de débito
type DebitNoteHandler struct {
	sunatService sunat.Service
}

// NewDebitNoteHandler crea una nueva instancia de DebitNoteHandler
func NewDebitNoteHandler(isProd bool) *DebitNoteHandler {
	return &DebitNoteHandler{
		sunatService: sunat.NewService(isProd),
	}
}

// Handle maneja la creación de una nota de débito
func (h *DebitNoteHandler) Handle(c *gin.Context) {
	var req services.DebitNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir a UBL y firmar
	xmlContent, err := services.ConvertToUBLDebitNote(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preparar y validar el documento
	invoiceID := req.Emisor.RUC + "-" + req.Comprobante.TipoComprobante + "-" + req.Comprobante.Serie + "-" + req.Comprobante.Numero
	result, err := h.sunatService.PrepareAndValidate(xmlContent, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enviar a SUNAT
	xmlPath, ok := result["file"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ruta XML no encontrada"})
		return
	}

	ticket, err := h.sunatService.SendDebitNote(xmlPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

			c.JSON(http.StatusOK, gin.H{"ticket": ticket, "document_id": invoiceID})
}
