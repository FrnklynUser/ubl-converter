package routes

import (
	"ubl-converter/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(isProd bool) *gin.Engine {
	r := gin.Default()

	// Health check endpoints
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "UBL Converter API v1"})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Exponer archivos generados (XML, PDF, etc.)
	r.Static("/files", "./temp")

	api := r.Group("/api/v1")
	{
		// Core functionality endpoints
		convertHandler := handlers.NewConvertHandler(isProd)
		sendHandler := handlers.NewSendHandler(isProd)
		api.POST("/convert", convertHandler.ConvertirAUBL)
		api.POST("/send", sendHandler.Handle)

		creditNoteHandler := handlers.NewCreditNoteHandler(isProd)
		debitNoteHandler := handlers.NewDebitNoteHandler(isProd)
		api.POST("/credit-notes", creditNoteHandler.Handle)
		api.POST("/debit-notes", debitNoteHandler.Handle)

		// SUNAT consultation endpoints
		sunatHandler := handlers.NewSUNATHandler(isProd)
		sunat := api.Group("/sunat")
		{
			sunat.POST("/consulta-cdr", sunatHandler.ConsultaCDR)
			sunat.POST("/consulta-estado", sunatHandler.ConsultaEstado)
			sunat.GET("/consulta-ticket", sunatHandler.ConsultaTicket)
		}

		documentHandler := handlers.NewDocumentHandler(isProd)
		api.GET("/documents/:id/status", documentHandler.GetStatus)
		api.GET("/documents/:id/xml", documentHandler.GetXML)
		api.GET("/documents/:id/pdf", documentHandler.GetPDF)
	}

	return r
}
