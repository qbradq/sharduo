package world

import "github.com/qbradq/sharduo/internal/packets/server"

// A NewCharacterRequest asks the root instance to create a new mobile for an
// account.
type NewCharacterRequest struct {
	State *server.NetState
	Name  string
}
