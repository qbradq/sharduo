package client

import (
	"github.com/qbradq/sharduo/common"
)

// A NetState object represents the state of a client's connection with the server
type NetState struct {
	roles    common.Role
	ps       common.PacketSender
	compress bool
}

// NewNetState creates a new NetState
func NewNetState(ps common.PacketSender) *NetState {
	return &NetState{
		roles: common.RoleNone,
		ps:    ps,
	}
}

// Authenticated returns true if the client has already been authenticated with an 0x91 packet
func (n *NetState) Authenticated() bool {
	return n.roles.HasAll(common.RoleAuthenticated)
}

// Roles returns a Role value representing all roles of the authenticated account
func (n *NetState) Roles() common.Role {
	return n.roles
}

// AddRole adds a Role to the account
func (n *NetState) AddRole(r common.Role) {
	n.roles = n.roles | r
}

// RemoveRole removes a Role from the account
func (n *NetState) RemoveRole(r common.Role) {
	n.roles = n.roles & (^r)
}

// PacketSender returns the PacketSender object that can be used to reply to this client
func (n *NetState) PacketSender() common.PacketSender {
	return n.ps
}

// CompressOutput returns true if packets sent to this client should be compressed
func (n *NetState) CompressOutput() bool {
	return n.compress
}

// BeginCompression makes all new output packets compressed
func (n *NetState) BeginCompression() {
	n.compress = true
}
