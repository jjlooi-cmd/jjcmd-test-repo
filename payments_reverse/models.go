package payments_reverse

// Reversal status values for PayNet DuitNow Reversal Issuer API.
// Ref: https://docs.developer.paynet.my/api-reference/v3/reversal/issuer
const (
	StatusSuccessful = "SUCCESSFUL"
	StatusNegative   = "NEGATIVE"
	StatusReject     = "REJECT"
)

// Transaction status and reason codes per PayNet API.
const (
	TransactionStatusACSP    = "ACSP"    // AcceptedSettlementInProcess
	TransactionStatusACTC   = "ACTC"    // AcceptedTechnical - accepted for further processing
	TransactionStatusRJCT   = "RJCT"    // Rejected
	ReasonCodeAccepted      = "U000"    // Success/Transaction Accepted
	ReasonCodeMissingField   = "API.005"  // Missing mandatory field
	ReasonCodeInvalidBody    = "API.001"  // Invalid request body / message validation
	ReasonCodeNameAccepted   = "ACCEPTED"
	ReasonCodeNameValidation = "MESSAGE_VALIDATION_ERROR"
	ReasonDescriptionAccepted = "Success/ Transaction Accepted"
)

// ReversalRequest is the incoming webhook payload for POST /webhooks/v3/payments/reverse.
// DuitNow Reversal - Issuer: RPP sends reversal request to Issuer (OFI) to reverse a credit transfer.
// Ref: https://docs.developer.paynet.my/api-reference/v3/reversal/issuer#/webhooks/webhooks-v3-payments-reverse/post
type ReversalRequest struct {
	AppHeader                    AppHeader           `json:"appHeader"`
	OriginalInstructionId        string              `json:"originalInstructionId,omitempty"`        // Reference to original credit transfer
	OriginalEndToEndId           string              `json:"originalEndToEndId,omitempty"`           // Original end-to-end id
	OriginalBusinessMessageId    string              `json:"originalBusinessMessageId,omitempty"`    // Original transfer business message id
	Debtor                       Party               `json:"debtor"`
	DebtorAccount                DebtorAccount       `json:"debtorAccount"`
	DebtorAgent                  Agent               `json:"debtorAgent"`
	Creditor                     Party               `json:"creditor"`
	CreditorAgent                Agent               `json:"creditorAgent"`
	CreditorAccount              CreditorAccount     `json:"creditorAccount"`
	InstructedAmount             InstructedAmount    `json:"instructedAmount"`
	ReversalReasonInformation    *ReversalReasonInfo `json:"reversalReasonInformation,omitempty"`
}

// ReversalReasonInfo - optional reason for the reversal.
type ReversalReasonInfo struct {
	Reason string `json:"reason,omitempty"`
}

type AppHeader struct {
	EndToEndId        string `json:"endToEndId"`
	TransactionId        string `json:"transactionId"`
	BusinessMessageId string `json:"businessMessageId"`
	CreationDateTime  string `json:"creationDateTime"`
}

type Party struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type DebtorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type"`
	ResidentStatus    string `json:"residentStatus"`
	ProductType       string `json:"productType"`
	ShariaCompliance  string `json:"shariaCompliance"`
	AccountHolderType string `json:"accountHolderType"`
}

type Agent struct {
	Id string `json:"id"`
}

type CreditorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type InstructedAmount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// ReversalResponse is the response body for POST /webhooks/v3/payments/reverse.
// Ref: https://docs.developer.paynet.my/api-reference/v3/reversal/issuer#/webhooks/webhooks-v3-payments-reverse/post#response-body
type ReversalResponse struct {
	AppHeader ResponseAppHeader `json:"appHeader"`
	Data      ResponseData      `json:"data"`
	Resp      ResponseStatus    `json:"resp"`
}

type ResponseAppHeader struct {
	EndToEndId                string `json:"endToEndId"`
	BusinessMessageId         string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
	TransactionId             string `json:"transactionId"`
}

// ResponseData - data block with settlement, creditor, creditorAccount (per spec).
type ResponseData struct {
	SettlementCycleNumber   string                  `json:"settlementCycleNumber"`
	InterbankSettlementDate string                  `json:"interbankSettlementDate"`
	Creditor                ResponseCreditor        `json:"creditor"`
	CreditorAccount         ResponseCreditorAccount `json:"creditorAccount"`
}

type ResponseCreditor struct {
	Name string `json:"name"`
}

type ResponseCreditorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type ResponseStatus struct {
	Status string         `json:"status"`
	Reason ResponseReason `json:"reason"`
}

type ResponseReason struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Details        string `json:"details,omitempty"`
	AdditionalCode string `json:"additionalCode,omitempty"`
}
