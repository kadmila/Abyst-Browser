package ann

import "github.com/quic-go/quic-go"

const (
	AbyssQuicRedundantConnection quic.ApplicationErrorCode = 1000
	AbyssQuicAhmpStreamFail      quic.ApplicationErrorCode = 1001
	AbyssQuicCryptoFail          quic.ApplicationErrorCode = 1002
	AbyssQuicAuthenticationFail  quic.ApplicationErrorCode = 1002
)

// TODO: custom error wrapper - to let users know
// can they ignore a failed Accept or not.
// type AbyssNetworkError struct{}
// type AbyssCryptoError struct{}
