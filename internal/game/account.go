package game

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

// Hashes a password suitable for the accounts database.
func HashPassword(password string) string {
	hd := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hd[:])
}

// NewAccount creates a new account object
func NewAccount(username, passwordHash string) *Account {
	return &Account{
		username:     username,
		passwordHash: passwordHash,
	}
}

// Account holds all of the account information for one user
type Account struct {
	// Username
	username string
	// Password hash
	passwordHash string
	// Serial of the player's permanent mobile (not the currently controlled
	// mobile)
	player uo.Serial
}

// Marshal writes the account data to a segment
func (a *Account) Marshal(s *marshal.TagFileSegment) {
	s.PutInt(uint32(a.player))
	s.PutString(a.username)
	s.PutString(a.passwordHash)
	// Slap a tag collection terminator on here so we can safely expand later if
	// we need to.
	s.PutTag(marshal.TagEndOfList, marshal.TagValueBool, true)
}

// Unmarshal reads the account data from a segment
func (a *Account) Unmarshal(s *marshal.TagFileSegment) {
	a.player = uo.Serial(s.Int())
	a.username = s.String()
	a.passwordHash = s.String()
	// An empty tags collection exists so we can add stuff later if needed
	s.Tags()
}

// Username returns the username of the account
func (a *Account) Username() string {
	return a.username
}

// ComparePasswordHash returns true if the hash given matches this account's
func (a *Account) ComparePasswordHash(hash string) bool {
	return a.passwordHash == hash
}

// Player returns the player mobile serial, or uo.SerialMobileNil if none
func (a *Account) Player() uo.Serial { return a.player }

// SetPlayer sets the player mobile serial, or uo.SerialMobileNil if none
func (a *Account) SetPlayer(s uo.Serial) { a.player = s }
