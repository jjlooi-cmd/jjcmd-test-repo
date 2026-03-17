// Package common_header implements DuitNow Pay common request headers per PayNet docs.
// Ref: https://docs.developer.paynet.my/docs/duitnow-pay/integration/construct-header
package common_header

import (
	"net/http"
	"strings"
)

// Header names required by DuitNow Pay Construct Header.
const (
	HeaderXIPAddress      = "X-Ip-Address"
	HeaderXGPSCoordinates = "X-Gps-Coordinates"
)

// Values holds the common header values for DuitNow Pay API requests.
// Every DuitNow Pay API endpoint (besides Authorization) shall include these headers.
type Values struct {
	// IPAddress is IPv4 or IPv6 where the transaction is triggered from.
	// e.g. "10.236.166.145" or "2001:0db8:85a3:0000:0000:8a2e:0370:7334".
	// If unable to obtain, use "0".
	IPAddress string
	// GPSCoordinates is location in decimal degree format, e.g. "40.689263, 74.044505".
	// If unable to obtain, use "0".
	GPSCoordinates string
}

// Default returns Values with "0" for both fields (allowed by PayNet when unable to obtain).
func Default() Values {
	return Values{
		IPAddress:      "10.236.166.145",
		GPSCoordinates: "40.689263, 74.044505",
	}
}

// ApplyToRequest sets the DuitNow Pay common headers on req.
// Empty IPAddress or GPSCoordinates are treated as "0" per PayNet guidance.
func ApplyToRequest(req *http.Request, v Values) {
	ip := strings.TrimSpace(v.IPAddress)
	if ip == "" {
		ip = "0"
	}
	gps := strings.TrimSpace(v.GPSCoordinates)
	if gps == "" {
		gps = "0"
	}
	req.Header.Set(HeaderXIPAddress, ip)
	req.Header.Set(HeaderXGPSCoordinates, gps)
}
