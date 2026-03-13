package payments_transfer_xc

// Transaction status values for payment transfer response.
// See PayNet API: ACSP (AcceptedSettlementInProcess), RJCT (Rejected).
const (
	TransactionStatusACSP = "ACSP" // AcceptedSettlementInProcess - payment accepted for execution
	TransactionStatusRJCT = "RJCT" // Rejected
)

// PayNet API spec values per https://docs.developer.paynet.my/docs/duitNow-QR/response-codes
const (
	ReasonCodeAccepted       = "U000"   // Success/Transaction Accepted (with ACSP)
	ReasonCodeMissingField   = "API.005" // Missing mandatory field
	ReasonCodeInvalidBody    = "API.001" // Invalid request body / message validation
	ReasonCodeNameAccepted   = "ACCEPTED"
	ReasonCodeNameValidation = "MESSAGE_VALIDATION_ERROR"
	ReasonDescriptionAccepted = "Success/ Transaction Accepted"
)

// TransferRequest is the incoming webhook payload for POST /webhooks/v3/payments/transfer-xc.
// Schema from PayNet Merchant Presented QR Domestic - Acquirer API (payments transfer).
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-payments-transfer-xc/post
type TransferRequest struct {
	AppHeader         AppHeader         `json:"appHeader"`
	Debtor            Party             `json:"debtor"`
	DebtorAccount     DebtorAccount     `json:"debtorAccount"`
	DebtorAgent       Agent             `json:"debtorAgent"`
	CreditorAgent     Agent             `json:"creditorAgent"`
	CreditorAccount   CreditorAccount   `json:"creditorAccount"`
	QR                QR                `json:"qr"`
	InstructedAmount  InstructedAmount  `json:"instructedAmount"`
}

// AppHeader - application header with endToEndId, businessMessageId, creationDateTime.
type AppHeader struct {
	EndToEndId        string `json:"endToEndId"`
	BusinessMessageId string `json:"businessMessageId"`
	CreationDateTime  string `json:"creationDateTime"`
}

// Party - debtor or creditor party (id, name).
type Party struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// DebtorAccount - debtor account details.
type DebtorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type"`
	ResidentStatus    string `json:"residentStatus"`
	ProductType       string `json:"productType"`
	ShariaCompliance  string `json:"shariaCompliance"`
	AccountHolderType string `json:"accountHolderType"`
}

// Agent - debtor or creditor agent (e.g. BIC).
type Agent struct {
	Id string `json:"id"`
}

// CreditorAccount - creditor account id and type.
type CreditorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

// QR - QR code payload.
type QR struct {
	Code string `json:"code"`
}

// InstructedAmount - payment amount and currency.
type InstructedAmount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// TransferResponse matches the response body from PayNet API reference for payments-transfer-xc.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-payments-transfer-xc/post#response-body
type TransferResponse struct {
	AppHeader ResponseAppHeader `json:"appHeader"`
	Resp      ResponseStatus    `json:"resp"`
}

// ResponseAppHeader - appHeader in response; originalBusinessMessageId = request's businessMessageId.
type ResponseAppHeader struct {
	EndToEndId                string `json:"endToEndId"`
	BusinessMessageId        string `json:"businessMessageId"`
	CreationDateTime         string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
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
