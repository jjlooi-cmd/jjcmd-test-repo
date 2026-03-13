# account_enquire_xc

Implementation of the **PayNet webhook** `POST /webhooks/v3/accounts/enquire-xc` for **Merchant Presented QR (MPM) – Domestic Acquirer**.

## API reference

- **Spec**: [Merchant Presented QR: Domestic - Acquirer – webhooks/v3/accounts/enquire-xc (POST)](https://docs.developer.paynet.my/api-reference/v3/QR-MPM/acquirer/domestic#/webhooks/webhooks-v3-accounts-enquire-xc/post)
- **Product**: QR MPM Domestic (Acquirer)
- **Purpose**: Account enquiry – validate that a beneficiary account is valid and ready to receive payment before accepting a payment request.

## Behaviour

- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request**: See `EnquireRequest` in `models.go` (e.g. `messageId`, `proxyId` / `accountNumber`, `bankCode`, `correlationId`).
- **Response**: JSON with `messageId`, `transactionStatus` (`SUCCESSFUL` | `NEGATIVE` | `REJECT`), and when successful `beneficiaryAccountName`; for failures `reasonCode` and `message`.

Validation:

1. **Message validation**  
   Missing or invalid required fields (e.g. `messageId`) → `transactionStatus: REJECT`, HTTP 200.

2. **Business validation**  
   No valid identifier (e.g. no `proxyId` or `accountNumber`) → `transactionStatus: NEGATIVE`, HTTP 200.

3. **Account resolution**  
   Implemented in `resolveAccount()` in `handler.go`. Current logic is a **stub**: it returns `SUCCESSFUL` with a placeholder name for a few test proxy/account values; all other requests return `NEGATIVE` with `ACCOUNT_NOT_FOUND`. Replace with your real account/proxy lookup (e.g. core banking or PayNet proxy resolution).

## Usage

Register the handler in your HTTP server:

```go
import "example.com/sample-repo/account_enquire_xc"

http.HandleFunc("/webhooks/v3/accounts/enquire-xc", account_enquire_xc.Handler)
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
  "qr": { "code": "00020201021126410014A000000615000101065887340209123456789520460105303458540115802MY5909QRCSDNBHD6005BANGI6304DAF5" }
}
```

**Response (SUCCESSFUL):**

```json
{
  "messageId": "msg-001",
  "transactionStatus": "SUCCESSFUL",
  "beneficiaryAccountName": "ACCOUNT HOLDER NAME"
}
```

**Response (NEGATIVE):**

```json
{
  "messageId": "msg-001",
  "transactionStatus": "NEGATIVE",
  "reasonCode": "ACCOUNT_NOT_FOUND",
  "message": "Beneficiary account not found or not eligible"
}
```

For the exact schema (field names, lengths, codes), use the official PayNet API Reference and export the OpenAPI spec if available.
