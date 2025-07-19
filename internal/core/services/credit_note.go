package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"

	"ubl-converter/internal/pkg/signature"
	"ubl-converter/internal/pkg/ubl"
)

// CreditNoteRequest estructura para la solicitud de nota de crédito
type CreditNoteRequest struct {
	Emisor         EmisorData      `json:"emisor"`
	Receptor       ReceptorData    `json:"receptor"`
	Comprobante    ComprobanteData `json:"comprobante"`
	Detalle        []DetalleItem   `json:"detalle"`
	Motivo         string          `json:"motivo"`
	ComprobanteRef string          `json:"comprobante_ref"`
}



// Nota de crédito UBL con firma y extensiones
type UBLCreditNoteWithExtensions struct {
	XMLName    xml.Name            `xml:"CreditNote"`
	Xmlns      string              `xml:"xmlns,attr"`
	XmlnsExt   string              `xml:"xmlns:ext,attr"`
	XmlnsCac   string              `xml:"xmlns:cac,attr"`
	XmlnsCbc   string              `xml:"xmlns:cbc,attr"`
	Extensions CustomUBLExtensions `xml:"ext:UBLExtensions"`
	ubl.CreditNote
}

// ConvertToUBLCreditNote convierte una solicitud a una nota de crédito UBL firmada
func ConvertToUBLCreditNote(request *CreditNoteRequest) (string, error) {
	if request == nil {
		return "", fmt.Errorf("request cannot be nil")
	}

	totalIGV, err := strconv.ParseFloat(request.Comprobante.TotalIGV, 64)
	if err != nil {
		return "", fmt.Errorf("total IGV inválido: %v", err)
	}
	total, err := strconv.ParseFloat(request.Comprobante.Total, 64)
	if err != nil {
		return "", fmt.Errorf("total inválido: %v", err)
	}

	creditNote := ubl.CreditNote{
		UBLVersionID:         "2.1",
		CustomizationID:      "2.0",
		ID:                   fmt.Sprintf("%s-%s", request.Comprobante.Serie, request.Comprobante.Numero),
		IssueDate:            request.Comprobante.FechaEmision,
		IssueTime:            request.Comprobante.HoraEmision,
		DocumentCurrencyCode: request.Comprobante.Moneda,
		DiscrepancyResponse: []ubl.DiscrepancyResponse{{
			ReferenceID:  request.ComprobanteRef,
			ResponseCode: "01", // configurable si es necesario
			Description:  request.Motivo,
		}},
		BillingReference: []ubl.BillingReference{{
			InvoiceDocumentReference: ubl.InvoiceDocumentReference{
				ID: request.ComprobanteRef,
			},
		}},
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
		}},
		LegalMonetaryTotal: ubl.MonetaryTotal{
			PayableAmount: ubl.MonetaryAmount{
				Value:      total,
				CurrencyID: request.Comprobante.Moneda,
			},
		},
		Signature: ubl.Signature{
			ID: "signatureKG",
			SignatoryParty: ubl.SignatoryParty{
				PartyIdentification: []ubl.PartyIdentification{{
					ID: request.Emisor.RUC,
				}},
				PartyName: []ubl.PartyName{{
					Name: request.Emisor.RazonSocial,
				}},
			},
			DigitalSignatureAttachment: ubl.DigitalSignatureAttachment{
				ExternalReference: ubl.ExternalReference{
					URI: "#signatureKG",
				},
			},
		},
	}

	// Agregar líneas de detalle
	creditNote.CreditNoteLines = make([]ubl.CreditNoteLine, len(request.Detalle))
	for i, item := range request.Detalle {
		cantidad, _ := strconv.ParseFloat(item.Cantidad, 64)
		valorUnitario, _ := strconv.ParseFloat(item.ValorUnitario, 64)
		totalBase, _ := strconv.ParseFloat(item.TotalBase, 64)
		igv, _ := strconv.ParseFloat(item.IGV, 64)

		creditNote.CreditNoteLines[i] = ubl.CreditNoteLine{
			ID: strconv.Itoa(item.Item),
			CreditedQuantity: ubl.Quantity{
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

	// Serializar nota sin firma
	xmlBytes, err := xml.MarshalIndent(creditNote, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializando XML: %v", err)
	}

	// Cargar certificado
	certInfo, err := signature.LoadCertificate()
	if err != nil {
		return "", fmt.Errorf("error cargando certificado: %v", err)
	}

	// Firmar el XML
	signedXML, err := signature.SignXMLAsElement(string(xmlBytes), certInfo)
	if err != nil {
		return "", fmt.Errorf("error firmando XML: %v", err)
	}

	// Envolver en UBL con extensiones
	wrapped := UBLCreditNoteWithExtensions{
		Xmlns:    "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2",
		XmlnsExt: "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",
		XmlnsCac: "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCbc: "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		Extensions: CustomUBLExtensions{
			Extension: []CustomUBLExtension{{
				ExtensionContent: CustomExtensionContent{
					XML: signedXML,
				},
			}},
		},
		CreditNote: creditNote,
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
