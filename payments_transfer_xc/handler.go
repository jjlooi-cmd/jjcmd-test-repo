package payments_transfer_xc

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/sample-repo/jws_generation"
)

// Handler implements POST /webhooks/v3/payments/transfer-xc for PayNet QR MPM Domestic Acquirer.
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-payments-transfer-xc/post
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[payments_transfer_xc] Incoming request Authorization (token): %s", r.Header.Get("Authorization"))
	if r.Method != http.MethodPost {
		setPayNetResponseHeaders(w, r, "")
		writeJSON(w, http.StatusMethodNotAllowed, buildTransferResponse(TransferRequest{}, TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonCodeNameValidation, "POST required", ""))
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[payments_transfer_xc] invalid JSON: %v", err)
		setPayNetResponseHeaders(w, r, "")
		writeJSON(w, http.StatusBadRequest, buildTransferResponse(TransferRequest{}, TransactionStatusRJCT, ReasonCodeInvalidBody, ReasonCodeNameValidation, "Request body must be valid JSON", ""))
		return
	}
	defer r.Body.Close()

	log.Printf("[payments_transfer_xc] --- Incoming request ---")
	log.Printf("[payments_transfer_xc] Method: %s URL: %s", r.Method, r.URL.String())
	log.Printf("[payments_transfer_xc] Headers: X-Client-Id=%s X-Api-Version=%s x-business-message-id=%s",
		r.Header.Get("X-Client-Id"),
		r.Header.Get("X-Api-Version"),
		r.Header.Get("x-business-message-id"))
	bodyBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[payments_transfer_xc] Body:\n%s", string(bodyBytes))
	log.Printf("[payments_transfer_xc] ------------------------")

	businessMessageId := strings.TrimSpace(req.AppHeader.BusinessMessageId)
	if businessMessageId == "" {
		setPayNetResponseHeaders(w, r, "")
		writeTransferResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "appHeader.businessMessageId is required", "")
		return
	}

	creditorBic := strings.TrimSpace(req.CreditorAgent.Id)
	responseBizMsgId := responseBusinessMessageId(businessMessageId, creditorBic)

	if strings.TrimSpace(req.CreditorAccount.Id) == "" {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeTransferResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "creditorAccount.id is required", "")
		return
	}

	amount := strings.TrimSpace(req.InstructedAmount.Amount)
	currency := strings.TrimSpace(req.InstructedAmount.Currency)
	if amount == "" {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeTransferResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "instructedAmount.amount is required", "")
		return
	}
	if currency == "" {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeTransferResponse(w, http.StatusOK, req, TransactionStatusRJCT, ReasonCodeMissingField, ReasonCodeNameValidation, "instructedAmount.currency is required", "")
		return
	}

	// Acquirer business logic: execute payment (e.g. debit debtor, credit creditor).
	// This example uses a stub; replace with real payment processing.
	accepted, creditorName := processPayment(req)
	if accepted {
		setPayNetResponseHeaders(w, r, responseBizMsgId)
		writeTransferResponse(w, http.StatusOK, req, TransactionStatusACSP, ReasonCodeAccepted, ReasonCodeNameAccepted, ReasonDescriptionAccepted, creditorName)
		return
	}
	setPayNetResponseHeaders(w, r, responseBizMsgId)
	writeTransferResponse(w, http.StatusOK, req, TransactionStatusACSP, ReasonCodeNameValidation, ReasonCodeNameValidation, "Payment not accepted", "")
}

// processPayment performs the payment transfer (acquirer side). Stub: accept known test values; returns accepted, creditorName.
func processPayment(req TransferRequest) (bool, string) {
	creditorAccountId := strings.TrimSpace(req.CreditorAccount.Id)
	creditorAgentId := strings.TrimSpace(req.CreditorAgent.Id)
	switch creditorAccountId {
	case "123456789", "22345678901":
		return true, "Jane Smith"
	}
	if creditorAgentId == "MBBEMYKL" && creditorAccountId != "" {
		return true, "Jane Smith"
	}
	return false, ""
}

