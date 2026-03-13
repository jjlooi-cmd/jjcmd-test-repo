# payments_transfer_xc

Implementation of the **PayNet webhook** `POST /webhooks/v3/payments/transfer-xc` for **Merchant Presented QR (MPM) – Domestic Acquirer**.

## API reference

- **Spec**: [Merchant Presented QR: Domestic - Acquirer – webhooks/v3/payments/transfer-xc (POST)](https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-payments-transfer-xc/post)
- **Product**: QR MPM Domestic (Acquirer)
- **Purpose**: Payment transfer – execute the actual payment (debit debtor, credit creditor) after account enquiry.

## Behaviour

- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request**: See `TransferRequest` in `models.go` (appHeader, debtor, debtorAccount, debtorAgent, creditorAgent, creditorAccount, qr, instructedAmount).
- **Response**: JSON with appHeader (including originalBusinessMessageId), resp.status (`ACSP` | `RJCT`), resp.reason (name, code, description). Response is signed with JWS in `Authorization: Bearer <token>`.

Validation:

1. **Message validation**  
   Missing or invalid required fields (e.g. businessMessageId, creditorAccount.id, instructedAmount) → `resp.status: RJCT`, HTTP 200.

2. **Business logic**  
   Implemented in `processPayment()` in `handler.go`. Current logic is a **stub**: it returns ACSP for known test creditor accounts (e.g. 123456789, 22345678901, or any account when creditorAgent is MBBEMYKL); otherwise RJCT. Replace with your real payment execution (e.g. core banking debit/credit).

## Usage

Register the handler in your HTTP server:

```go
import "example.com/sample-repo/qr_acquirer/payments_transfer_xc"

http.HandleFunc("/webhooks/v3/payments/transfer-xc", payments_transfer_xc.Handler)
```

## Request / response examples

**Request (from PayNet sample):**

Headers: `X-Client-Id`, `X-Api-Version`, `x-business-message-id`, `Authorization: Bearer <JWT>`.

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
  "qr": { "code": "00020201021126410014A000000615000101065887340209123456789520460105303458540115802MY5909QRCSDNBHD6005BANGI6304DAF5" },
  "instructedAmount": { "amount": "10.00", "currency": "MYR" }
}
```

**Response (ACSP – accepted):**

```json
{
  "appHeader": {
    "endToEndId": "...",
    "businessMessageId": "...",
    "creationDateTime": "...",
    "originalBusinessMessageId": "..."
  },
  "resp": {
    "status": "ACSP",
    "reason": {
      "name": "ACCEPTED",
      "code": "45",
      "description": "Success/ Transaction Accepted"
    }
  }
}
```

**Response (RJCT – rejected):**

```json
{
  "appHeader": { ... },
  "resp": {
    "status": "RJCT",
    "reason": {
      "name": "MESSAGE_VALIDATION_ERROR",
      "code": "45",
      "description": "creditorAccount.id is required"
    }
  }
}
```

For the exact schema (field names, lengths, codes), use the official [PayNet API Reference](https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-payments-transfer-xc/post).
