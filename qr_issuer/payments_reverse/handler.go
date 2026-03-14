package payments_reverse

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/sample-repo/jws_generation"
)

const (
	bicStartPosition       = 8
	bicLength              = 8
	originatorCodePosition = 19
	jwsIssuer              = "MBBEMYKL"
	jwsCredentialKey       = "64feb830"
)

// Handler implements POST /webhooks/v3/payments/reverse for PayNet DuitNow Reversal - Issuer.
// Request: PaymentReverseWebhookRequest; 200: PaymentReverseResponse; 400: ErrorResponse.
// Ref: document (3).yaml - DuitNow Reversal Issuer.
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[payments_reverse] Incoming request Authorization (token): %s", r.Header.Get("Authorization"))
	if r.Method != http.MethodPost {
		setResponseHeaders(w, r, "")
		writeErrorResponse(w, r, "", TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonNameValidation, "POST required", "")
		return
	}

	var req PaymentReverseWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[payments_reverse] invalid JSON: %v", err)
		setResponseHeaders(w, r, "")
		writeErrorResponse(w, r, "", TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonNameValidation, ReasonDescInvalidBody, "")
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
		setResponseHeaders(w, r, "")
		writeErrorResponse(w, r, "", TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "appHeader.businessMessageId is required", "appHeader.businessMessageId")
		return
	}

	// Issuer BIC: we are the debtor in the reversal (debtorAgent = originating FI = us).
	issuerBic := strings.TrimSpace(req.DebtorAgent.Id)
	responseBizMsgId := responseBusinessMessageId(businessMessageId, issuerBic)

	// Validate required fields per schema.
	if strings.TrimSpace(req.SettlementCycleNumber) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "settlementCycleNumber is required", "settlementCycleNumber")
		return
	}
	if strings.TrimSpace(req.InterbankSettlementDate) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "interbankSettlementDate is required", "interbankSettlementDate")
		return
	}
	if strings.TrimSpace(req.AppHeader.EndToEndId) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "appHeader.endToEndId is required", "appHeader.endToEndId")
		return
	}
	if strings.TrimSpace(req.AppHeader.TransactionId) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "appHeader.transactionId is required", "appHeader.transactionId")
		return
	}
	if strings.TrimSpace(req.AppHeader.CreationDateTime) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "appHeader.creationDateTime is required", "appHeader.creationDateTime")
		return
	}
	if strings.TrimSpace(req.Debtor.Name) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "debtor.name is required", "debtor.name")
		return
	}
	if strings.TrimSpace(req.DebtorAccount.Id) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "debtorAccount.id is required", "debtorAccount.id")
		return
	}
	if strings.TrimSpace(req.DebtorAccount.Type) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "debtorAccount.type is required", "debtorAccount.type")
		return
	}
	if strings.TrimSpace(req.DebtorAgent.Id) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "debtorAgent.id is required", "debtorAgent.id")
		return
	}
	if strings.TrimSpace(req.Creditor.Name) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "creditor.name is required", "creditor.name")
		return
	}
	if strings.TrimSpace(req.CreditorAccount.Id) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "creditorAccount.id is required", "creditorAccount.id")
		return
	}
	if strings.TrimSpace(req.CreditorAccount.Type) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "creditorAccount.type is required", "creditorAccount.type")
		return
	}
	if strings.TrimSpace(req.CreditorAgent.Id) == "" {
		setResponseHeaders(w, r, responseBizMsgId)
		writeErrorResponse(w, r, responseBizMsgId, TransactionStatusRJCT, ReasonCodeMissingField, ReasonNameValidation, "creditorAgent.id is required", "creditorAgent.id")
		return
	}

	// Issuer business logic: process reversal (stub - accept and return ACSP).
	accepted, creditorName := processReversal(req)
	if !accepted {
		setResponseHeaders(w, r, responseBizMsgId)
		writeSuccessResponse(w, req, responseBizMsgId, TransactionStatusRJCT, ReasonNameValidation, ReasonCodeMissingField, "Reversal not accepted", "", "")
		return
	}

	setResponseHeaders(w, r, responseBizMsgId)
	writeSuccessResponse(w, req, responseBizMsgId, TransactionStatusACSP, ReasonNameAccepted, ReasonCodeAccepted, ReasonDescAccepted, "", creditorName)
}

