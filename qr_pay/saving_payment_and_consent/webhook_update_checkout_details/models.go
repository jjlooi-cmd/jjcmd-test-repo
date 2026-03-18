package webhook_update_checkout_details

// UpdateCheckoutDetailsRequest is the request body for the PayNet Webhook: Update Checkout Details (consent flow).
// This webhook maps the endToEndId to the checkoutId. It allows the acquirer to relate the endToEndId
// in the redirect URL back to the checkoutId when the issuer redirects with only the endToEndId (Step 14).
//
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/self-hosted-page/initiate-consent#webhook-update-checkout-details
type UpdateCheckoutDetailsRequest struct {
	// CheckoutId is the unique external identifier (uuid v4) provided by the acquirer to PayNet when initiating a payment intent. Max length: 36.
	CheckoutId string `json:"checkoutId"`
	// ConsentEndToEndId is the unique message identification from RPP. Can be used to reconcile with RPP BackOffice or Reports. Max length: 35.
	ConsentEndToEndId string `json:"consentEndToEndId"`
	// ConsentId is the consent that is authorized for AutoDebit payment. Max length: 35.
	ConsentId string `json:"consentId"`
	// Issuer is the name of payer's issuing bank / wallet. Max length: 100.
	Issuer string `json:"issuer"`
}
