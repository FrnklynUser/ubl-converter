package handlers

import (
	"net/http"

	"ubl-converter/internal/core/services/sunat"

	"github.com/gin-gonic/gin"
)

type SUNATHandler struct {
	sunatService sunat.Service
}

func NewSUNATHandler(isProd bool) *SUNATHandler {
	return &SUNATHandler{
		sunatService: sunat.NewService(isProd),
	}
}

type ConsultaRequest struct {
	RUC             string `json:"ruc" binding:"required"`
	TipoComprobante string `json:"tipo_comprobante" binding:"required"`
	Serie           string `json:"serie" binding:"required"`
	Numero          string `json:"numero" binding:"required"`
}

// GET /sunat/consulta-cdr
func (h *SUNATHandler) ConsultaCDR(c *gin.Context) {
	var req ConsultaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.sunatService.ConsultaCDR(req.RUC, req.TipoComprobante, req.Serie, req.Numero)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cdr": result})
}

// GET /sunat/consulta-estado
func (h *SUNATHandler) ConsultaEstado(c *gin.Context) {
	var req ConsultaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.sunatService.ConsultaEstado(req.RUC, req.TipoComprobante, req.Serie, req.Numero)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"estado": result})
}

type ConsultaTicketRequest struct {
	Ticket string `json:"ticket" binding:"required"`
}

// GET /sunat/consulta-ticket
func (h *SUNATHandler) ConsultaTicket(c *gin.Context) {
	var req ConsultaTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.sunatService.ConsultaTicket(req.Ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"estado": result})
}
