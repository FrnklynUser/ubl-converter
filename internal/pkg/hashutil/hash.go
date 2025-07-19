package hashutil

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// HashXML calcula el hash SHA256 de un archivo XML
func HashXML(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error abriendo archivo: %v", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("error calculando hash: %v", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
