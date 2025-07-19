package ubl

// Invoice represents a UBL invoice document
type Invoice struct {
	UBLVersionID         string `xml:"cbc:UBLVersionID"`
	CustomizationID      string `xml:"cbc:CustomizationID"`
	ID                   string `xml:"cbc:ID"`
	IssueDate            string `xml:"cbc:IssueDate"`
	IssueTime            string `xml:"cbc:IssueTime"`
	InvoiceTypeCode      string `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode string `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric     int    `xml:"cbc:LineCountNumeric,omitempty"`

	AccountingSupplierParty SupplierParty `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty CustomerParty `xml:"cac:AccountingCustomerParty"`
	TaxTotal                []TaxTotal    `xml:"cac:TaxTotal"`
	LegalMonetaryTotal      MonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines            []InvoiceLine `xml:"cac:InvoiceLine"`
}

// SupplierParty represents the supplier party in an invoice
type SupplierParty struct {
	CustomerAssignedAccountID string `xml:"cbc:CustomerAssignedAccountID"`
	Party                     Party  `xml:"cac:Party"`
}

// CustomerParty represents the customer party in an invoice
type CustomerParty struct {
	CustomerAssignedAccountID string `xml:"cbc:CustomerAssignedAccountID"`
	Party                     Party  `xml:"cac:Party"`
}

// Party represents a party (organization, person, etc.)
type Party struct {
	PartyName        []PartyName        `xml:"cac:PartyName"`
	PartyLegalEntity []PartyLegalEntity `xml:"cac:PartyLegalEntity"`
}

// PartyName represents the name of a party
type PartyName struct {
	Name string `xml:"cbc:Name"`
}

// PartyLegalEntity represents legal entity information
type PartyLegalEntity struct {
	RegistrationName string `xml:"cbc:RegistrationName"`
	CompanyID        string `xml:"cbc:CompanyID"`
}

// TaxTotal represents tax information
type TaxTotal struct {
	TaxAmount   MonetaryAmount `xml:"cbc:TaxAmount"`
	TaxSubtotal []TaxSubtotal  `xml:"cac:TaxSubtotal"`
}

// TaxSubtotal represents tax subtotal information
type TaxSubtotal struct {
	TaxableAmount MonetaryAmount `xml:"cbc:TaxableAmount"`
	TaxAmount     MonetaryAmount `xml:"cbc:TaxAmount"`
	TaxCategory   TaxCategory    `xml:"cac:TaxCategory"`
}

// TaxCategory represents tax category information
type TaxCategory struct {
	ID                     string    `xml:"cbc:ID"`
	Name                   string    `xml:"cbc:Name,omitempty"`
	Percent                float64   `xml:"cbc:Percent"`
	TaxExemptionReasonCode string    `xml:"cbc:TaxExemptionReasonCode,omitempty"`
	TaxScheme              TaxScheme `xml:"cac:TaxScheme"`
}

// TaxScheme represents tax scheme information
type TaxScheme struct {
	ID          string `xml:"cbc:ID"`
	Name        string `xml:"cbc:Name"`
	TaxTypeCode string `xml:"cbc:TaxTypeCode"`
}

// MonetaryTotal represents monetary total information
type MonetaryTotal struct {
	LineExtensionAmount MonetaryAmount `xml:"cbc:LineExtensionAmount"`
	TaxInclusiveAmount  MonetaryAmount `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount       MonetaryAmount `xml:"cbc:PayableAmount"`
}

// MonetaryAmount represents a monetary amount with currency
type MonetaryAmount struct {
	Value      float64 `xml:",chardata"`
	CurrencyID string  `xml:"currencyID,attr"`
}

// InvoiceLine represents an invoice line
type InvoiceLine struct {
	ID                  string         `xml:"cbc:ID"`
	InvoicedQuantity    Quantity       `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount MonetaryAmount `xml:"cbc:LineExtensionAmount"`
	TaxTotal            []TaxTotal     `xml:"cac:TaxTotal"`
	Item                Item           `xml:"cac:Item"`
	Price               Price          `xml:"cac:Price"`
}

// Quantity represents a quantity with unit code
type Quantity struct {
	Value    float64 `xml:",chardata"`
	UnitCode string  `xml:"unitCode,attr"`
}

// Item represents an item in an invoice line
type Item struct {
	Description string `xml:"cbc:Description"`
}

// Price represents a price in an invoice line
type Price struct {
	PriceAmount MonetaryAmount `xml:"cbc:PriceAmount"`
}

func NewInvoice() *Invoice {
	return &Invoice{
		UBLVersionID:    "2.1",
		CustomizationID: "2.0",
	}
}
