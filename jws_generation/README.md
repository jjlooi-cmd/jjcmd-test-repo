# JWS Generation (PayNet DuitNow)

This package implements **JSON Web Signature (JWS)** for PayNet DuitNow Transfer API authentication, as per [PayNet’s JWS documentation](https://docs.developer.paynet.my/docs/duitnow-transfer/integration/security-&-encryption/message-signature/JSON-web-signature).

## Overview

- **Generate JWS** for outbound API requests and set `Authorization: Bearer <token>`.
- **Verify JWS** on API responses using PayNet’s public key (certificate).

Supported algorithms: **RS256**, **RS512**.

## Usage

### 1. Generate JWS for an API request

```go
import "example.com/sample-repo/jws_generation"

// Load private key: use default (sample_private_key.key in this package folder) or a custom path.
privateKey, err := jws_generation.LoadDefaultPrivateKey()
// Or: privateKey, err := jws_generation.LoadPrivateKey("path/to/private.key")
if err != nil {
    return err
}

// Request body as minified JSON (same bytes you will send in the request).
payloadForHash, _ := json.Marshal(yourRequestBody)

token, err := jws_generation.GenerateJWS(jws_generation.GenerateOptions{
    PrivateKey:        privateKey,
    Algorithm:         jws_generation.RS256,
    Issuer:            "BOEEMYK1",              // Your BIC / issuer
    BusinessMessageID: "20230412BOEEMYK1000ORB00000001",
    CredentialKey:     "ERTqafGRyt35MAKX5pBMU", // JWT Credential Key from onboarding
    PayloadForHash:    payloadForHash,
})
if err != nil {
    return err
}

// Set header when calling PayNet API.
req.Header.Set("Authorization", "Bearer "+token)
```

- **PayloadForHash**: Use the **exact** request body bytes (minified JSON). If the request has no body, you can pass `nil` (treated as empty string and hashed accordingly).

### 2. Verify JWS on an API response

```go
// Load certificate: use default (sample_external_certificate.cer in this package folder) or a custom path.
publicKey, _, err := jws_generation.LoadDefaultCertificate()
// Or: publicKey, _, err := jws_generation.LoadCertificate("path/to/paynet-cert.cer")
if err != nil {
    return err
}

token := r.Header.Get("Authorization") // or from response header
responseBody := []byte(`{"data":{...}}`) // minified response body

err = jws_generation.VerifyJWS(jws_generation.VerifyOptions{
    Token:          token,
    PublicKey:      publicKey,
    Algorithm:      jws_generation.RS256,
    PayloadForHash: responseBody,
})
if err != nil {
    return err // invalid or expired token / payload mismatch
}
```

### 3. Key and certificate files

Place your key and certificate in the `jws_generation` folder with these names to use the default loaders:

- **sample_private_key.key** – PEM, either `RSA PRIVATE KEY` (PKCS#1) or `PRIVATE KEY` (PKCS#8).
- **sample_external_certificate.cer** – PEM with PayNet’s X.509 certificate (for verification).

For custom paths, use `LoadPrivateKey(path)` and `LoadCertificate(path)` instead.

To generate a key pair (e.g. for sandbox):

```bash
# Private key + CSR
openssl req -newkey rsa:2048 -nodes -keyout example.key -out example.csr

# Self-signed cert (sandbox only; use a proper CA for production)
openssl x509 -signkey example.key -in example.csr -req -days 365 -out example.cer
```

## Specification summary

| Step | Description |
|------|-------------|
| 1 | Compute **ds** = SHA-256 of the minified request/response payload. |
| 2 | JWS **header**: `{"alg":"RS256","typ":"JWT"}` (or RS512). |
| 3 | JWS **body**: `iss`, `iat`, `exp`, `key`, `jti`, `ds`. |
| 4 | Sign `base64url(header).base64url(body)` with the private key. |
| 5 | Token = `header.payload.signature`; send as `Authorization: Bearer <token>`. |

Expiry (`exp`) is set to 15 minutes from issue by default; it must be within 1 hour of `iat` per PayNet.
