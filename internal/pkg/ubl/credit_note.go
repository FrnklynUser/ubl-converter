package ubl

// CreditNote represents a UBL credit note document
type CreditNote struct {
	UBLVersionID         string `xml:"cbc:UBLVersionID"`
	CustomizationID      string `xml:"cbc:CustomizationID"`
	ID                   string `xml:"cbc:ID"`
	IssueDate            string `xml:"cbc:IssueDate"`
	IssueTime            string `xml:"cbc:IssueTime"`
	Note                 string `xml:"cbc:Note,omitempty"`
	DocumentCurrencyCode string `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric     int    `xml:"cbc:LineCountNumeric,omitempty"`

	DiscrepancyResponse  []DiscrepancyResponse `xml:"cac:DiscrepancyResponse"`
	BillingReference     []BillingReference    `xml:"cac:BillingReference"`

	AccountingSupplierParty SupplierParty `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty CustomerParty `xml:"cac:AccountingCustomerParty"`
	TaxTotal                []TaxTotal    `xml:"cac:TaxTotal"`
	LegalMonetaryTotal      MonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	CreditNoteLines         []CreditNoteLine `xml:"cac:CreditNoteLine"`
	Signature               Signature        `xml:"cac:Signature"`
}

// DiscrepancyResponse represents a response to a discrepancy
 type DiscrepancyResponse struct {
	ReferenceID         string `xml:"cbc:ReferenceID"`
	ResponseCode        string `xml:"cbc:ResponseCode"`
	Description         string `xml:"cbc:Description"`
}

// BillingReference represents a reference to a billing document
type BillingReference struct {
	InvoiceDocumentReference InvoiceDocumentReference `xml:"cac:InvoiceDocumentReference"`
}

// InvoiceDocumentReference represents a reference to an invoice document
type InvoiceDocumentReference struct {
	ID           string `xml:"cbc:ID"`
	DocumentType string `xml:"cbc:DocumentTypeCode"`
}

// CreditNoteLine represents a credit note line
type CreditNoteLine struct {
	ID                  string         `xml:"cbc:ID"`
	CreditedQuantity    Quantity       `xml:"cbc:CreditedQuantity"`
	LineExtensionAmount MonetaryAmount `xml:"cbc:LineExtensionAmount"`
	TaxTotal            []TaxTotal     `xml:"cac:TaxTotal"`
	Item                Item           `xml:"cac:Item"`
	Price               Price          `xml:"cac:Price"`
}