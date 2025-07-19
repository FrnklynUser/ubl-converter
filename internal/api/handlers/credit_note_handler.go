package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"ubl-converter/internal/core/services"
	"ubl-converter/internal/core/services/sunat"

	"github.com/gin-gonic/gin"
)

// CreditNoteHandler estructura para el manejador de notas de crédito
type CreditNoteHandler struct {
	sunatService sunat.Service
}

// NewCreditNoteHandler crea una nueva instancia de CreditNoteHandler
func NewCreditNoteHandler(isProd bool) *CreditNoteHandler {
	return &CreditNoteHandler{
		sunatService: sunat.NewService(isProd),
	}
}

// Handle maneja la creación de una nota de crédito
func (h *CreditNoteHandler) Handle(c *gin.Context) {
	var req services.CreditNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir a UBL y firmar
	xmlContent, err := services.ConvertToUBLCreditNote(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// El tipo de comprobante para nota de crédito es '07'
	invoiceID := fmt.Sprintf("%s-07-%s-%s", req.Emisor.RUC, req.Comprobante.Serie, req.Comprobante.Numero)

	// Guardar el XML en un archivo temporal
	tempDir := "temp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0755)
	}
	xmlPath := filepath.Join(tempDir, invoiceID+".xml")
	if err := ioutil.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error guardando XML: %v", err)})
		return
	}

	// Enviar a SUNAT
	ticket, err := h.sunatService.SendCreditNote(xmlPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

			c.JSON(http.StatusOK, gin.H{"ticket": ticket, "document_id": invoiceID})
}
