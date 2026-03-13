package account_enquire_xc

// TransactionStatus represents the outcome of an account enquiry.
// See PayNet API: SUCCESSFUL, NEGATIVE, REJECT.
const (
	StatusSuccessful = "SUCCESSFUL"
	StatusNegative   = "NEGATIVE"
	StatusReject     = "REJECT"
)

// Mock success values for PayNet response (HTTP 200).
const (
	TransactionStatusACSP    = "ACSP"    // AcceptedSettlementCompleted
	TransactionStatusReason  = "00"
	CategoryPointOfSales     = "POINT_OF_SALES"
)

// EnquireRequest is the incoming webhook payload for POST /webhooks/v3/accounts/enquire-xc.
// Schema from PayNet Merchant Presented QR Domestic - Acquirer API (sample request).
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-accounts-enquire-xc/post
type EnquireRequest struct {
	AppHeader       AppHeader       `json:"appHeader"`
	Debtor          Party           `json:"debtor"`
	DebtorAccount   DebtorAccount   `json:"debtorAccount"`
	DebtorAgent     Agent           `json:"debtorAgent"`
	CreditorAgent   Agent           `json:"creditorAgent"`
	CreditorAccount CreditorAccount `json:"creditorAccount"`
	QR              QR              `json:"qr"`
}

type AppHeader struct {
	EndToEndId         string `json:"endToEndId"`
	BusinessMessageId  string `json:"businessMessageId"`
	CreationDateTime   string `json:"creationDateTime"`
}

type Party struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type DebtorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type"`              // e.g. CURRENT
	ResidentStatus    string `json:"residentStatus"`    // e.g. RESIDENT
	ProductType       string `json:"productType"`       // e.g. ISLAMIC
	ShariaCompliance  string `json:"shariaCompliance"`  // e.g. YES
	AccountHolderType string `json:"accountHolderType"` // e.g. SINGLE
}

type Agent struct {
	Id string `json:"id"`
}

type CreditorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"` // e.g. DEFAULT
}

type QR struct {
	Code string `json:"code"`
}

// EnquireResponse is the webhook response for account enquiry.
// RPP expects this structure to relay to the issuer.
type EnquireResponse struct {
	// MessageId same as request businessMessageId or RPP-assigned (max 35)
	MessageId string `json:"messageId"`
	// TransactionStatus e.g. ACSP (AcceptedSettlementCompleted), or NEGATIVE/REJECT
	TransactionStatus string `json:"transactionStatus"`
	// TransactionStatusReason e.g. 00 for success
	TransactionStatusReason string `json:"transactionStatusReason,omitempty"`
	// Category e.g. POINT_OF_SALES
	Category string `json:"category,omitempty"`
	// BeneficiaryAccountName resolved creditor account name (when SUCCESSFUL)
	BeneficiaryAccountName string `json:"beneficiaryAccountName,omitempty"`
	// ReasonCode reason code for NEGATIVE/REJECT (e.g. RC.001)
	ReasonCode string `json:"reasonCode,omitempty"`
	// Message human-readable reason (max 1024)
	Message string `json:"message,omitempty"`
}
