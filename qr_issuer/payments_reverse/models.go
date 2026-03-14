package payments_reverse

// Transaction status and reason codes per PayNet DuitNow Reversal API (document (4).yaml).
const (
	TransactionStatusACSP    = "ACSP"   // AcceptedSettlementInProcess
	TransactionStatusACTC    = "ACTC"   // AcceptedTechnical
	TransactionStatusRJCT    = "RJCT"   // Rejected
	ReasonCodeAccepted       = "U000"
	ReasonCodeMissingField   = "API.005"
	ReasonCodeInvalidBody    = "API.001"
	ReasonCodeNameAccepted   = "ACCEPTED"
	ReasonCodeNameValidation = "MESSAGE_VALIDATION_ERROR"
	ReasonDescriptionAccepted = "Success/ Transaction Accepted"
)

// PaymentReverseRequest is the incoming webhook payload for POST /v3/payments/reverse (DuitNow Reversal).
// PayNet sends this to the Issuer when an Acquirer requests a reversal.
// Ref: document (4).yaml — PaymentReverseRequest schema.
type PaymentReverseRequest struct {
	AppHeader                  ReverseAppHeader            `json:"appHeader"`
	InterbankSettlementAmount   InterbankSettlementAmount   `json:"interbankSettlementAmount"`
	Debtor                     ReverseParty                `json:"debtor"`
	DebtorAccount              ReverseDebtorAccount        `json:"debtorAccount"`
	DebtorAgent                ReverseAgent                `json:"debtorAgent"`
	Creditor                   ReverseParty                `json:"creditor"`
	CreditorAccount            ReverseCreditorAccount      `json:"creditorAccount"`
	CreditorAgent              ReverseAgent                `json:"creditorAgent"`
	PaymentDescription         string                      `json:"paymentDescription,omitempty"`
	AcceptedSourceOfFunds      []string                    `json:"acceptedSourceOfFunds,omitempty"`
}

type ReverseAppHeader struct {
	EndToEndId        string  `json:"endToEndId"`
	TransactionId     string  `json:"transactionId"`
	BusinessMessageId string  `json:"businessMessageId"`
	CreationDateTime  string  `json:"creationDateTime"`
	PossibleDuplicate *bool   `json:"possibleDuplicate,omitempty"`
	CopyDuplicate     string  `json:"copyDuplicate,omitempty"` // CODU, COPY, EXPY
}

type InterbankSettlementAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency,omitempty"` // default MYR
}

type ReverseParty struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"` // RET, COR
}

type ReverseDebtorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"` // DEFAULT, CURRENT, SAVINGS, CREDIT_CARD, WALLET, LOAN
}

type ReverseAgent struct {
	Id string `json:"id"`
}

type ReverseCreditorAccount struct {
	Id   string `json:"id"`
	Type string `json:"type"` // CURRENT, SAVINGS, CREDIT_CARD, WALLET, LOAN, DEFAULT, PROXY
}

// PaymentReverseResponse is the response body for POST /v3/payments/reverse (Issuer webhook).
// Ref: document (4).yaml — PaymentReverseResponse schema.
type PaymentReverseResponse struct {
	AppHeader ReverseResponseAppHeader `json:"appHeader"`
	Data      ReverseResponseData     `json:"data"`
	Resp      ReverseResponseStatus   `json:"resp"`
}

type ReverseResponseAppHeader struct {
	EndToEndId                string `json:"endToEndId"`
	BusinessMessageId         string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
	TransactionId             string `json:"transactionId"`
}

type ReverseResponseData struct {
	SettlementCycleNumber   string                       `json:"settlementCycleNumber,omitempty"`
	InterbankSettlementDate string                       `json:"interbankSettlementDate,omitempty"`
	Creditor                ReverseResponseCreditor     `json:"creditor"`
	CreditorAccount         ReverseResponseCreditorAcct `json:"creditorAccount"`
}

type ReverseResponseCreditor struct {
	Name string `json:"name"`
}

type ReverseResponseCreditorAcct struct {
	Id   string `json:"id"`
	Type string `json:"type,omitempty"`
}

type ReverseResponseStatus struct {
	Status string              `json:"status"`
	Reason ReverseResponseReason `json:"reason"`
}

type ReverseResponseReason struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Details        string `json:"details,omitempty"`
	AdditionalCode string `json:"additionalCode,omitempty"`
}
