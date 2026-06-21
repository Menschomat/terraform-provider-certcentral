package provider

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

type ParsedCertMetadata struct {
	IssuedAt           string
	ExpiresAt          string
	DaysRemaining      int64
	IsValid            bool
	SerialNumber       string
	IssuerCommonName   string
	SignatureAlgorithm string
	KeyAlgorithm       string
	KeySize            int64
}

func ParsePEMCertificate(pemString string) (*ParsedCertMetadata, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode certificate PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 certificate: %w", err)
	}

	now := time.Now()
	daysRemaining := int64(cert.NotAfter.Sub(now).Hours() / 24)
	isValid := now.After(cert.NotBefore) && now.Before(cert.NotAfter)

	var keyAlgo string
	var keySize int64
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		keyAlgo = "RSA"
		keySize = int64(pub.N.BitLen())
	case *ecdsa.PublicKey:
		keyAlgo = "ECDSA"
		keySize = int64(pub.Curve.Params().BitSize)
	default:
		keyAlgo = "Unknown"
	}

	return &ParsedCertMetadata{
		IssuedAt:           cert.NotBefore.Format(time.RFC3339),
		ExpiresAt:          cert.NotAfter.Format(time.RFC3339),
		DaysRemaining:      daysRemaining,
		IsValid:            isValid,
		SerialNumber:       fmt.Sprintf("%X", cert.SerialNumber),
		IssuerCommonName:   cert.Issuer.CommonName,
		SignatureAlgorithm: signatureAlgorithmString(cert.SignatureAlgorithm),
		KeyAlgorithm:       keyAlgo,
		KeySize:            keySize,
	}, nil
}

func signatureAlgorithmString(algo x509.SignatureAlgorithm) string {
	switch algo {
	case x509.MD2WithRSA:
		return "MD2-RSA"
	case x509.MD5WithRSA:
		return "MD5-RSA"
	case x509.SHA1WithRSA:
		return "SHA1-RSA"
	case x509.SHA256WithRSA:
		return "SHA256-RSA"
	case x509.SHA384WithRSA:
		return "SHA384-RSA"
	case x509.SHA512WithRSA:
		return "SHA512-RSA"
	case x509.DSAWithSHA1:
		return "DSA-SHA1"
	case x509.DSAWithSHA256:
		return "DSA-SHA256"
	case x509.ECDSAWithSHA1:
		return "ECDSA-SHA1"
	case x509.ECDSAWithSHA256:
		return "ECDSA-SHA256"
	case x509.ECDSAWithSHA384:
		return "ECDSA-SHA384"
	case x509.ECDSAWithSHA512:
		return "ECDSA-SHA512"
	case x509.SHA256WithRSAPSS:
		return "SHA256-RSAPSS"
	case x509.SHA384WithRSAPSS:
		return "SHA384-RSAPSS"
	case x509.SHA512WithRSAPSS:
		return "SHA512-RSAPSS"
	case x509.PureEd25519:
		return "Ed25519"
	default:
		s := algo.String()
		if s == "" {
			return "Unknown"
		}
		// Clean up common string formats from standard lib if it's fallback
		s = strings.ReplaceAll(s, "ECDSAWith", "ECDSA-")
		s = strings.ReplaceAll(s, "With", "-")
		return s
	}
}
