package server

import (
	"github.com/qbradq/sharduo/internal/accounting"
	"github.com/qbradq/sharduo/internal/common"
)

// A NetState object represents the state of a client's connection with the server.
// All exported functions are thread-safe.
type NetState struct {
	ps       PacketSender
	compress bool
	account  *common.Account
}

// NewNetState creates a new NetState
func NewNetState(ps PacketSender) *NetState {
	return &NetState{
		ps: ps,
	}
}

// SetAccount sets the associated account
func (n *NetState) SetAccount(a *accounting.Account) {
	n.account = a
}

// Authenticated returns true if the client has already been authenticated with an 0x91 packet
func (n *NetState) Authenticated() bool {
	// common.Role.HasAll() is read-only on a single value
	return n.account != nil && n.account.Roles().HasAll(common.RoleAuthenticated)
}

// PacketSender returns the PacketSender object that can be used to reply to this client
func (n *NetState) PacketSender() PacketSender {
	// NetState.ps is never written to
	return n.ps
}

// CompressOutput returns true if packets sent to this client should be compressed
func (n *NetState) CompressOutput() bool {
	// bool read is atomic
	return n.compress
}

// BeginCompression makes all new output packets compressed
func (n *NetState) BeginCompression() {
	// bool set is atomic
	n.compress = true
}
