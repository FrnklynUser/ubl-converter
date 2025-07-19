package services

// ConvertService interfaz para conversión de documentos
type ConvertService interface {
	ConvertirAUBL(request *FacturaRequest) (string, error)
}

type convertService struct{}

// NewConvertService crea una nueva instancia del servicio de conversión
func NewConvertService() ConvertService {
	return &convertService{}
}

func (s *convertService) ConvertirAUBL(request *FacturaRequest) (string, error) {
	// Aquí iría la lógica de conversión a UBL
	return "", nil
}
