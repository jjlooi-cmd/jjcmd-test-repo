package payment_intent

// PaymentIntentRequest is the request payload for POST /v1/payment/intent (DuitNow Pay one-time payment).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/payment-intent#send-the-payment-intent-request
type PaymentIntentRequest struct {
	// DataType: 01 = Payment (redirect to checkout WebView), 02 = Save payment method. Use "01" for one-time payment.
	DataType string `json:"dataType"`
	// TransactionFlow: 01 = Redirect Retail, 02 = Redirect Corporate. Optional; defaults to "01".
	TransactionFlow string   `json:"transactionFlow,omitempty"`
	CheckoutID      string   `json:"checkoutId"`      // Required; UUID v4 from acquirer
	SourceOfFunds   []string `json:"sourceOfFunds,omitempty"` // e.g. ["01"] for CASA
	Amount          string   `json:"amount"`          // Required; MYR, two decimals (e.g. "10.00")
	MerchantName    string   `json:"merchantName,omitempty"`
	MerchantRefID   string   `json:"merchantReferenceId"` // Required; shown to user during authorization
	Merchant        Merchant `json:"merchant"`
	Customer        Customer `json:"customer"`
	Language        string   `json:"language"` // Required; "en" or "bm"
}

// Merchant object for payment intent.
type Merchant struct {
	ProductID string `json:"productId"` // Required; from PayNet Developer Portal
}

// Customer object for payment intent.
type Customer struct {
	Name                 string `json:"name"`                           // Required; payer name
	IdentityValidation   string `json:"identityValidation"`              // Required; 00/01/02/03
	IdentificationType   string `json:"identificationType,omitempty"`     // Optional; 01-05
	Identification       string `json:"identification,omitempty"`         // Conditional; required if identificationType set
}

// PaymentIntentResponse is the response body from POST /v1/payment/intent.
type PaymentIntentResponse struct {
	Data    PaymentIntentData `json:"data"`
	Message string            `json:"message"`
}

// PaymentIntentData holds session id and redirect URL.
type PaymentIntentData struct {
	ID          string `json:"id"`          // Session ID from PayNet
	RedirectURL string `json:"redirectUrl"` // URL to open DuitNow Pay WebView
}
