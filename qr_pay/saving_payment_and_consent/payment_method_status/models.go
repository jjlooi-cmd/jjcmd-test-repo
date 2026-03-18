package payment_method_status

// PaymentMethodStatusResponse is the response body from GET /v1/bw/consent/request?checkoutId=... (Enquire Payment Method Status).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-method-status
type PaymentMethodStatusResponse struct {
	Data    PaymentMethodStatusData `json:"data"`
	Message string                  `json:"message"` // Reason code e.g. "U000"
}

// PaymentMethodStatusData holds consent status. TransactionStatus: PDNG=Pending, ACTV=Active, CANC=Cancelled, PDAU=Pending authorization, EXPI=Expired.
type PaymentMethodStatusData struct {
	MessageId          string `json:"messageId"`         // Unique message identification from RPP. Max length 35.
	TransactionStatus string `json:"transactionStatus"`  // PDNG, ACTV, CANC, PDAU, EXPI
	ConsentId         string `json:"consentId"`          // Consent authorized for AutoDebit. Max length 35.
	Issuer            string `json:"issuer"`             // Name of payer's issuing bank/wallet. Max length 100.
}

// PaymentMethodDetailsResponse is the response body from GET /v1/bw/consent?consentId=... (Enquire Payment Method Details).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/payment-details
type PaymentMethodDetailsResponse struct {
	Data    PaymentMethodDetailsData `json:"data"`
	Message string                  `json:"message"` // Reason code e.g. "U000"
}

// PaymentMethodDetailsData holds consent details from the payment-details API.
type PaymentMethodDetailsData struct {
	MessageId string  `json:"messageId"` // Unique message identification from RPP. Max length 35.
	Consent   ConsentDetails `json:"consent"`
	Issuer    string  `json:"issuer"`    // Name of payer's issuing bank/wallet. Max length 100.
}

// ConsentDetails is the consent object in Enquire Payment Method Details response.
// Frequency: 01=Unlimited, 02=Daily, 03=Weekly, 04=Monthly, 05=Quarterly, 06=Yearly.
type ConsentDetails struct {
	ConsentId               string `json:"consentId"`               // Consent authorized for AutoDebit. Max length 35.
	ConsentStatus           string `json:"consentStatus"`          // e.g. ACTV; refer to consent status mapping.
	EffectiveDate           string `json:"effectiveDate"`          // YYYY-MM-DD
	ExpiryDate              string `json:"expiryDate"`             // YYYY-MM-DD
	Frequency               string `json:"frequency"`              // 01-06
	AllowTerminatedByDebtor string `json:"allowTerminatedByDebtor"` // "true" or "false"
}
