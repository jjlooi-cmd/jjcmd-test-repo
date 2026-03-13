package payments_reverse

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/sample-repo/jws_generation"
)

// Handler implements POST /webhooks/v3/payments/reverse for PayNet DuitNow Reversal - Issuer.
// Ref: https://docs.developer.paynet.my/api-reference/v3/reversal/issuer#/webhooks/webhooks-v3-payments-reverse/post
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[payments_reverse] Incoming request Authorization (token): %s", r.Header.Get("Authorization"))
	if r.Method != http.MethodPost {
		setPayNetResponseHeaders(w, r, "")
		writeJSON(w, http.StatusMethodNotAllowed, buildReversalResponse(ReversalRequest{}, TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonCodeNameValidation, "POST required", "", "", ""))
		return
	}

	var req ReversalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[payments_reverse] invalid JSON: %v", err)
		setPayNetResponseHeaders(w, r, "")
		writeJSON(w, http.StatusBadRequest, buildReversalResponse(ReversalRequest{}, TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonCodeNameValidation, "Request body must be valid JSON", "", "", ""))
		return
	}
	defer r.Body.Close()

	log.Printf("[payments_reverse] --- Incoming request ---")
	log.Printf("[payments_reverse] Method: %s URL: %s", r.Method, r.URL.String())
	log.Printf("[payments_reverse] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		r.Header.Get("X-Client-Id"),
		r.Header.Get("X-Api-Version"),
		r.Header.Get("x-business-message-id"))
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[payments_reverse] Body:\n%s", string(bodyBytes))
	log.Printf("[payments_reverse] ------------------------")

	businessMessageId := strings.TrimSpace(req.AppHeader.BusinessMessageId)
	if businessMessageId == "" {
		setPayNetResponseHeaders(w, r, "")
		writeReversalResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "appHeader.businessMessageId is required", "", "", "")
		return
	}

	// Issuer (OFI) response: use debtorAgent (Issuer BIC) for response business message ID.
	debtorBic := strings.TrimSpace(req.DebtorAgent.Id)
	responseBizMsgId := responseBusinessMessageId(businessMessageId, debtorBic)

	if strings.TrimSpace(req.DebtorAccount.Id) == "" {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeReversalResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "debtorAccount.id is required", "", "", "")
		return
	}

	amount := strings.TrimSpace(req.InstructedAmount.Amount)
	currency := strings.TrimSpace(req.InstructedAmount.Currency)
	if amount == "" {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeReversalResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "instructedAmount.amount is required", "", "", "")
		return
	}
	if currency == "" {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeReversalResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "instructedAmount.currency is required", "", "", "")
		return
	}

	// Issuer business logic: process reversal (credit back original debtor, validate original transaction).
	// This example uses a stub; replace with real reversal processing (e.g. lookup original transfer, credit debtor).
	accepted, creditorName := processReversal(req)
	if accepted {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeReversalResponse(w, http.StatusOK, req, TransactionStatusACTC, ReasonCodeAccepted, ReasonCodeNameAccepted, ReasonDescriptionAccepted, "", "", creditorName)
		return
	}
	setPayNetResponseHeaders(w, r, responseBizMsgId)
	writeReversalResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeNameValidation, ReasonCodeNameValidation, "Reversal not accepted", "", "", "")
}

// processReversal performs the reversal (Issuer side). Stub: accept when debtor account and amount are present; replace with real logic.
// Returns accepted, creditorName (for response data.creditor.name).
func processReversal(req ReversalRequest) (bool, string) {
	debtorAccountId := strings.TrimSpace(req.DebtorAccount.Id)
	debtorAgentId := strings.TrimSpace(req.DebtorAgent.Id)
	creditorName := strings.TrimSpace(req.Creditor.Name)
	if creditorName == "" {
		creditorName = "string" // placeholder per spec sample
	}
	// Stub: accept known test values; otherwise accept if debtor agent and account present.
	switch debtorAccountId {
	case "123456789", "22345678901":
		return true, creditorName
	}
	if debtorAgentId != "" && debtorAccountId != "" && strings.TrimSpace(req.InstructedAmount.Amount) != "" {
		return true, creditorName
	}
	return false, ""
}

