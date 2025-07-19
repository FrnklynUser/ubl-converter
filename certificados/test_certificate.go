package main

import (
	"fmt"
	"log"
	"path/filepath"
	"ubl-converter/internal/pkg/signature"
)

func main() {
	// Ruta al certificado convertido
	pfxPath := filepath.Join(".", "C23022479065_converted.pfx")
	password := "Franklin123" // Usa la contraseña correcta

	fmt.Println("=== Prueba de Certificado Convertido ===")
	fmt.Printf("Cargando certificado desde: %s\n", pfxPath)

	// Intentar cargar el certificado convertido
	cert, key, err := signature.LoadKeyPairFromPFX(pfxPath, password)
	if err != nil {
		log.Fatalf("Error cargando certificado convertido: %v", err)
	}

	fmt.Println("✅ Certificado convertido cargado exitosamente!")
	fmt.Printf("📋 Información del certificado:\n")
	fmt.Printf("   - Sujeto: %s\n", cert.Subject.String())
	fmt.Printf("   - Emisor: %s\n", cert.Issuer.String())
	fmt.Printf("   - Válido desde: %s\n", cert.NotBefore.Format("2006-01-02 15:04:05"))
	fmt.Printf("   - Válido hasta: %s\n", cert.NotAfter.Format("2006-01-02 15:04:05"))
	fmt.Printf("   - Número de serie: %s\n", cert.SerialNumber.String())
	
	if key != nil {
		fmt.Printf("🔑 Clave privada cargada correctamente (tamaño: %d bits)\n", key.Size()*8)
	}

	fmt.Println("\n🎉 ¡El problema del certificado ha sido resuelto!")
	fmt.Println("Ahora puedes usar el certificado C23022479065_converted.pfx en tu aplicación.")
}
