package payments_transfer_xc

import "strconv"

// DecimalAmount is a float64 that marshals to JSON with exactly two decimal places (e.g. 10.00).
// Go's encoding/json marshals float64 10.0 as "10"; PayNet may expect "10.00". This type fixes that.
type DecimalAmount float64

// MarshalJSON implements json.Marshaler so value is written with 2 decimal places.
func (d DecimalAmount) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(d), 'f', 2, 64)), nil
}

// UnmarshalJSON implements json.Unmarshaler for round-trip and parsing.
func (d *DecimalAmount) UnmarshalJSON(data []byte) error {
	s := string(data)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*d = DecimalAmount(v)
	return nil
}

// TransferRequest is the request payload for POST /v3/payments/transfer-xc (Issuer).
// Schema: QRPaymentRequest from PayNet Merchant Presented QR Domestic - Issuer API spec.
// Ref: document (2).yaml § components.schemas.QRPaymentRequest
type TransferRequest struct {
	AppHeader                 AppHeader                 `json:"appHeader"`
	InterbankSettlementAmount  InterbankSettlementAmount  `json:"interbankSettlementAmount"`
	Debtor                     Party                      `json:"debtor"`
	DebtorAccount              DebtorAccount              `json:"debtorAccount"`
	DebtorAgent                Agent                      `json:"debtorAgent"`
	Creditor                   Creditor                   `json:"creditor"`
	CreditorAccount            CreditorAccount            `json:"creditorAccount"`
	CreditorAgent              Agent                      `json:"creditorAgent"`
	RecipientReference         string `json:"recipientReference"`
	PaymentDescription         string `json:"paymentDescription,omitempty"`
	QR                         QR                         `json:"qr"`
}

// AppHeader - application header; transactionId required (use same as businessMessageId per spec).
type AppHeader struct {
	EndToEndId        string `json:"endToEndId"`
	TransactionId     string `json:"transactionId"`
	BusinessMessageId string `json:"businessMessageId"`
	CreationDateTime  string `json:"creationDateTime"`
	PossibleDuplicate *bool  `json:"possibleDuplicate,omitempty"`
	CopyDuplicate     string `json:"copyDuplicate,omitempty"` // CODU | COPY | EXPY
}

// InterbankSettlementAmount - amount and currency (value in MYR, up to 2 decimal places).
type InterbankSettlementAmount struct {
	Value    DecimalAmount `json:"value"` // marshals as e.g. 10.00, not 10
	Currency string       `json:"currency,omitempty"` // default MYR
}

// Party - debtor party (id, name).
type Party struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// DebtorAccount - debtor account details.
type DebtorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type"` // CURRENT | SAVINGS | CREDIT_CARD | WALLET
	ResidentStatus    string `json:"residentStatus,omitempty"`
	ProductType       string `json:"productType,omitempty"`
	ShariaCompliance  string `json:"shariaCompliance,omitempty"`
	AccountHolderType string `json:"accountHolderType,omitempty"`
}

// Agent - debtor or creditor agent (e.g. BIC).
type Agent struct {
	Id string `json:"id"`
}

// Creditor - creditor (recipient) information; name required from QR Enquiry response.
type Creditor struct {
	Name string `json:"name"`
}

// CreditorAccount - receiving account; id required (QR ID from PayNet QR).
type CreditorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type,omitempty"`
}

// QR - QR payload; code, category, and acceptedSourceOfFunds required per spec.
type QR struct {
	Code                 string   `json:"code"`
	Category             string   `json:"category"`   // POINT_OF_SALES | PERSON_TO_PERSON
	AcceptedSourceOfFunds []string `json:"acceptedSourceOfFunds"`
	PromoCode            string   `json:"promoCode,omitempty"`
}

// TransferResponse is the response body from POST /v3/payments/transfer-xc (Issuer).
// Schema: QRPaymentResponse from PayNet Merchant Presented QR Domestic - Issuer API spec.
// Ref: document (2).yaml § components.schemas.QRPaymentResponse
type TransferResponse struct {
	AppHeader ResponseAppHeader `json:"appHeader"`
	Data      ResponseData      `json:"data"`
	Resp      ResponseStatus    `json:"resp"`
}

// ResponseAppHeader - appHeader in response.
type ResponseAppHeader struct {
	EndToEndId                string `json:"endToEndId"`
	BusinessMessageId         string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
	TransactionId             string `json:"transactionId"`
}

// ResponseData - data block; interbankSettlementDate, creditor, creditorAccount required.
type ResponseData struct {
	SettlementCycleNumber   string                  `json:"settlementCycleNumber,omitempty"`
	InterbankSettlementDate string                  `json:"interbankSettlementDate"`
	Creditor                ResponseCreditor        `json:"creditor"`
	CreditorAccount         ResponseCreditorAccount `json:"creditorAccount"`
	QR                      *ResponseQR            `json:"qr,omitempty"`
}

// ResponseQR - optional in response; category, acceptedSourceOfFunds, promoCode.
type ResponseQR struct {
	Category              string   `json:"category,omitempty"`
	AcceptedSourceOfFunds []string `json:"acceptedSourceOfFunds,omitempty"`
	PromoCode             string   `json:"promoCode,omitempty"`
}

// ResponseCreditor - creditor name in response.
type ResponseCreditor struct {
	Name string `json:"name"`
}

// ResponseCreditorAccount - creditor account in response.
type ResponseCreditorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type,omitempty"`
	ResidentStatus    string `json:"residentStatus,omitempty"`
	ProductType       string `json:"productType,omitempty"`
	ShariaCompliance  string `json:"shariaCompliance,omitempty"`
	AccountHolderType string `json:"accountHolderType,omitempty"`
	CustomerCategory  string `json:"customerCategory,omitempty"`
}

// ResponseStatus - resp block with status and reason.
type ResponseStatus struct {
	Status string         `json:"status"` // ACTC | ACSP | RJCT
	Reason ResponseReason `json:"reason"`
}

// ResponseReason - name, code, description required; details, additionalCode optional.
type ResponseReason struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Details        string `json:"details,omitempty"`
	AdditionalCode string `json:"additionalCode,omitempty"`
}
