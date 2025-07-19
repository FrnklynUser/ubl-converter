package pdfutil

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

// BasicInvoiceFields contiene datos mínimos para mostrar en el PDF y construir el QR.
// Solo se definen los campos utilizados; el XML completo de UBL es extenso.
// Adaptar los paths/etiquetas reales según necesidad.
//
// Nota: Si necesitas más campos, extiende esta estructura.

type InvoiceLine struct {
	ID          string `xml:"ID"`               // número de ítem
	Description string `xml:"Item>Description"` // <cac:Item><cbc:Description>
	Quantity    string `xml:"InvoicedQuantity"`
	Price       string `xml:"Price>PriceAmount"`
	LineTotal   string `xml:"LineExtensionAmount"`
}

type BasicInvoiceFields struct {
	XMLName xml.Name `xml:"Invoice"`

	// Serie y número: <cbc:ID>F001-123</cbc:ID>
	ID string `xml:"ID"`

	// Tipo de comprobante: 01=Factura, 03=Boleta
	InvoiceTypeCode string `xml:"InvoiceTypeCode"`

	// Fecha de emisión: <cbc:IssueDate>2025-07-18</cbc:IssueDate>
	IssueDate string `xml:"IssueDate"`

	// Totales dentro de <cac:LegalMonetaryTotal>
    LegalMonetaryTotal struct {
        LineExtensionAmount string `xml:"LineExtensionAmount"`
        PayableAmount       string `xml:"PayableAmount"`
    } `xml:"LegalMonetaryTotal"`

	// IGV total
	TaxTotal struct {
		Amount string `xml:"TaxAmount"`
	} `xml:"TaxTotal"`

	// RUC Emisor: <cac:AccountingSupplierParty><cbc:CustomerAssignedAccountID>20123456789</cbc:CustomerAssignedAccountID></cac:AccountingSupplierParty>
	SupplierParty struct {
		RUC string `xml:"CustomerAssignedAccountID"`
	} `xml:"AccountingSupplierParty"`

	InvoiceLines []InvoiceLine `xml:"InvoiceLine"`
}

// GenerateInvoicePDF genera un PDF con un QR usando los datos del XML.
// xmlPath: ruta del Invoice UBL
// pdfPath: ruta destino del PDF a crear.
func GenerateInvoicePDF(xmlPath, pdfPath string) error {
	data, err := ioutil.ReadFile(xmlPath)
	if err != nil {
		return fmt.Errorf("leer xml: %w", err)
	}

	var invoice BasicInvoiceFields
	if err := xml.Unmarshal(data, &invoice); err != nil {
		return fmt.Errorf("parsear xml: %w", err)
	}

	// Construir string para QR (ejemplo SUNAT)
	// Formato oficial: RUC|TipoDoc|Serie|Numero|Total|Fecha|...
	// Para demo solo algunos campos.
    docType := invoice.InvoiceTypeCode

    // Definir etiqueta de documento
    docLabel := "Factura"
    if invoice.InvoiceTypeCode == "03" || (len(invoice.ID) > 0 && invoice.ID[0] == 'B') {
        docLabel = "Boleta"
    }
    if docType == "" {
        if len(invoice.ID) > 0 && invoice.ID[0] == 'B' {
            docType = "03"
        } else {
            docType = "01"
        }
    }
    qrStr := fmt.Sprintf("%s|%s|%s|%s|%s|%s|", invoice.SupplierParty.RUC, docType, invoice.ID[:4], invoice.ID[5:], invoice.LegalMonetaryTotal.PayableAmount, invoice.IssueDate)

	pngBytes, err := qrcode.Encode(qrStr, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("generar QR: %w", err)
	}

	// PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetTitle(docLabel+" "+invoice.ID, false)

    // Registrar imagen
	opt := gofpdf.ImageOptions{ImageType: "PNG"}
	pdf.RegisterImageOptionsReader("qr.png", opt, bytes.NewReader(pngBytes))

    // Cabecera simple
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, docLabel+" "+invoice.ID)
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 8, "RUC Emisor: "+invoice.SupplierParty.RUC)
	pdf.Ln(8)
	pdf.Cell(40, 8, "Fecha Emision: "+invoice.IssueDate)
	pdf.Ln(8)
	pdf.Cell(40, 8, "Total: "+invoice.LegalMonetaryTotal.PayableAmount)
	pdf.Ln(12)

	// ---- Tabla de Ítems ----
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(10, 8, "N°", "1", 0, "C", false, 0, "")
	pdf.CellFormat(80, 8, "Descripción", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 8, "Cant.", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 8, "P.Unit", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 8, "Total", "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	for _, line := range invoice.InvoiceLines {
		pdf.CellFormat(10, 8, line.ID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(80, 8, line.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 8, line.Quantity, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 8, line.Price, "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 8, line.LineTotal, "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	// Totales
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(140, 8, "Op. Gravada")
	pdf.CellFormat(30, 8, invoice.LegalMonetaryTotal.LineExtensionAmount, "1", 0, "R", false, 0, "")
	pdf.Ln(-1)
	pdf.Cell(140, 8, "IGV")
	pdf.CellFormat(30, 8, invoice.TaxTotal.Amount, "1", 0, "R", false, 0, "")
	pdf.Ln(-1)
	pdf.Cell(140, 8, "Total")
	pdf.CellFormat(30, 8, invoice.LegalMonetaryTotal.PayableAmount, "1", 0, "R", false, 0, "")
	pdf.Ln(12)

	// Poner QR en esquina superior derecha
	pdf.ImageOptions("qr.png", 160, 10, 40, 40, false, opt, 0, "")

	// Guardar archivo
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		return fmt.Errorf("guardar pdf: %w", err)
	}

	return nil
}

// BuildPDFPath devuelve ruta destino en carpeta temp con extensión .pdf
func BuildPDFPath(invoiceID string) string {
	return filepath.Join("temp", fmt.Sprintf("%s.pdf", invoiceID))
}

// BuildQRString ejemplo simple externo si se requiere separado
func BuildQRString(ruc, serie, numero, total, fecha string) string {
	return fmt.Sprintf("%s|01|%s|%s|%s|%s|", ruc, serie, numero, total, fecha)
}
