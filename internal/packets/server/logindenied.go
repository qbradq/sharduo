package server

// LoginDeniedReason represents the reason for issuing a LoginDenied packet
type LoginDeniedReason byte

// All meaningfull LoginDeniedReason values
const (
	LoginDeniedReasonBadPass        LoginDeniedReason = iota // Password invalid for user
	LoginDeniedReasonAccountInUse                            // The account already has an active season
	LoginDeniedReasonAccountBlocked                          // The account has been blocked for some reason
)

// A LoginDenied packet informs the client that an account login attempt has
// failed and provides a
type LoginDenied struct {
	Reason LoginDeniedReason
}

// Compile encodes the state of the Packet object using w
func (p *LoginDenied) Compile(w *PacketWriter) {
	w.PutByte(0x82)
	w.PutByte(byte(p.Reason))
}
