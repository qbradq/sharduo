package game

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	marshal.RegisterCtor(marshal.ObjectTypeAccount, func() interface{} { return &Account{} })
}

// Role describes the roles that an account may have.
type Role byte

const (
	RolePlayer        Role = 0b00000001 // Access most game functions
	RoleModerator     Role = 0b00000010 // Global chat moderation commands
	RoleAdministrator Role = 0b00000100 // Server administration commands
	RoleGameMaster    Role = 0b00001000 // All other commands and actions
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
		roles:        RolePlayer,
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
	// The roles this account has been assigned
	roles Role
}

// TemplateName returns the template name for accounts
func (a *Account) TemplateName() string { return "Account" }

// SetTemplateName does nothing for accounts
func (a *Account) SetTemplateName(name string) {}

// Marshal writes the account data to a segment
func (a *Account) Marshal(s *marshal.TagFileSegment) {
	s.PutInt(uint32(a.player))
	s.PutString(a.username)
	s.PutString(a.passwordHash)
	s.PutByte(byte(a.roles))
	// NOTE: We can safely add Tag values below
}

// Deserialize does nothing
func (a *Account) Deserialize(f *util.TagFileObject) {}

// Unmarshal reads the account data from a segment
func (a *Account) Unmarshal(s *marshal.TagFileSegment) *marshal.TagCollection {
	a.player = uo.Serial(s.Int())
	a.username = s.String()
	a.passwordHash = s.String()
	a.roles = Role(s.Byte())
	return s.Tags()
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

// HasRole returns true if the account has the given role
func (a *Account) HasRole(r Role) bool { return a.roles&r != 0 }
