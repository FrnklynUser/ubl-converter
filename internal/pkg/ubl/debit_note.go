package ubl

// DebitNote represents a UBL debit note document
type DebitNote struct {
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
	DebitNoteLines       []DebitNoteLine       `xml:"cac:DebitNoteLine"`

	AccountingSupplierParty SupplierParty `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty CustomerParty `xml:"cac:AccountingCustomerParty"`
	TaxTotal                []TaxTotal    `xml:"cac:TaxTotal"`
	RequestedMonetaryTotal  MonetaryTotal `xml:"cac:RequestedMonetaryTotal"`
	Signature               Signature             `xml:"cac:Signature"`
}

// DebitNoteLine represents a debit note line
type DebitNoteLine struct {
	ID                  string         `xml:"cbc:ID"`
	DebitedQuantity     Quantity       `xml:"cbc:DebitedQuantity"`
	LineExtensionAmount MonetaryAmount `xml:"cbc:LineExtensionAmount"`
	TaxTotal            []TaxTotal     `xml:"cac:TaxTotal"`
	Item                Item           `xml:"cac:Item"`
	Price               Price          `xml:"cac:Price"`
}