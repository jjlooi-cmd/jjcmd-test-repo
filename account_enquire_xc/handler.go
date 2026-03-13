package account_enquire_xc

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Handler implements POST /webhooks/v3/accounts/enquire-xc for PayNet QR MPM Domestic Acquirer.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-accounts-enquire-xc/post
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		setPayNetResponseHeaders(w, r, "")
		writeJSON(w, http.StatusMethodNotAllowed, actualResponse(EnquireRequest{}, TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonCodeNameValidation, "POST required", "", nil, ""))
		return
	}

	var req EnquireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[account_enquire_xc] invalid JSON: %v", err)
		setPayNetResponseHeaders(w, r, "")
		writeJSON(w, http.StatusBadRequest, actualResponse(EnquireRequest{}, TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonCodeNameValidation, "Request body must be valid JSON", "", nil, ""))
		return
	}
	defer r.Body.Close()

	// Print request: headers and body
	log.Printf("[account_enquire_xc] --- Incoming request ---")
	log.Printf("[account_enquire_xc] Method: %s URL: %s", r.Method, r.URL.String())
	log.Printf("[account_enquire_xc] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		r.Header.Get("X-Client-Id"),
		r.Header.Get("X-Api-Version"),
		r.Header.Get("x-business-message-id"))
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[account_enquire_xc] Body:\n%s", string(bodyBytes))
	log.Printf("[account_enquire_xc] ------------------------")

	businessMessageId := strings.TrimSpace(req.AppHeader.BusinessMessageId)

	// Message validation: required appHeader.businessMessageId (API.005 = Missing mandatory field)
	if businessMessageId == "" {
		setPayNetResponseHeaders(w, r, "")
		writeEnquireResponse(w, http.StatusOK, req, StatusReject, ReasonCodeMissingField, ReasonCodeNameValidation, "appHeader.businessMessageId is required", "")
		return
	}

	// Business validation: creditor account id (API.005 = Missing mandatory field)
	if strings.TrimSpace(req.CreditorAccount.Id) == "" {
		setPayNetResponseHeaders(w, r, businessMessageId)
		writeEnquireResponse(w, http.StatusOK, req, StatusNegative, ReasonCodeMissingField, ReasonCodeNameValidation, "creditorAccount.id is required", "")
		return
	}

	// Acquirer business logic: resolve creditor account (e.g. DB lookup, internal API).
	// This example uses a stub resolver; replace with real lookup.
	status, reasonCode, message, accountName := resolveAccount(req)
	reasonCodeName := ""
	if status != StatusSuccessful && reasonCode != "" {
		reasonCodeName = ReasonCodeNameRecordNotFound
	}
	setPayNetResponseHeaders(w, r, businessMessageId)
	writeEnquireResponse(w, http.StatusOK, req, status, reasonCode, reasonCodeName, message, accountName)
}

// resolveAccount performs the account enquiry (acquirer side).
// Uses creditorAccount.id and creditorAgent.id; replace with actual lookup.
func resolveAccount(req EnquireRequest) (status, reasonCode, message, accountName string) {
	creditorAccountId := strings.TrimSpace(req.CreditorAccount.Id)
	creditorAgentId := strings.TrimSpace(req.CreditorAgent.Id)

	// Stub: accept known test values and return SUCCESSFUL; otherwise NEGATIVE.
	switch creditorAccountId {
	case "123456789", "22345678901":
		return StatusSuccessful, "", "", "CREDITOR ACCOUNT NAME"
	}
	// Optional: allow by agent (e.g. MBBEMYKL) for testing
	if creditorAgentId == "MBBEMYKL" && creditorAccountId != "" {
		return StatusSuccessful, "", "", "CREDITOR ACCOUNT NAME"
	}
	return StatusNegative, ReasonCodeRecordNotFound, "Beneficiary account not found or not eligible", ""
}

func writeEnquireResponse(w http.ResponseWriter, statusCode int, req EnquireRequest, status, reasonCode, reasonCodeName, message, accountName string) {
	txnStatus := TransactionStatusRJCT
	reasonCodeVal := reasonCode
	reasonName := reasonCodeName
	reasonDesc := message
	category := ""
	qrAcceptedFunds := []string(nil)
	if status == StatusSuccessful {
		txnStatus = TransactionStatusACSP
		reasonCodeVal = ReasonCodeAccepted
		reasonName = ReasonCodeNameAccepted
		reasonDesc = ReasonDescriptionAccepted
		category = CategoryPointOfSales
		qrAcceptedFunds = AcceptedSourceOfFundsDefault
	}
	resp := actualResponse(req, txnStatus, reasonCodeVal, reasonName, reasonDesc, category, qrAcceptedFunds, accountName)
	writeJSON(w, statusCode, resp)
}

// actualResponse builds the response per API reference sample (appHeader, data, resp).
func actualResponse(req EnquireRequest, status, reasonCode, reasonName, reasonDescription, qrCategory string, acceptedSourceOfFunds []string, creditorName string) EnquireResponse {
	origBizMsgId := req.AppHeader.BusinessMessageId
	return EnquireResponse{
		AppHeader: ResponseAppHeader{
			EndToEndId:                req.AppHeader.EndToEndId,
			BusinessMessageId:         req.AppHeader.BusinessMessageId,
			CreationDateTime:          req.AppHeader.CreationDateTime,
			OriginalBusinessMessageId: origBizMsgId,
		},
		Data: ResponseData{
			QR: ResponseQR{
				Category:              qrCategory,
				AcceptedSourceOfFunds: acceptedSourceOfFunds,
			},
			Creditor: ResponseCreditor{Name: creditorName},
			CreditorAccount: ResponseCreditorAccount{
				Id:                req.CreditorAccount.Id,
				Type:              req.CreditorAccount.Type,
				ResidentStatus:    "",
				ProductType:       "",
				ShariaCompliance:  "",
				AccountHolderType: "",
				CustomerCategory:  "",
			},
		},
		Resp: ResponseStatus{
			Status: status,
			Reason: ResponseReason{
				Name:        reasonName,
				Code:        reasonCode,
				Description: reasonDescription,
			},
		},
	}
}

// setPayNetResponseHeaders sets mandatory response headers for PayNet webhook (echo from request where applicable).
func setPayNetResponseHeaders(w http.ResponseWriter, r *http.Request, businessMessageId string) {
	w.Header().Set("Content-Type", "application/json")
	if v := r.Header.Get("X-Client-Id"); v != "" {
		w.Header().Set("X-Client-Id", v)
	}
	if v := r.Header.Get("X-Api-Version"); v != "" {
		w.Header().Set("X-Api-Version", v)
	}
	if businessMessageId != "" {
		w.Header().Set("x-business-message-id", businessMessageId)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	bodyBytes, _ := json.MarshalIndent(body, "", "  ")
	log.Printf("[account_enquire_xc] --- Outgoing response ---")
	log.Printf("[account_enquire_xc] HTTP %d", statusCode)
	log.Printf("[account_enquire_xc] Response Headers:")
	for k, v := range w.Header() {
		log.Printf("[account_enquire_xc]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[account_enquire_xc] Body:\n%s", string(bodyBytes))
	log.Printf("[account_enquire_xc] -------------------------")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
