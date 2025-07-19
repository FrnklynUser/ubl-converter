package ubl

// Signature representa el elemento cac:Signature en UBL
type Signature struct {
	ID                       string                   `xml:"cbc:ID"`
	SignatoryParty           SignatoryParty           `xml:"cac:SignatoryParty"`
	DigitalSignatureAttachment DigitalSignatureAttachment `xml:"cac:DigitalSignatureAttachment"`
}

// SignatoryParty representa el elemento cac:SignatoryParty en UBL
type SignatoryParty struct {
	PartyIdentification []PartyIdentification `xml:"cac:PartyIdentification"`
	PartyName           []PartyName           `xml:"cac:PartyName"`
}

// PartyIdentification representa el elemento cac:PartyIdentification en UBL
type PartyIdentification struct {
	ID string `xml:"cbc:ID"`
}

// DigitalSignatureAttachment representa el elemento cac:DigitalSignatureAttachment en UBL
type DigitalSignatureAttachment struct {
	ExternalReference ExternalReference `xml:"cac:ExternalReference"`
}

// ExternalReference representa el elemento cac:ExternalReference en UBL
type ExternalReference struct {
	URI string `xml:"cbc:URI"`
}
