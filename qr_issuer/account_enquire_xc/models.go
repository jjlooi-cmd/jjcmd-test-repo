package account_enquire_xc

// EnquireRequest is the request payload for POST /v3/accounts/enquire-xc (Issuer).
// Schema from PayNet Merchant Presented QR Domestic - Issuer API.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/issuer/domestic#/paths/v3-accounts-enquire-xc/post
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
	EndToEndId        string `json:"endToEndId"`
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

type QR struct {
	Code string `json:"code"`
}

// EnquireResponse is the response body from POST /v3/accounts/enquire-xc.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/issuer/domestic#/paths/v3-accounts-enquire-xc/post
type EnquireResponse struct {
	AppHeader ResponseAppHeader `json:"appHeader"`
	Data      ResponseData      `json:"data"`
	Resp      ResponseStatus    `json:"resp"`
}

type ResponseAppHeader struct {
	EndToEndId                string `json:"endToEndId"`
	BusinessMessageId        string `json:"businessMessageId"`
	CreationDateTime          string `json:"creationDateTime"`
	OriginalBusinessMessageId string `json:"originalBusinessMessageId"`
}

type ResponseData struct {
	QR              ResponseQR              `json:"qr"`
	Creditor        ResponseCreditor        `json:"creditor"`
	CreditorAccount ResponseCreditorAccount `json:"creditorAccount"`
}

type ResponseQR struct {
	Category              string   `json:"category"`
	AcceptedSourceOfFunds []string `json:"acceptedSourceOfFunds"`
}

type ResponseCreditor struct {
	Name string `json:"name"`
}

type ResponseCreditorAccount struct {
	Id                string `json:"id"`
	Type              string `json:"type"`
	ResidentStatus    string `json:"residentStatus"`
	ProductType       string `json:"productType"`
	ShariaCompliance  string `json:"shariaCompliance"`
	AccountHolderType string `json:"accountHolderType"`
	CustomerCategory  string `json:"customerCategory"`
}

type ResponseStatus struct {
	Status string         `json:"status"`
	Reason ResponseReason `json:"reason"`
}

type ResponseReason struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}
