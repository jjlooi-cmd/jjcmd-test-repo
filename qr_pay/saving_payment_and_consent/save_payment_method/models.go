package save_payment_method

// SavePaymentMethodRequest is the request payload for POST /v1/payment/intent with dataType "02" (Save Payment Method - DuitNow Consent).
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/paynet-hosted-page/save-payment-method
type SavePaymentMethodRequest struct {
	// DataType: "02" = Save payment method (redirect to save payment method WebView).
	DataType          string   `json:"dataType"`
	CheckoutID        string   `json:"checkoutId"`        // Required; UUID v4 from acquirer
	SourceOfFunds     []string `json:"sourceOfFunds"`     // Required; e.g. ["01"] for CASA
	MerchantName      string   `json:"merchantName,omitempty"`
	MerchantRefID     string   `json:"merchantReferenceId"` // Required; shown to user during authorization
	Merchant          Merchant `json:"merchant"`
	Customer          Customer `json:"customer"`
	Consent           Consent  `json:"consent"`
	Language          string   `json:"language"` // Required; "en" or "bm"
}

// Merchant object for save payment method (productId from PayNet Developer Portal).
type Merchant struct {
	ProductID string `json:"productId"`
}

// Customer object for save payment method.
type Customer struct {
	Name                 string `json:"name"`
	IdentityValidation   string `json:"identityValidation"`   // 00/01/02/03
	IdentificationType   string `json:"identificationType,omitempty"`   // Optional; 01-05
	Identification       string `json:"identification,omitempty"`        // Conditional; required if identificationType set
}

// Consent object for save payment method (dataType=02). Required when saving payment method.
// Frequency: 01=Unlimited, 02=Daily, 03=Weekly, 04=Monthly, 05=Quarterly, 06=Yearly.
type Consent struct {
	MaxAmount     string `json:"maxAmount"`     // Max payment amount, two decimals e.g. "500.00"
	EffectiveDate string `json:"effectiveDate"`  // YYYY-MM-DD
	ExpiryDate    string `json:"expiryDate"`    // YYYY-MM-DD
	Frequency     string `json:"frequency"`     // 01-06
}

// SavePaymentMethodResponse is the response body from POST /v1/payment/intent (save payment method).
type SavePaymentMethodResponse struct {
	Data    SavePaymentMethodData `json:"data"`
	Message string                 `json:"message"`
}

// SavePaymentMethodData holds session id and redirect URL.
type SavePaymentMethodData struct {
	ID          string `json:"id"`
	RedirectURL string `json:"redirectUrl"`
}
