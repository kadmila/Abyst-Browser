package sec

import (
	"crypto/x509"
	"encoding/pem"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewWorldSessionCertificate(t *testing.T) {
	// Create a root secret
	rootKey, err := NewRootPrivateKey()
	if err != nil {
		t.Fatalf("Failed to create root private key: %v", err)
	}

	rootSecret, err := NewAbyssRootSecrets(rootKey)
	if err != nil {
		t.Fatalf("Failed to create root secret: %v", err)
	}

	// Create test parameters
	worldID := uuid.New()
	sessionID := uuid.New()
	envURL, err := url.Parse("https://example.com/world/env")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	// Generate the certificate
	certPEM, err := rootSecret.NewWorldSessionCertificate(worldID, sessionID, envURL)
	if err != nil {
		t.Fatalf("Failed to create world session certificate: %v", err)
	}

	// Verify it's PEM format
	if !strings.HasPrefix(certPEM, "-----BEGIN CERTIFICATE-----") {
		t.Errorf("Certificate is not in PEM format")
	}

	// Parse the PEM
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		t.Fatalf("Failed to decode PEM block")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify subject CN
	expectedCN := worldID.String() + "." + sessionID.String() + ".world." + rootSecret.ID()
	if cert.Subject.CommonName != expectedCN {
		t.Errorf("Subject CN = %s, want %s", cert.Subject.CommonName, expectedCN)
	}

	// Verify issuer CN
	if cert.Issuer.CommonName != rootSecret.ID() {
		t.Errorf("Issuer CN = %s, want %s", cert.Issuer.CommonName, rootSecret.ID())
	}

	// Verify URI
	if len(cert.URIs) != 1 {
		t.Errorf("Expected 1 URI, got %d", len(cert.URIs))
	} else if cert.URIs[0].String() != envURL.String() {
		t.Errorf("URI = %s, want %s", cert.URIs[0].String(), envURL.String())
	}

	// Verify certificate properties
	if cert.IsCA {
		t.Errorf("Certificate should not be a CA")
	}

	if cert.KeyUsage != x509.KeyUsageDigitalSignature {
		t.Errorf("KeyUsage = %v, want %v", cert.KeyUsage, x509.KeyUsageDigitalSignature)
	}

	t.Logf("✓ World session certificate created successfully")
	t.Logf("  Subject CN: %s", cert.Subject.CommonName)
	t.Logf("  Issuer CN: %s", cert.Issuer.CommonName)
	t.Logf("  URI: %s", cert.URIs[0].String())
	t.Logf("  Valid: %v - %v", cert.NotBefore, cert.NotAfter)
}