func writeTransferResponse(w http.ResponseWriter, statusCode int, req TransferRequest, txnStatus, reasonCode, reasonName, reasonDesc, creditorName string) {
	resp := buildTransferResponse(req, txnStatus, reasonCode, reasonName, reasonDesc, creditorName)

	bodyBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[payments_transfer_xc] marshal response for JWS: %v", err)
		writeJSON(w, statusCode, resp)
		return
	}
	privateKey, err := jws_generation.LoadDefaultPrivateKey()
	if err != nil {
		log.Printf("[payments_transfer_xc] load private key for JWS: %v", err)
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
		log.Printf("[payments_transfer_xc] generate JWS: %v", err)
		writeJSON(w, statusCode, resp)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	writeJSON(w, statusCode, resp)
}

// buildTransferResponse builds the response per PayNet API spec (appHeader, data, resp).
// Ref: https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-payments-transfer-xc/post#response-body
func buildTransferResponse(req TransferRequest, txnStatus, reasonCode, reasonName, reasonDescription, creditorName string) TransferResponse {
	origBizMsgId := req.AppHeader.BusinessMessageId
	creditorBic := strings.TrimSpace(req.CreditorAgent.Id)
	responseBizMsgId := responseBusinessMessageId(origBizMsgId, creditorBic)
	// transactionId per spec: same as originalBusinessMessageId (request's businessMessageId from RPP).
	transactionId := req.AppHeader.EndToEndId
	interbankSettlementDate := time.Now().Format("2006-01-02")
	return TransferResponse{
		AppHeader: ResponseAppHeader{
			EndToEndId:                req.AppHeader.EndToEndId,
			BusinessMessageId:         responseBizMsgId,
			CreationDateTime:          req.AppHeader.CreationDateTime,
			OriginalBusinessMessageId: origBizMsgId,
			// TransactionId:             transactionId,
			TransactionId: transactionId,
		},
		Data: ResponseData{
			QR: ResponseQR{
				Category:              CategoryPointOfSales,
				AcceptedSourceOfFunds: AcceptedSourceOfFundsDefault,
			},
			SettlementCycleNumber:   "001",
			InterbankSettlementDate: interbankSettlementDate,
			// Creditor:               ResponseCreditor{Name: creditorName},
			Creditor: ResponseCreditor{Name: "Jane Smith"},
			CreditorAccount: ResponseCreditorAccount{
				Id:                req.CreditorAccount.Id,
				Type:              orDefault(req.CreditorAccount.Type, "WALLET"),
				ResidentStatus:    "RESIDENT",
				ProductType:       "ISLAMIC",
				ShariaCompliance:  "YES",
				AccountHolderType: "SINGLE",
				CustomerCategory:  "RET",
			},
		},
		Resp: ResponseStatus{
			// Status: txnStatus,
			Status: TransactionStatusACSP,
			Reason: ResponseReason{
				Name: reasonName,
				// Code:        reasonCode,
				Code:        "00",
				Description: reasonDescription,
			},
		},
	}
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return def
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

func responseBusinessMessageId(requestId string, creditorBic string) string {
	if len(requestId) <= originatorCodePosition {
		return requestId
	}
	b := []byte(requestId)
	creditorBic = strings.TrimSpace(creditorBic)
	if creditorBic != "" && len(b) >= bicStartPosition+bicLength {
		bic := creditorBic
		if len(bic) > bicLength {
			bic = bic[:bicLength]
		} else if len(bic) < bicLength {
			bic = bic + strings.Repeat(" ", bicLength-len(bic))
		}
		copy(b[bicStartPosition:bicStartPosition+bicLength], bic)
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
	log.Printf("[payments_transfer_xc] --- Outgoing response ---")
	log.Printf("[payments_transfer_xc] HTTP %d", statusCode)
	log.Printf("[payments_transfer_xc] Authorization (outgoing token): %s", w.Header().Get("Authorization"))
	log.Printf("[payments_transfer_xc] Response Headers:")
	for k, v := range w.Header() {
		log.Printf("[payments_transfer_xc]   %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("[payments_transfer_xc] Body:\n%s", string(bodyBytes))
	log.Printf("[payments_transfer_xc] -------------------------")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
