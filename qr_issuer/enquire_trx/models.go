package enquire_trx

// TransactionEnquiryRequest is the request payload for POST /v3/transactions/enquire.
// Schema from document (5).yaml - TransactionEnquiryRequest (RPP transaction code 630).
type TransactionEnquiryRequest struct {
	AppHeader     TrxEnquiryAppHeader `json:"appHeader"`
	DebtorAgent   Agent               `json:"debtorAgent"`
	CreditorAgent Agent               `json:"creditorAgent"`
}

// TrxEnquiryAppHeader holds business message id, creation time and transaction id.
type TrxEnquiryAppHeader struct {
	BusinessMessageId string `json:"businessMessageId"`
	CreationDateTime  string `json:"creationDateTime"`
	TransactionId     string `json:"transactionId"`
}

// Agent identifies a participant (BIC).
type Agent struct {
	Id string `json:"id"`
}

// TransactionEnquiryResponse is the response body from POST /v3/transactions/enquire.
type TransactionEnquiryResponse struct {
	AppHeader TrxEnquiryResponseAppHeader `json:"appHeader"`
	Data      TrxEnquiryData              `json:"data"`
	Resp      TrxEnquiryResp              `json:"resp"`
}

// TrxEnquiryResponseAppHeader holds response header fields.
type TrxEnquiryResponseAppHeader struct {
	BusinessMessageId        string `json:"businessMessageId"`
	CreationDateTime         string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
	TransactionId            string `json:"transactionId"`
}

// TrxEnquiryData holds the transaction enquiry result.
type TrxEnquiryData struct {
	Transaction               TrxEnquiryTransaction   `json:"transaction"`
	SettlementCycleNumber     string                  `json:"settlementCycleNumber"`
	InterbankSettlementAmount Amount                  `json:"interbankSettlementAmount"`
	ConvertedAmount           *Amount                 `json:"convertedAmount,omitempty"`
	InterbankSettlementDate   string                  `json:"interbankSettlementDate"`
	Debtor                    TrxEnquiryPartyName     `json:"debtor"`
	DebtorAgent               Agent                   `json:"debtorAgent"`
	Creditor                  TrxEnquiryPartyName     `json:"creditor"`
}

// TrxEnquiryTransaction holds status and reason.
type TrxEnquiryTransaction struct {
	Status string         `json:"status"` // ACTC, ACSP, RJCT
	Reason TrxEnquiryReason `json:"reason"`
}

// TrxEnquiryReason holds status reason details.
type TrxEnquiryReason struct {
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	Description   string  `json:"description"`
	Details       string  `json:"details,omitempty"`
	AdditionalCode string `json:"additionalCode,omitempty"`
}

// Amount holds value and currency.
type Amount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency,omitempty"`
}

// TrxEnquiryPartyName holds party name (debtor/creditor).
type TrxEnquiryPartyName struct {
	Name string `json:"name"`
}

// TrxEnquiryResp is the top-level resp block (status and reason).
type TrxEnquiryResp struct {
	Status string           `json:"status"`
	Reason TrxEnquiryReason `json:"reason"`
}
