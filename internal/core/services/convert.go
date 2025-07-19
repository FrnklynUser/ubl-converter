package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"

	"ubl-converter/internal/pkg/signature"
	"ubl-converter/internal/pkg/ubl"
)

// EmisorData estructura para los datos del emisor
type EmisorData struct {
	RUC          string `json:"ruc"`
	RazonSocial  string `json:"razon_social"`
	Direccion    string `json:"direccion"`
	Distrito     string `json:"distrito"`
	Provincia    string `json:"provincia"`
	Departamento string `json:"departamento"`
	Ubigeo       string `json:"ubigeo"`
}

// ReceptorData estructura para los datos del receptor
type ReceptorData struct {
	RUC         string `json:"ruc"`
	RazonSocial string `json:"razon_social"`
}

// ComprobanteData estructura para los datos del comprobante
type ComprobanteData struct {
	Serie           string `json:"serie"`
	Numero          string `json:"numero"`
	FechaEmision    string `json:"fecha_emision"`
	HoraEmision     string `json:"hora_emision"`
	TipoComprobante string `json:"tipo_comprobante"`
	Moneda          string `json:"moneda"`
	TotalGravado    string `json:"total_gravado"`
	TotalIGV        string `json:"total_igv"`
	Total           string `json:"total"`
}

// DetalleItem estructura para los items del detalle
type DetalleItem struct {
	Item              int    `json:"item"`
	Descripcion       string `json:"descripcion"`
	Cantidad          string `json:"cantidad"`
	ValorUnitario     string `json:"valor_unitario"`
	PrecioUnitario    string `json:"precio_unitario"`
	TipoPrecio        string `json:"tipo_precio"`
	IGV               string `json:"igv"`
	TipoAfectacionIGV string `json:"tipo_afectacion_igv"`
	TotalBase         string `json:"total_base"`
	PorcentajeIGV     string `json:"porcentaje_igv"`
	UnidadMedida      string `json:"unidad_medida"`
	Total             string `json:"total"`
}

// FacturaRequest estructura para la solicitud de conversión
type FacturaRequest struct {
	Emisor      EmisorData      `json:"emisor"`
	Receptor    ReceptorData    `json:"receptor"`
	Comprobante ComprobanteData `json:"comprobante"`
	Detalle     []DetalleItem   `json:"detalle"`
}

// UBLInvoiceWithExtensions estructura para la factura con extensiones
type UBLInvoiceWithExtensions struct {
	XMLName    xml.Name      `xml:"urn:oasis:names:specification:ubl:schema:xsd:Invoice-2 Invoice"`
	Extensions UBLExtensions `xml:"ext:UBLExtensions"`
	*ubl.Invoice
}

// UBLExtensions estructura para las extensiones UBL
type UBLExtensions struct {
	Extension []UBLExtension `xml:"ext:UBLExtension"`
}

// UBLExtension estructura para una extensión UBL
type UBLExtension struct {
	ExtensionContent ExtensionContent `xml:"ext:ExtensionContent"`
}

// ExtensionContent estructura para el contenido de una extensión
type ExtensionContent struct {
	XML string `xml:",innerxml"`
}

