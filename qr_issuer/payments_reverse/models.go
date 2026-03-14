package payments_reverse

// Transaction status values per PayNet DuitNow Reversal Issuer API (document v3.0.0).
const (
	TransactionStatusACTC = "ACTC" // AcceptedTechnical - accepted for further processing
	TransactionStatusACSP = "ACSP" // AcceptedSettlementInProcess
	TransactionStatusRJCT = "RJCT" // Rejected
)

// Reason codes and names for response.
const (
	ReasonCodeAccepted        = "00"
	ReasonNameAccepted        = "ACCEPTED"
	ReasonDescAccepted        = "Success/Transaction Accepted"
	ReasonCodeMissingField    = "API.005"
	ReasonCodeInvalidBody     = "API.001"
	ReasonNameValidation      = "MESSAGE_VALIDATION_ERROR"
	ReasonDescMissingField    = "Missing mandatory field"
	ReasonDescInvalidBody     = "Request body must be valid JSON"
)

// PaymentReverseWebhookRequest is the request body for POST /webhooks/v3/payments/reverse (Issuer).
// CT Reversal Webhook Request - DuitNow Reversal Issuer.
// Ref: document (3).yaml components.schemas.PaymentReverseWebhookRequest
type PaymentReverseWebhookRequest struct {
	SettlementCycleNumber      string                      `json:"settlementCycleNumber"`
	InterbankSettlementDate    string                      `json:"interbankSettlementDate"`
	AppHeader                  PaymentReverseAppHeader     `json:"appHeader"`
	InterbankSettlementAmount  InterbankSettlementAmount   `json:"interbankSettlementAmount"`
	Debtor                     PaymentReverseParty         `json:"debtor"`
	DebtorAccount              PaymentReverseAccount       `json:"debtorAccount"`
	DebtorAgent                PaymentReverseAgent         `json:"debtorAgent"`
	Creditor                   PaymentReverseParty         `json:"creditor"`
	CreditorAccount            PaymentReverseAccount       `json:"creditorAccount"`
	CreditorAgent              PaymentReverseAgent         `json:"creditorAgent"`
	PaymentDescription         string                      `json:"paymentDescription,omitempty"`
	AcceptedSourceOfFunds      []string                    `json:"acceptedSourceOfFunds,omitempty"`
}

// PaymentReverseAppHeader - appHeader in reversal request.
type PaymentReverseAppHeader struct {
	EndToEndId        string  `json:"endToEndId"`
	TransactionId     string  `json:"transactionId"`
	BusinessMessageId string  `json:"businessMessageId"`
	CreationDateTime  string  `json:"creationDateTime"`
	PossibleDuplicate *bool   `json:"possibleDuplicate,omitempty"`
	CopyDuplicate     string  `json:"copyDuplicate,omitempty"` // CODU, COPY, EXPY
}

// InterbankSettlementAmount - amount with value and currency.
type InterbankSettlementAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency,omitempty"` // default MYR
}

// PaymentReverseParty - debtor or creditor (name required, type optional).
type PaymentReverseParty struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"` // RET, COR
}

// PaymentReverseAccount - account id and type.
type PaymentReverseAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"` // DEFAULT, CURRENT, SAVINGS, CREDIT_CARD, WALLET, LOAN, PROXY
}

// PaymentReverseAgent - BIC (id).
type PaymentReverseAgent struct {
	Id string `json:"id"`
}

// PaymentReverseResponse is the 200 response for POST /webhooks/v3/payments/reverse.
// Ref: document (3).yaml components.schemas.PaymentReverseResponse
type PaymentReverseResponse struct {
	AppHeader PaymentReverseResponseAppHeader `json:"appHeader"`
	Data      PaymentReverseResponseData      `json:"data"`
	Resp      PaymentReverseResponseStatus    `json:"resp"`
}

// PaymentReverseResponseAppHeader - response appHeader.
type PaymentReverseResponseAppHeader struct {
	EndToEndId                string `json:"endToEndId"`
	BusinessMessageId         string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
	TransactionId             string `json:"transactionId"`
}

// PaymentReverseResponseData - response data (creditor, creditorAccount required).
type PaymentReverseResponseData struct {
	SettlementCycleNumber   string                           `json:"settlementCycleNumber,omitempty"`
	InterbankSettlementDate string                           `json:"interbankSettlementDate,omitempty"`
	Creditor                PaymentReverseResponseCreditor   `json:"creditor"`
	CreditorAccount         PaymentReverseResponseCreditorAcct `json:"creditorAccount"`
}

// PaymentReverseResponseCreditor - creditor in response.
type PaymentReverseResponseCreditor struct {
	Name string `json:"name"`
}

// PaymentReverseResponseCreditorAcct - creditor account in response.
type PaymentReverseResponseCreditorAcct struct {
	Id   string `json:"id"`
	Type string `json:"type,omitempty"` // CURRENT, SAVINGS, CREDIT_CARD, WALLET, DEFAULT
}

// PaymentReverseResponseStatus - resp block (status + reason).
type PaymentReverseResponseStatus struct {
	Status string                          `json:"status"`
	Reason PaymentReverseResponseReason    `json:"reason"`
}

// PaymentReverseResponseReason - reason in response.
type PaymentReverseResponseReason struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Details        string `json:"details,omitempty"`
	AdditionalCode string `json:"additionalCode,omitempty"`
}

// ErrorResponse is the 400 response (API Validation Error).
// Ref: document (3).yaml components.schemas.ErrorResponse
type ErrorResponse struct {
	AppHeader ErrorResponseAppHeader `json:"appHeader"`
	Resp      ErrorResponseStatus    `json:"resp"`
}

// ErrorResponseAppHeader - appHeader for error response.
type ErrorResponseAppHeader struct {
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
	RejectionDateTime         string `json:"rejectionDateTime,omitempty"`
}

// ErrorResponseStatus - resp for error response.
type ErrorResponseStatus struct {
	Status string                    `json:"status"` // defaults to RJCT
	Reason ErrorResponseReason       `json:"reason"`
}

// ErrorResponseReason - reason in error response.
type ErrorResponseReason struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Details        string `json:"details,omitempty"`
	AdditionalCode string `json:"additionalCode,omitempty"`
	ErrorLocation  string `json:"errorLocation,omitempty"`
}
