package sec

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha3"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/netip"
	"time"

	"github.com/kadmila/Abyss-Browser/abyss_core/tools/functional"
)

// PrivateKey interface ensures we only use private keys which can
// derive its public key - as all key does (...)
// golang crypto/ guys should be more confident.
type PrivateKey interface {
	Public() crypto.PublicKey
}

func NewRootPrivateKey() (PrivateKey, error) {
	_, privkey, err := ed25519.GenerateKey(rand.Reader)
	return privkey, err
}

// AbyssRootSecret is the root identity of a user.
// It implements ani.IAbyssPeerIdentity, along with administrator features.
type AbyssRootSecret struct {
	root_priv_key PrivateKey

	id                  string
	root_self_cert      string //pem
	root_self_cert_der  []byte
	root_self_cert_x509 *x509.Certificate

	handshake_priv_key *rsa.PrivateKey //may support others in future
	issue_time         time.Time       //handshake encryption key issue time

	handshake_key_cert     string //pem
	handshake_key_cert_der []byte
}

func NewAbyssRootSecrets(root_private_key PrivateKey) (*AbyssRootSecret, error) {
	root_public_key := root_private_key.Public()

	//root certificate
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	id, err := abyssIDFromKey(root_public_key)
	if err != nil {
		return nil, err
	}
	r_template := x509.Certificate{
		Issuer: pkix.Name{
			CommonName: id,
		},
		Subject: pkix.Name{
			CommonName: id,
		},
		NotBefore:             time.Now().Add(time.Duration(-1) * time.Second), //1-sec backdate, for badly synced peers.
		SerialNumber:          serialNumber,
		KeyUsage:              x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	r_derBytes, err := x509.CreateCertificate(rand.Reader, &r_template, &r_template, root_public_key, root_private_key)
	if err != nil {
		return nil, err
	}
	r_x509, err := x509.ParseCertificate(r_derBytes)
	if err != nil {
		return nil, err
	}
	var r_pem_buf bytes.Buffer
	err = pem.Encode(&r_pem_buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: r_derBytes,
	})
	if err != nil {
		return nil, err
	}

	//handshake key
	handshake_private_key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	issue_time := time.Now().Add(time.Duration(-1) * time.Second) //1-sec backdate, for badly synced peers.
	h_template := x509.Certificate{
		Issuer: pkix.Name{
			CommonName: id,
		},
		Subject: pkix.Name{
			CommonName: "h." + id,
		},
		NotBefore:             issue_time,
		SerialNumber:          serialNumber,
		KeyUsage:              x509.KeyUsageEncipherOnly,
		BasicConstraintsValid: true,
	}
	h_derBytes, err := x509.CreateCertificate(rand.Reader, &h_template, &r_template, &handshake_private_key.PublicKey, root_private_key)
	if err != nil {
		return nil, err
	}
	var h_pem_buf bytes.Buffer
	err = pem.Encode(&h_pem_buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: h_derBytes,
	})
	if err != nil {
		return nil, err
	}

	// { // debug
	// 	err = r_x509.CheckSignatureFrom(r_x509)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	h_cert, err := x509.ParseCertificate(h_derBytes)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	err = h_cert.CheckSignatureFrom(r_x509)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return &AbyssRootSecret{
		root_priv_key: root_private_key,

		id:                  id,
		root_self_cert:      r_pem_buf.String(),
		root_self_cert_der:  r_derBytes,
		root_self_cert_x509: r_x509,

		handshake_priv_key: handshake_private_key,
		issue_time:         issue_time,

		handshake_key_cert:     h_pem_buf.String(),
		handshake_key_cert_der: h_derBytes,
	}, nil
}

