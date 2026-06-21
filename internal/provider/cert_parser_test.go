package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestParsePEMCertificate(t *testing.T) {
	// 1. Generate mock P-256 ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	serialNumber := big.NewInt(1234567890)
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		Issuer: pkix.Name{
			CommonName: "Mock Authority",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 2. Generate self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// 3. Encode to PEM block
	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	pemString := string(pemBlock)

	// 4. Test the parser
	metadata, err := ParsePEMCertificate(pemString)
	if err != nil {
		t.Fatalf("ParsePEMCertificate failed: %v", err)
	}

	// 5. Assertions
	if metadata.SerialNumber != "499602D2" { // 1234567890 in Hex is 499602D2
		t.Errorf("Expected SerialNumber '499602D2', got %q", metadata.SerialNumber)
	}

	if metadata.IssuerCommonName != "test.example.com" {
		t.Errorf("Expected IssuerCommonName 'test.example.com', got %q", metadata.IssuerCommonName)
	}

	if metadata.KeyAlgorithm != "ECDSA" {
		t.Errorf("Expected KeyAlgorithm 'ECDSA', got %q", metadata.KeyAlgorithm)
	}

	if metadata.KeySize != 256 {
		t.Errorf("Expected KeySize 256, got %d", metadata.KeySize)
	}

	if !strings.Contains(metadata.SignatureAlgorithm, "ECDSA") {
		t.Errorf("Expected SignatureAlgorithm to contain 'ECDSA', got %q", metadata.SignatureAlgorithm)
	}

	expectedIssued := notBefore.UTC().Format(time.RFC3339)
	if metadata.IssuedAt != expectedIssued {
		t.Errorf("Expected IssuedAt %q, got %q", expectedIssued, metadata.IssuedAt)
	}

	expectedExpires := notAfter.UTC().Format(time.RFC3339)
	if metadata.ExpiresAt != expectedExpires {
		t.Errorf("Expected ExpiresAt %q, got %q", expectedExpires, metadata.ExpiresAt)
	}

	if !metadata.IsValid {
		t.Error("Expected IsValid to be true")
	}

	if metadata.DaysRemaining != 0 { // ~23 hours remaining, which is 0 days when dividing hours/24
		t.Errorf("Expected DaysRemaining 0, got %d", metadata.DaysRemaining)
	}
}

func TestParsePEMCertificate_Invalid(t *testing.T) {
	_, err := ParsePEMCertificate("invalid-pem-data")
	if err == nil {
		t.Error("Expected error for invalid PEM data, got nil")
	}
}
