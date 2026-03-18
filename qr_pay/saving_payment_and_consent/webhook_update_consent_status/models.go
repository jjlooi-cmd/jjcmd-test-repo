package webhook_update_consent_status

// UpdateConsentDetailsRequest is the request body for the PayNet Webhook: Update Consent Details.
// This webhook updates the acquirer when a save payment method is initiated. It returns the consentId with the status.
// Acquirer shall provide an acknowledgement back to API Gateway.
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-consent#webhook-update-consent-details
type UpdateConsentDetailsRequest struct {
	// CheckoutId is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a payment intent. Max length: 36.
	CheckoutId string `json:"checkoutId"`
	// EndToEndId is the unique message identification from RPP. Can be used to reconcile with RPP BackOffice or Reports. Max length: 35.
	EndToEndId string `json:"endToEndId"`
	// Issuer is the name of payer's issuing bank / wallet. Max length: 100.
	Issuer string `json:"issuer"`
	// ConsentStatus contains the consent status details.
	ConsentStatus ConsentStatus `json:"consentStatus"`
}

// ConsentStatus holds the consent status from the webhook.
// Code refers to status codes (e.g. ACSP). Message refers to reason codes (e.g. U000).
type ConsentStatus struct {
	// ConsentId is the consent authorized for AutoDebit payment. Max length: 35.
	ConsentId string `json:"consentId"`
	// Code refers to status codes. Max length: 4.
	Code string `json:"code"`
	// Message refers to reason codes. Max length: 1024.
	Message string `json:"message"`
}