func (r *AbyssRootSecret) DecryptHandshake(encrypted_payload, encrypted_aes_secret []byte) ([]byte, error) {
	// decrypt AES-GCM secret
	aes_secret, err := rsa.DecryptOAEP(sha3.New256(), nil, r.handshake_priv_key, encrypted_aes_secret, nil)
	if err != nil {
		return nil, err
	}
	aes_key := aes_secret[:32]
	aes_nonce := aes_secret[32:]
	// construct AES-GCM decryptor
	block, err := aes.NewCipher(aes_key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// decrypt payload
	return aesGCM.Open(nil, aes_nonce, encrypted_payload, nil)
}

func (r *AbyssRootSecret) NewTLSIdentity() (*TLSIdentity, error) {
	tls_public_key, tls_private_key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	// memo: this is a problem with golang' crypto library.
	// We can't use CheckSignatureFrom() function to verify non-CA self-signed certficate.
	// In future, this TLS certificate should be
	// IsCA: false, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	tls_self_template := x509.Certificate{
		// no name
		NotBefore:             time.Now().Add(time.Duration(-1) * time.Second), //1-sec backdate, for badly synced peers.
		NotAfter:              time.Now().Add(7 * 24 * time.Hour),              // Valid for 7 days
		SerialNumber:          serialNumber,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	tls_self_derBytes, err := x509.CreateCertificate(rand.Reader, &tls_self_template, &tls_self_template, tls_public_key, tls_private_key)
	if err != nil {
		return nil, err
	}

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	bind_template := x509.Certificate{
		Issuer: pkix.Name{
			CommonName: r.id,
		},
		Subject: pkix.Name{
			CommonName: "tls." + r.id,
		},
		NotBefore:             time.Now().Add(time.Duration(-1) * time.Second), //1-sec backdate, for badly synced peers.
		NotAfter:              time.Now().Add(7 * 24 * time.Hour),              // Valid for 7 days
		SerialNumber:          serialNumber,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  false,
		BasicConstraintsValid: true,
	}
	bind_derBytes, err := x509.CreateCertificate(rand.Reader, &bind_template, r.root_self_cert_x509, tls_public_key, r.root_priv_key)
	if err != nil {
		return nil, err
	}

	// { // debug
	// 	tls_cert, err := x509.ParseCertificate(tls_self_derBytes)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	err = tls_cert.CheckSignatureFrom(tls_cert)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	bind_cert, err := x509.ParseCertificate(bind_derBytes)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	err = bind_cert.CheckSignatureFrom(r.root_self_cert_x509)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return &TLSIdentity{
		priv_key:        tls_private_key,
		tls_self_cert:   tls_self_derBytes,
		abyss_bind_cert: bind_derBytes,
	}, nil
}

// NewAddressCertificate creates a short-lived certificate (< 1 hour) containing
// the peer's address candidates. The certificate is signed by the peer's root key
// and has CN = "loc.{peer_id}".
//
// The addresses are encoded in the certificate with SAN extension
func (r *AbyssRootSecret) NewAddressCertificate(addresses []netip.AddrPort) (*x509.Certificate, error) {
	// Generate a temporary key pair for the address certificate
	dummy_public_key, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	// Convert addresses to standard SAN format
	// Using RFC 3986: DNS names can contain IP addresses
	// Format as IP:port for both IPv4 and IPv6 for consistency
	address_str := functional.Filter(addresses, func(addr netip.AddrPort) string {
		return addr.String()
	})

	now := time.Now()
	loc_template := x509.Certificate{
		Issuer: pkix.Name{
			CommonName: r.id,
		},
		Subject: pkix.Name{
			CommonName: "loc." + r.id,
		},
		NotBefore:             now.Add(-1 * time.Second), // 1-sec backdate for clock skew
		NotAfter:              now.Add(50 * time.Minute), // < 1 hour validity
		SerialNumber:          serialNumber,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		IsCA:                  false,
		BasicConstraintsValid: true,
		DNSNames:              address_str,
	}

	loc_derBytes, err := x509.CreateCertificate(rand.Reader, &loc_template, r.root_self_cert_x509, dummy_public_key, r.root_priv_key)
	if err != nil {
		return nil, err
	}

	loc_cert, err := x509.ParseCertificate(loc_derBytes)
	if err != nil {
		return nil, err
	}

	// // debug: Verify signature (sanity check)
	// err = loc_cert.CheckSignatureFrom(r.root_self_cert_x509)
	// if err != nil {
	// 	return nil, err
	// }

	return loc_cert, nil
}

func (r *AbyssRootSecret) ID() string                         { return r.id }
func (r *AbyssRootSecret) RootCertificate() string            { return r.root_self_cert }
func (r *AbyssRootSecret) RootCertificateDer() []byte         { return r.root_self_cert_der }
func (r *AbyssRootSecret) HandshakeKeyCertificate() string    { return r.handshake_key_cert }
func (r *AbyssRootSecret) HandshakeKeyCertificateDer() []byte { return r.handshake_key_cert_der }
func (r *AbyssRootSecret) IssueTime() time.Time               { return r.issue_time }
