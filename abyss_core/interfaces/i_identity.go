package interfaces

type IHostIdentity interface {
	IDHash() string
	RootCertificate() string         //pem
	HandshakeKeyCertificate() string //pem
}
