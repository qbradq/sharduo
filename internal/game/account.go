package game

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
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
	util.BaseSerializeable
	// Username
	username string
	// Password hash
	passwordHash string
	// Serial of the player's permanent mobile (not the currently controlled
	// mobile)
	player uo.Serial
}

// ObjectType implements the Object interface.
func (a *Account) ObjectType() marshal.ObjectType { return marshal.ObjectTypeAccount }

// Marshal implements the marshal.Marshaler interface.
func (a *Account) Marshal(s *marshal.TagFileSegment) {
	// Common header
	s.PutInt(uint32(a.Serial()))
	s.PutInt(uint32(a.player))
	s.PutString(a.username)
	s.PutString(a.passwordHash)
	// We don't use the TagObject format here because there is a lot of overhead
	// for non-game objects. Just use a TagCollection.
	s.PutTag(marshal.TagEndOfList, marshal.TagValueBool, true)
}

// Read unmarshals from a segment
func (a *Account) Read(s *marshal.TagFileSegment) {
	a.SetSerial(uo.Serial(s.Int()))
	a.player = uo.Serial(s.Int())
	a.username = s.String()
	a.passwordHash = s.String()
	// An empty tags collection exists so we can add stuff later if needed
	s.Tags()
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (a *Account) Unmarshal(to *marshal.TagObject) {
	// Account deserialization happens inside world.Unmarshal
}

// AfterUnmarshal implements the marshal.Unmarshaler interface.
func (a *Account) AfterUnmarshal(to *marshal.TagObject) {
	// Account deserialization happens inside world.Unmarshal
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
