package account_enquire_xc

// TransactionStatus represents the outcome of an account enquiry.
// See PayNet API: SUCCESSFUL, NEGATIVE, REJECT.
const (
	StatusSuccessful = "SUCCESSFUL"
	StatusNegative   = "NEGATIVE"
	StatusReject     = "REJECT"
)

// PayNet API spec values per https://docs.developer.paynet.my/docs/duitNow-QR/response-codes
const (
	TransactionStatusACSP    = "ACSP"  // AcceptedSettlementInProcess - payment accepted for execution
	TransactionStatusRJCT    = "RJCT"  // Rejected
	ReasonCodeAccepted       = "U000"  // Success/Transaction Accepted (with ACSP)
	CategoryPointOfSales     = "POINT_OF_SALES"
	ReasonCodeMissingField   = "API.005" // Missing mandatory field
	ReasonCodeInvalidBody    = "API.001" // Invalid request body / message validation
	ReasonCodeNameAccepted   = "ACCEPTED"
	ReasonCodeNameValidation = "MESSAGE_VALIDATION_ERROR"
	ReasonCodeRecordNotFound     = "API.010" // Record Not Found
	ReasonCodeNameRecordNotFound = "RESOURCE_NOT_FOUND"
	ReasonDescriptionAccepted = "Success/ Transaction Accepted"
)

// AcceptedSourceOfFunds sample values per API reference response.
var AcceptedSourceOfFundsDefault = []string{"CASA", "CREDIT_CARD", "WALLET"}

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

// EnquireResponse matches the sample response body from PayNet API reference.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-accounts-enquire-xc/post#response-body
type EnquireResponse struct {
	AppHeader ResponseAppHeader `json:"appHeader"`
	Data      ResponseData      `json:"data"`
	Resp      ResponseStatus    `json:"resp"`
}

// ResponseAppHeader - appHeader in response; originalBusinessMessageId = request's businessMessageId (from RPP).
type ResponseAppHeader struct {
	EndToEndId               string `json:"endToEndId"`
	BusinessMessageId        string `json:"businessMessageId"`
	CreationDateTime         string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
}

// ResponseData - data block with qr, creditor, creditorAccount.
type ResponseData struct {
	QR              ResponseQR              `json:"qr"`
	Creditor        ResponseCreditor        `json:"creditor"`
	CreditorAccount ResponseCreditorAccount `json:"creditorAccount"`
}

// ResponseQR - category and acceptedSourceOfFunds.
type ResponseQR struct {
	Category             string   `json:"category"`
	AcceptedSourceOfFunds []string `json:"acceptedSourceOfFunds"`
}

// ResponseCreditor - creditor name only in response.
type ResponseCreditor struct {
	Name string `json:"name"`
}

// ResponseCreditorAccount - full creditor account in response (per sample).
type ResponseCreditorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type"`
	ResidentStatus    string `json:"residentStatus"`
	ProductType       string `json:"productType"`
	ShariaCompliance  string `json:"shariaCompliance"`
	AccountHolderType string `json:"accountHolderType"`
	CustomerCategory  string `json:"customerCategory"`
}

// ResponseStatus - resp block with status and reason.
type ResponseStatus struct {
	Status string         `json:"status"`
	Reason ResponseReason `json:"reason"`
}

// ResponseReason - name, code, description.
type ResponseReason struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}
