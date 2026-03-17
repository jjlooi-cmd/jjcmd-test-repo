package jwt_generation

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Default filenames for key and certificate in the same folder as this package.
const (
	DefaultPrivateKeyFilename   = "sample_private_key.key"
	DefaultCertificateFilename = "sample_external_certificate.cer"
)

// defaultPackageDir returns the directory containing this package's source (jwt_generation folder).
func defaultPackageDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

// LoadDefaultPrivateKey loads the private key from sample_private_key.key in this package folder.
func LoadDefaultPrivateKey() (*rsa.PrivateKey, error) {
	path := filepath.Join(defaultPackageDir(), DefaultPrivateKeyFilename)
	return LoadPrivateKey(path)
}

// LoadDefaultCertificate loads the certificate from sample_external_certificate.cer in this package folder.
// Returns the public key and certificate serial number for JWT verification.
func LoadDefaultCertificate() (publicKey *rsa.PublicKey, serialNumber string, err error) {
	path := filepath.Join(defaultPackageDir(), DefaultCertificateFilename)
	return LoadCertificate(path)
}

// LoadPrivateKey loads an RSA private key from a PEM file.
// Supports PKCS#1 ("RSA PRIVATE KEY") and PKCS#8 ("PRIVATE KEY").
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in private key file")
	}
	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKCS8: %w", err)
		}
		pk, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA")
		}
		return pk, nil
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}

// LoadCertificate loads an X.509 certificate from a PEM file and returns the public key and serial number.
// Used for verifying PayNet response JWT.
func LoadCertificate(path string) (publicKey *rsa.PublicKey, serialNumber string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read certificate: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, "", fmt.Errorf("no PEM block in certificate file")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("parse certificate: %w", err)
	}
	pk, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, "", fmt.Errorf("certificate public key is not RSA")
	}
	serialNumber = fmt.Sprintf("%x", cert.SerialNumber)
	return pk, serialNumber, nil
}
