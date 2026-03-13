# payments_reverse

Implementation of the **PayNet webhook** `POST /webhooks/v3/payments/reverse` for **DuitNow Reversal – Issuer**.

## API reference

- **Spec**: [DuitNow Reversal - Issuer – webhooks/v3/payments/reverse (POST)](https://docs.developer.paynet.my/api-reference/v3/reversal/issuer#/webhooks/webhooks-v3-payments-reverse/post)
- **Product**: DuitNow Reversal (Issuer / OFI)
- **Purpose**: Receive reversal requests from RPP and respond with SUCCESSFUL / NEGATIVE / REJECT after validating and processing the reversal (e.g. crediting back the original debtor).

## Behaviour

- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request**: See `ReversalRequest` in `models.go` (appHeader, debtor, debtorAccount, debtorAgent, creditorAgent, creditorAccount, instructedAmount, optional original transaction references).
- **Response**: JSON with `appHeader` (businessMessageId, originalBusinessMessageId) and `resp` (status ACSP/RJCT, reason code/name/description). Response is signed with JWS in `Authorization: Bearer <token>`.

Validation:

1. **Message validation**  
   Missing or invalid required fields (e.g. `appHeader.businessMessageId`, `debtorAccount.id`, `instructedAmount.amount`/`currency`) → `status: RJCT`, HTTP 200.

2. **Business validation**  
   Implemented in `processReversal()` in `handler.go`. Current logic is a **stub**: it accepts when debtor account and amount are present; replace with real reversal processing (lookup original transfer, credit debtor, etc.).

## Usage

Register the handler in your HTTP server:

```go
import "example.com/sample-repo/payments_reverse"

http.HandleFunc("/webhooks/v3/payments/reverse", payments_reverse.Handler)
```

## Request / response

Headers: `X-Client-Id`, `X-Api-Version`, `x-business-message-id`, `Authorization: Bearer <JWT>`.

**Example request (reversal from RPP to Issuer):**

```json
{
  "appHeader": {
    "endToEndId": "20260313PICAMYK1520OQR93208585",
    "businessMessageId": "20260313RPPEMYKL520HQR68305064",
    "creationDateTime": "2026-03-13T10:45:10.413+08:00"
  },
  "debtor": { "id": "****8901", "name": "John Doe" },
  "debtorAccount": {
    "id": "22345678901",
    "type": "CURRENT",
    "residentStatus": "RESIDENT",
    "productType": "ISLAMIC",
    "shariaCompliance": "YES",
    "accountHolderType": "SINGLE"
  },
  "debtorAgent": { "id": "PICAMYK1" },
  "creditorAgent": { "id": "MBBEMYKL" },
  "creditorAccount": { "id": "123456789", "type": "DEFAULT" },
  "instructedAmount": { "amount": "10.00", "currency": "MYR" }
}
```

**Example response (ACSP):**

```json
{
  "appHeader": {
    "endToEndId": "20260313PICAMYK1520OQR93208585",
    "businessMessageId": "20260313PICAMYK1520RQR68305064",
    "creationDateTime": "2026-03-13T10:45:10.413+08:00",
    "originalBusinessMessageId": "20260313RPPEMYKL520HQR68305064"
  },
  "resp": {
    "status": "ACSP",
    "reason": {
      "name": "ACCEPTED",
      "code": "00",
      "description": "Success/ Transaction Accepted"
    }
  }
}
```

For the exact schema (field names, lengths, codes), use the official [PayNet API Reference](https://docs.developer.paynet.my/api-reference/v3/reversal/issuer#/webhooks/webhooks-v3-payments-reverse/post).
