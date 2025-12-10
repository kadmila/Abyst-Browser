package ann

import "github.com/quic-go/quic-go"

const (
	AbyssQuicRedundantConnection quic.ApplicationErrorCode = 0x1000
	AbyssQuicAhmpStreamFail      quic.ApplicationErrorCode = 0x1001
	AbyssQuicCryptoFail          quic.ApplicationErrorCode = 0x1002
	AbyssQuicAuthenticationFail  quic.ApplicationErrorCode = 0x1003

	AbyssQuicClose    quic.ApplicationErrorCode = 0x1100
	AbyssQuicOverride quic.ApplicationErrorCode = 0x1101
)

// TODO: custom error wrapper - to let users know
// can they ignore a failed Accept or not.
// type AbyssNetworkError struct{}
// type AbyssCryptoError struct{}
