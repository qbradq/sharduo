package uod

import "github.com/qbradq/sharduo/internal/game"

// CharacterLogin is sent by the game service to request the world load in the
// player's mobile.
type CharacterLogin struct {
	CharacterIndex int // Index of character selected for login
}

// ID returns the pseudo packet ID.
func (p *CharacterLogin) ID() byte { return 0xFF }

// CharacterLogout is sent by event handlers to request a character logout
// sequence.
type CharacterLogout struct {
	Mobile *game.Mobile // The character mobile we are logging out
}

// ID returns the pseudo packet ID.
func (p *CharacterLogout) ID() byte { return 0xFE }
