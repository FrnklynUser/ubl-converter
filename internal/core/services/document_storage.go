package services

import "sync"

// DocumentData almacena la información de un documento procesado.
type DocumentData struct {
	Status     string
	XMLContent string
	PDFURL     string
	CDRZip     string
}

var (
	documentStore = make(map[string]DocumentData)
	storeMutex    = &sync.RWMutex{}
)

// SaveDocument guarda los datos de un documento en el almacenamiento en memoria.
func SaveDocument(id string, data DocumentData) {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	documentStore[id] = data
}

// GetDocument recupera los datos de un documento del almacenamiento en memoria.
func GetDocument(id string) (DocumentData, bool) {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	data, found := documentStore[id]
	return data, found
}
