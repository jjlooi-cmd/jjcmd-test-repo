package initiate_checkout

// InitiateCheckoutRequest is the request payload for POST /v1/bw/checkout (DuitNow Pay one-time payment, self-hosted page).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-checkout
type InitiateCheckoutRequest struct {
	// CheckoutID is the unique external identifier (UUID v4) provided by the acquirer to PayNet when initiating a payment intent.
	CheckoutID string `json:"checkoutId"`
	// TransactionFlow: 01 = Redirect Retail, 02 = Redirect Corporate. Optional; defaults to "01" if not set.
	TransactionFlow string   `json:"transactionFlow,omitempty"`
	Issuer          string   `json:"issuer"`                    // Required; name of payer's issuing bank/wallet (from bank list API)
	SourceOfFunds   []string `json:"sourceOfFunds"`             // Required; e.g. ["01"] for CASA
	Amount          string   `json:"amount"`                   // Required; MYR, two decimals (e.g. "10.00")
	Merchant        Merchant `json:"merchant"`
	MerchantName    string   `json:"merchantName,omitempty"`   // Optional; shown to user in checkout WebView
	MerchantRefID   string   `json:"merchantReferenceId"`      // Required; shown to user during authorization
	Customer        Customer `json:"customer"`
}

// Merchant object for initiate checkout.
type Merchant struct {
	ProductID string `json:"productId"` // Required; from PayNet Developer Portal
}

// Customer object for initiate checkout.
type Customer struct {
	Name                 string `json:"name"`                           // Required; payer name
	IdentityValidation   string `json:"identityValidation"`             // Required; 00/01/02/03
	IdentificationType   string `json:"identificationType,omitempty"`   // Optional; 01-05
	Identification       string `json:"identification,omitempty"`       // Conditional; required if identificationType set
}

// InitiateCheckoutResponse is the response body from POST /v1/bw/checkout.
type InitiateCheckoutResponse struct {
	Data    InitiateCheckoutData `json:"data"`
	Message string               `json:"message"`
}

// InitiateCheckoutData holds endToEndId, signature, issuer and payment method; endToEndIdSignature is used for browser redirection.
type InitiateCheckoutData struct {
	EndToEndID         string `json:"endToEndId"`         // Unique message ID from RPP; used to reconcile with RPP BackOffice or Reports
	EndToEndIDSignature string `json:"endToEndIdSignature"` // Signed with RPP private key; used to construct browser redirection
	Issuer             string `json:"issuer"`             // Payer's issuing bank/wallet name
	PaymentMethod      string `json:"paymentMethod"`      // e.g. "01" = DuitNow Online Banking / Wallets
}
