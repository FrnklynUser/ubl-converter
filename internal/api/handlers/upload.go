package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// UploadHandler maneja la subida de archivos XML
func UploadHandler(c *gin.Context) {
	// Obtener el archivo del request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se encontró el archivo en el request"})
		return
	}
	defer file.Close()

	// Validar la extensión del archivo
	if filepath.Ext(header.Filename) != ".xml" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo debe ser XML"})
		return
	}

	// Crear el directorio temporal si no existe
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando directorio temporal"})
		return
	}

	// Crear el archivo en el directorio temporal
	filename := filepath.Join(tempDir, header.Filename)
	out, err := os.Create(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creando archivo: %v", err)})
		return
	}
	defer out.Close()

	// Copiar el contenido del archivo
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error guardando archivo: %v", err)})
		return
	}

	// Leer el contenido del archivo para validarlo
	content, err := os.ReadFile(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error leyendo archivo: %v", err)})
		return
	}

	// TODO: Aquí puedes agregar la validación del XML contra el esquema XSD si lo deseas

	c.JSON(http.StatusOK, gin.H{
		"message": "Archivo cargado exitosamente",
		"file": gin.H{
			"name": header.Filename,
			"size": header.Size,
			"path": filename,
		},
		"content": string(content),
	})
}
