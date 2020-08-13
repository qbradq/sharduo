package uod

import (
	"github.com/qbradq/sharduo/pkg/uo"

	"github.com/qbradq/sharduo/internal/common"
)

type netState struct {
	outboundPackets chan []byte
	compress        bool
	account         *common.Account
}

func newNetState() *netState {
	return &netState{
		outboundPackets: make(chan []byte, 100),
	}
}

// SetAccount sets the associated account
func (n *netState) SetAccount(a *common.Account) {
	n.account = a
}

// Authenticated returns true if the client has already been authenticated with an 0x91 packet
func (n *netState) Authenticated() bool {
	// common.Role.HasAll() is read-only on a single value
	return n.account != nil && n.account.Roles().HasAll(common.RoleAuthenticated)
}

// CompressOutput returns true if packets sent to this client should be compressed
func (n *netState) CompressOutput() bool {
	// bool read is atomic
	return n.compress
}

// BeginCompression makes all new output packets compressed
func (n *netState) BeginCompression() {
	// bool set is atomic
	n.compress = true
}

// SendPacket sends a packet object to the client over the network. Returns
// false if the client's output packet channel is full or if the input packet
// object cannot be cast to a byte slice.
func (n *netState) SendPacket(p uo.ServerPacket) bool {
	// Channel write is thread-safe
	select {
	case n.outboundPackets <- p.Bytes():
		return true
	default:
		return false
	}
}