// ConvertirAUBL convierte los datos de la factura a formato UBL XML
func ConvertirAUBL(request *FacturaRequest) (string, error) {
	if request == nil {
		return "", fmt.Errorf("request cannot be nil")
	}

	// Validar datos requeridos
	if err := validateRequest(request); err != nil {
		return "", err
	}

	// Convertir valores string a float64
	totalGravado, err := strconv.ParseFloat(request.Comprobante.TotalGravado, 64)
	if err != nil {
		return "", fmt.Errorf("total gravado inválido: %v", err)
	}
	totalIGV, err := strconv.ParseFloat(request.Comprobante.TotalIGV, 64)
	if err != nil {
		return "", fmt.Errorf("total IGV inválido: %v", err)
	}
	total, err := strconv.ParseFloat(request.Comprobante.Total, 64)
	if err != nil {
		return "", fmt.Errorf("total inválido: %v", err)
	}

	// Construir estructura UBL base
	invoice := &ubl.Invoice{
		UBLVersionID:         "2.1",
		CustomizationID:      "2.0",
		ID:                   fmt.Sprintf("%s-%s", request.Comprobante.Serie, request.Comprobante.Numero),
		IssueDate:            request.Comprobante.FechaEmision,
		IssueTime:            request.Comprobante.HoraEmision,
		InvoiceTypeCode:      request.Comprobante.TipoComprobante,
		DocumentCurrencyCode: request.Comprobante.Moneda,

		AccountingSupplierParty: ubl.SupplierParty{
			CustomerAssignedAccountID: request.Emisor.RUC,
			Party: ubl.Party{
				PartyName: []ubl.PartyName{{Name: request.Emisor.RazonSocial}},
				PartyLegalEntity: []ubl.PartyLegalEntity{{
					RegistrationName: request.Emisor.RazonSocial,
					CompanyID:        request.Emisor.RUC,
				}},
			},
		},

		AccountingCustomerParty: ubl.CustomerParty{
			CustomerAssignedAccountID: request.Receptor.RUC,
			Party: ubl.Party{
				PartyLegalEntity: []ubl.PartyLegalEntity{{
					RegistrationName: request.Receptor.RazonSocial,
					CompanyID:        request.Receptor.RUC,
				}},
			},
		},

		TaxTotal: []ubl.TaxTotal{{
			TaxAmount: ubl.MonetaryAmount{
				Value:      totalIGV,
				CurrencyID: request.Comprobante.Moneda,
			},
			TaxSubtotal: []ubl.TaxSubtotal{{
				TaxableAmount: ubl.MonetaryAmount{
					Value:      totalGravado,
					CurrencyID: request.Comprobante.Moneda,
				},
				TaxAmount: ubl.MonetaryAmount{
					Value:      totalIGV,
					CurrencyID: request.Comprobante.Moneda,
				},
				TaxCategory: ubl.TaxCategory{
					ID:      "S",
					Percent: 18.0,
					TaxScheme: ubl.TaxScheme{
						ID:          "1000",
						Name:        "IGV",
						TaxTypeCode: "VAT",
					},
				},
			}},
		}},

		LegalMonetaryTotal: ubl.MonetaryTotal{
			LineExtensionAmount: ubl.MonetaryAmount{
				Value:      totalGravado,
				CurrencyID: request.Comprobante.Moneda,
			},
			TaxInclusiveAmount: ubl.MonetaryAmount{
				Value:      total,
				CurrencyID: request.Comprobante.Moneda,
			},
			PayableAmount: ubl.MonetaryAmount{
				Value:      total,
				CurrencyID: request.Comprobante.Moneda,
			},
		},
	}

	// Agregar líneas de detalle
	invoice.InvoiceLines = make([]ubl.InvoiceLine, len(request.Detalle))
	for i, item := range request.Detalle {
		cantidad, _ := strconv.ParseFloat(item.Cantidad, 64)
		valorUnitario, _ := strconv.ParseFloat(item.ValorUnitario, 64)
		totalBase, _ := strconv.ParseFloat(item.TotalBase, 64)
		igv, _ := strconv.ParseFloat(item.IGV, 64)
		porcentajeIGV, _ := strconv.ParseFloat(item.PorcentajeIGV, 64)

		invoice.InvoiceLines[i] = ubl.InvoiceLine{
			ID: strconv.Itoa(item.Item),
			InvoicedQuantity: ubl.Quantity{
				Value:    cantidad,
				UnitCode: item.UnidadMedida,
			},
			LineExtensionAmount: ubl.MonetaryAmount{
				Value:      totalBase,
				CurrencyID: request.Comprobante.Moneda,
			},
			TaxTotal: []ubl.TaxTotal{{
				TaxAmount: ubl.MonetaryAmount{
					Value:      igv,
					CurrencyID: request.Comprobante.Moneda,
				},
				TaxSubtotal: []ubl.TaxSubtotal{{
					TaxableAmount: ubl.MonetaryAmount{
						Value:      totalBase,
						CurrencyID: request.Comprobante.Moneda,
					},
					TaxAmount: ubl.MonetaryAmount{
						Value:      igv,
						CurrencyID: request.Comprobante.Moneda,
					},
					TaxCategory: ubl.TaxCategory{
						ID:                     "S",
						Percent:                porcentajeIGV,
						TaxExemptionReasonCode: item.TipoAfectacionIGV,
						TaxScheme: ubl.TaxScheme{
							ID:          "1000",
							Name:        "IGV",
							TaxTypeCode: "VAT",
						},
					},
				}},
			}},
			Item: ubl.Item{
				Description: item.Descripcion,
			},
			Price: ubl.Price{
				PriceAmount: ubl.MonetaryAmount{
					Value:      valorUnitario,
					CurrencyID: request.Comprobante.Moneda,
				},
			},
		}
	}

	// Serializar sin firma
	xmlBytes, err := xml.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializando XML: %v", err)
	}

	// Cargar certificado
	certInfo, err := signature.LoadCertificate()
	if err != nil {
		return "", fmt.Errorf("error cargando certificado: %v", err)
	}

	// Firmar XML
	signatureXML, err := signature.SignXMLAsElement(string(xmlBytes), certInfo)
	if err != nil {
		return "", fmt.Errorf("error firmando XML: %v", err)
	}

	// Crear factura con extensiones y firma
	wrapped := UBLInvoiceWithExtensions{
		Extensions: UBLExtensions{
			Extension: []UBLExtension{{
				ExtensionContent: ExtensionContent{
					XML: signatureXML,
				},
			}},
		},
		Invoice: invoice,
	}

	// Serializar XML final
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(wrapped); err != nil {
		return "", fmt.Errorf("error codificando XML final: %v", err)
	}

	return buf.String(), nil
}

func validateRequest(req *FacturaRequest) error {
	if req.Emisor.RUC == "" || len(req.Emisor.RUC) != 11 {
		return fmt.Errorf("RUC del emisor inválido")
	}
	if req.Receptor.RUC == "" || len(req.Receptor.RUC) != 11 {
		return fmt.Errorf("RUC del receptor inválido")
	}
	if req.Comprobante.Serie == "" || req.Comprobante.Numero == "" {
		return fmt.Errorf("serie y número son requeridos")
	}
	if len(req.Detalle) == 0 {
		return fmt.Errorf("el detalle no puede estar vacío")
	}
	return nil
}
