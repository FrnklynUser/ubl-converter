package ziputil

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateZIP crea un archivo ZIP conteniendo el archivo especificado
func CreateZIP(sourcePath string, destPath string) error {
	// Crear el archivo ZIP
	zipfile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creando archivo ZIP: %v", err)
	}
	defer zipfile.Close()

	// Crear el writer ZIP
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// Abrir el archivo fuente
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error abriendo archivo fuente: %v", err)
	}
	defer file.Close()

	// Crear un nuevo archivo en el ZIP
	writer, err := archive.Create(filepath.Base(sourcePath))
	if err != nil {
		return fmt.Errorf("error creando archivo en ZIP: %v", err)
	}

	// Copiar el contenido al ZIP
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("error copiando archivo al ZIP: %v", err)
	}

	return nil
}
