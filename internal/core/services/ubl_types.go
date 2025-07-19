package services

// Estructuras auxiliares para la firma digital UBL.
// Estas estructuras son compartidas por notas de crédito, débito y facturas.

// CustomUBLExtensions contiene las extensiones UBL.
type CustomUBLExtensions struct {
	Extension []CustomUBLExtension `xml:"ext:UBLExtension"`
}

// CustomUBLExtension es una extensión individual.
type CustomUBLExtension struct {
	ExtensionContent CustomExtensionContent `xml:"ext:ExtensionContent"`
}

// CustomExtensionContent contiene el XML de la firma.
type CustomExtensionContent struct {
	XML string `xml:",innerxml"`
}
