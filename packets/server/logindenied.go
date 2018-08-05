package server

// LoginDeniedReason represents the reason for issuing a LoginDenied packet
type LoginDeniedReason byte

const (
	// LoginDeniedReasonBadPass indicates the password is invalid for the username
	LoginDeniedReasonBadPass LoginDeniedReason = iota
	// LoginDeniedReasonAccountInUse indicates the account being logged into is in active use
	LoginDeniedReasonAccountInUse
	// LoginDeniedReasonAccountBlocked indicates the account has been blocked for any reason
	LoginDeniedReasonAccountBlocked
	// LoginDeniedReasonBadCredentials do not use
	LoginDeniedReasonBadCredentials
	// LoginDeniedReasonOriginCommunication do not use
	LoginDeniedReasonOriginCommunication
	// LoginDeniedReasonIGRConcurrency do not use
	LoginDeniedReasonIGRConcurrency
	// LoginDeniedReasonIGRLimit do not use
	LoginDeniedReasonIGRLimit
	// LoginDeniedReasonIGRFailure do not use
	LoginDeniedReasonIGRFailure
)

// A LoginDenied packet informs the client that an account login attempt has
// failed and provides a
type LoginDenied struct {
	Reason LoginDeniedReason
}

// Compile encodes the state of the Packet object into a byte slice provided
// by the caller. The returned slice will be a slice of the input slice.
func (p *LoginDenied) Compile(buf []byte) []byte {
	buf[0] = 0x82
	buf[1] = byte(p.Reason)
	return buf[0:2]
}
