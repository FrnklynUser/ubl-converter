package services

// Item representa un ítem o producto en una factura
type Item struct {
	Codigo      string  // Código del producto
	Descripcion string  // Descripción del producto
	Cantidad    float64 // Cantidad
	PrecioUnit  float64 // Precio unitario
	Subtotal    float64 // Subtotal (sin IGV)
	IGV         float64 // IGV del ítem
}
