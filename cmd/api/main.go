package main

import (
	"log"
	"ubl-converter/internal/api/routes"
)

func main() {
	// Por defecto iniciamos en modo beta (isProd = false)
	r := routes.SetupRouter(false)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error iniciando el servidor:", err)
	}
}
