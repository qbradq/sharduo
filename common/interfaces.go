package common

// A Compiler is an object that can encode its state into a byte slice in Ultima
// Online wire protocol format
type Compiler interface {
	Compile(w *PacketWriter)
}

// A PacketSender can transmit Packet objects to consumers gracefully and
// without blocking
type PacketSender interface {
	PacketSend(p Compiler)
}

// A NetState contains the state information for a client's connection to the server
type NetState interface {
	// Authenticated returns true if the client has already been authenticated with an 0x91 packet
	Authenticated() bool
	// Roles returns a Role value representing all roles of the authenticated account
	Roles() Role
	// AddRole adds a Role to the account
	AddRole(r Role)
	// RemoveRole removes a Role from the account
	RemoveRole(r Role)
	// PacketSender returns the PacketSender object that can be used to reply to this client
	PacketSender() PacketSender
	// BeginCompression makes all new output packets compressed
	BeginCompression()
	// CompressOutput returns true if packets sent to this client should be compressed
	CompressOutput() bool
}