// processReversal performs the reversal (Issuer side). Stub: accept when required fields present.
// Returns accepted, creditorName (for response data.creditor.name).
func processReversal(req PaymentReverseWebhookRequest) (bool, string) {
	creditorName := strings.TrimSpace(req.Creditor.Name)
	if creditorName == "" {
		creditorName = "Creditor"
	}
	// Stub: accept when all required validations already passed in handler.
	return true, creditorName
}

// responseBusinessMessageId builds response BMID: request BMID with BIC at 8-16 replaced by issuer BIC and originator 'R'.
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

func setResponseHeaders(w http.ResponseWriter, r *http.Request, businessMessageId string) {
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

func signAndWriteJSON(w http.ResponseWriter, statusCode int, businessMessageId string, body interface{}) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Printf("[payments_reverse] marshal response: %v", err)
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
		return
	}
	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Printf("[payments_reverse] load private key: %v", err)
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
		return
	}
	token, err := jws_generation.GenerateJWS(jws_generation.GenerateOptions{
		PrivateKey:        privateKey,
		Algorithm:         jws_generation.RS512,
		Issuer:            jwsIssuer,
		BusinessMessageID: businessMessageId,
		CredentialKey:     jwsCredentialKey,
		PayloadForHash:    bodyBytes,
	})
	if err != nil {
		log.Printf("[payments_reverse] generate JWS: %v", err)
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	log.Printf("[payments_reverse] --- Outgoing response --- HTTP %d", statusCode)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func writeSuccessResponse(w http.ResponseWriter, req PaymentReverseWebhookRequest, responseBizMsgId, status, reasonName, reasonCode, reasonDesc, reasonDetails, creditorName string) {
	if creditorName == "" {
		creditorName = strings.TrimSpace(req.Creditor.Name)
	}
	if creditorName == "" {
		creditorName = "Creditor"
	}
	creditorAccountType := strings.TrimSpace(req.CreditorAccount.Type)
	if creditorAccountType == "" {
		creditorAccountType = "DEFAULT"
	}
	interbankDate := strings.TrimSpace(req.InterbankSettlementDate)
	if interbankDate == "" {
		interbankDate = time.Now().Format("2006-01-02")
	}
	resp := PaymentReverseResponse{
		AppHeader: PaymentReverseResponseAppHeader{
			EndToEndId:                req.AppHeader.EndToEndId,
			BusinessMessageId:         responseBizMsgId,
			CreationDateTime:          time.Now().Format(time.RFC3339),
			OriginalBusinessMessageId: req.AppHeader.BusinessMessageId,
			TransactionId:             req.AppHeader.TransactionId,
		},
		Data: PaymentReverseResponseData{
			SettlementCycleNumber:   strings.TrimSpace(req.SettlementCycleNumber),
			InterbankSettlementDate: interbankDate,
			Creditor:                PaymentReverseResponseCreditor{Name: creditorName},
			CreditorAccount:         PaymentReverseResponseCreditorAcct{Id: req.CreditorAccount.Id, Type: creditorAccountType},
		},
		Resp: PaymentReverseResponseStatus{
			Status: status,
			Reason: PaymentReverseResponseReason{
				Name:        reasonName,
				Code:        reasonCode,
				Description: reasonDesc,
				Details:     reasonDetails,
			},
		},
	}
	signAndWriteJSON(w, http.StatusOK, responseBizMsgId, resp)
}

func writeErrorResponse(w http.ResponseWriter, r *http.Request, responseBizMsgId, status, reasonCode, reasonName, reasonDesc, errorLocation string) {
	origBizMsgId := r.Header.Get("x-business-message-id")
	if origBizMsgId == "" {
		origBizMsgId = responseBizMsgId
	}
	if responseBizMsgId == "" {
		responseBizMsgId = origBizMsgId
	}
	errResp := ErrorResponse{
		AppHeader: ErrorResponseAppHeader{
			OriginalBusinessMessageId: origBizMsgId,
			RejectionDateTime:         time.Now().Format(time.RFC3339),
		},
		Resp: ErrorResponseStatus{
			Status: TransactionStatusRJCT,
			Reason: ErrorResponseReason{
				Name:          reasonName,
				Code:          reasonCode,
				Description:   reasonDesc,
				ErrorLocation: errorLocation,
			},
		},
	}
	signAndWriteJSON(w, http.StatusBadRequest, responseBizMsgId, errResp)
}
