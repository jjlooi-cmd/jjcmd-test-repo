package initiate_consent

// InitiateConsentRequest is the request payload for POST /v1/bw/consent (Initiate Consent - Save Payment Method, self-hosted page).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-consent
type InitiateConsentRequest struct {
	// CheckoutID is the unique external identifier (UUID v4) provided by the acquirer to PayNet when initiating a payment intent.
	CheckoutID string `json:"checkoutId"`
	// Issuer is the name of payer's issuing bank/wallet; obtain from the bank list API.
	Issuer string `json:"issuer"`
	// SourceOfFunds: 01=CASA, 02=Credit Card (not supported), 03=eWallet (not supported).
	SourceOfFunds []string `json:"sourceOfFunds"`
	Merchant      Merchant `json:"merchant"`
	// MerchantRefID is the payment reference shown to the user during authorization with their issuer.
	MerchantRefID string   `json:"merchantReferenceId"`
	Customer      Customer `json:"customer"`
	Consent       Consent  `json:"consent"`
}

// Merchant object for initiate consent.
type Merchant struct {
	ProductID string `json:"productId"` // Product ID assigned by PayNet during merchant registration in Developer Portal.
}

// Customer object for initiate consent.
type Customer struct {
	Name                 string `json:"name"`                           // Required; name of payer
	IdentityValidation   string `json:"identityValidation"`             // Required; 00=None, 01=Debtor Name, 02=Debtor ID, 03=Both
	IdentificationType   string `json:"identificationType,omitempty"`   // Optional; 01-05 (IC, Army, Passport, Registration, Mobile)
	Identification       string `json:"identification,omitempty"`        // Conditional; required if identificationType set
}

// Consent object for initiate consent. Max amount, effective/expiry dates, and frequency.
// Frequency: 01=Unlimited, 02=Daily, 03=Weekly, 04=Monthly, 05=Quarterly, 06=Yearly.
type Consent struct {
	MaxAmount     string `json:"maxAmount"`     // Maximum payment amount in two decimals, e.g. "10.00"
	EffectiveDate string `json:"effectiveDate"` // YYYY-MM-DD
	ExpiryDate    string `json:"expiryDate"`    // YYYY-MM-DD
	Frequency     string `json:"frequency"`     // 01-06
}

// InitiateConsentResponse is the response body from POST /v1/bw/consent.
type InitiateConsentResponse struct {
	Data    InitiateConsentData `json:"data"`
	Message string              `json:"message"` // Reason code, e.g. "U000"
}

// InitiateConsentData holds endToEndId, signature and issuer; endToEndIdSignature is used for browser redirection.
type InitiateConsentData struct {
	EndToEndID          string `json:"endToEndId"`          // Unique message ID from RPP; used to reconcile with RPP BackOffice or Reports
	EndToEndIDSignature string `json:"endToEndIdSignature"` // Signed with RPP private key; used to construct browser redirection
	Issuer              string `json:"issuer"`              // Payer's issuing bank/wallet name
}