func writeReversalResponse(w http.ResponseWriter, statusCode int, req ReversalRequest, txnStatus, reasonCode, reasonName, reasonDesc, reasonDetails, reasonAdditionalCode, creditorName string) {
	resp := buildReversalResponse(req, txnStatus, reasonCode, reasonName, reasonDesc, reasonDetails, reasonAdditionalCode, creditorName)

	bodyBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[payments_reverse] marshal response for JWS: %v", err)
		writeJSON(w, statusCode, resp)
		return
	}
	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Printf("[payments_reverse] load private key for JWS: %v", err)
		writeJSON(w, statusCode, resp)
		return
	}
	token, err := jws_generation.GenerateJWS(jws_generation.GenerateOptions{
		PrivateKey:        privateKey,
		Algorithm:         jws_generation.RS512,
		Issuer:            jwsIssuer,
		BusinessMessageID: resp.AppHeader.BusinessMessageId,
		CredentialKey:     jwsCredentialKey,
		PayloadForHash:    bodyBytes,
	})
	if err != nil {
		log.Printf("[payments_reverse] generate JWS: %v", err)
		writeJSON(w, statusCode, resp)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	writeJSON(w, statusCode, resp)
}

// buildReversalResponse builds the response per PayNet Reversal Issuer API spec (appHeader, data, resp).
// Ref: https://docs.developer.paynet.my/api-reference/v3/reversal/issuer#/webhooks/webhooks-v3-payments-reverse/post#response-body
func buildReversalResponse(req ReversalRequest, txnStatus, reasonCode, reasonName, reasonDescription, reasonDetails, reasonAdditionalCode, creditorName string) ReversalResponse {
	origBizMsgId := req.AppHeader.BusinessMessageId
	// debtorBic := strings.TrimSpace(req.DebtorAgent.Id)
	// responseBizMsgId := responseBusinessMessageId(origBizMsgId, debtorBic)
	transactionId := req.AppHeader.TransactionId
	interbankSettlementDate := time.Now().Format("2006-01-02")
	if creditorName == "" {
		creditorName = "string"
	}
	creditorAccountType := strings.TrimSpace(req.CreditorAccount.Type)
	if creditorAccountType == "" {
		creditorAccountType = "CURRENT"
	}
	return ReversalResponse{
		AppHeader: ResponseAppHeader{
			EndToEndId: req.AppHeader.EndToEndId,
			// BusinessMessageId:         responseBizMsgId,
			BusinessMessageId:         "MBBEMYKL",
			CreationDateTime:          req.AppHeader.CreationDateTime,
			OriginalBusinessMessageId: origBizMsgId,
			TransactionId:             transactionId,
		},
		Data: ResponseData{
			SettlementCycleNumber:   "001",
			InterbankSettlementDate: interbankSettlementDate,
			Creditor:                ResponseCreditor{Name: creditorName},
			CreditorAccount: ResponseCreditorAccount{
				Id:   req.CreditorAccount.Id,
				Type: creditorAccountType,
			},
		},
		Resp: ResponseStatus{
			// Status: txnStatus,
			Status: TransactionStatusACSP,
			Reason: ResponseReason{
				Name:           reasonName,
				Code:           "00",
				Description:    reasonDescription,
				Details:        reasonDetails,
				AdditionalCode: reasonAdditionalCode,
			},
		},
	}
}

const (
	jwsIssuer        = "RPPEMYKL"
	jwsCredentialKey = "64feb830"
)

const (
	bicStartPosition       = 8
	bicLength              = 8
	originatorCodePosition = 19
)

func responseBusinessMessageId(requestId string, bic string) string {
	if len(requestId) <= originatorCodePosition {
		return requestId
	}
	b := []byte(requestId)
	bic = strings.TrimSpace(bic)
	if bic != "" && len(b) >= bicStartPosition+bicLength {
		bicVal := bic
		if len(bicVal) > bicLength {
			bicVal = bicVal[:bicLength]
		} else if len(bicVal) < bicLength {
			bicVal = bicVal + strings.Repeat(" ", bicLength-len(bicVal))
		}
		copy(b[bicStartPosition:bicStartPosition+bicLength], bicVal)
	}
	b[originatorCodePosition] = 'R'
	return string(b)
}

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
	log.Printf("[payments_reverse] --- Outgoing response ---")
	log.Printf("[payments_reverse] HTTP %d", statusCode)
	log.Printf("[payments_reverse] Authorization (outgoing token): %s", w.Header().Get("Authorization"))
	log.Printf("[payments_reverse] Response Headers:")
	for k, v := range w.Header() {
		log.Printf("[payments_reverse]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[payments_reverse] Body:\n%s", string(bodyBytes))
	log.Printf("[payments_reverse] -------------------------")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
