package game

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	marshal.RegisterCtor(marshal.ObjectTypeAccount, func() interface{} { return &Account{} })
}

const (
	MaxStabledPets int = 10 // Maximum number of pets to hold in the stables
)

// Role describes the roles that an account may have.
type Role byte

const (
	RolePlayer        Role = 0b00000001 // Access most game functions
	RoleModerator     Role = 0b00000010 // Global chat moderation commands
	RoleAdministrator Role = 0b00000100 // Server administration commands
	RoleGameMaster    Role = 0b00001000 // Most other commands and actions
	RoleDeveloper     Role = 0b00010000 // Commands and actions that can be dangerous on a live shard
	RoleSuperUser     Role = 0b10000000 // Marks the account as the super user
	RoleStaff         Role = 0b10011100 // All roles considered "staff", with a GM body
	RoleAll           Role = 0b11111111 // All roles current and future
)

// Hashes a password suitable for the accounts database.
func HashPassword(password string) string {
	hd := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hd[:])
}

// Account holds all of the account information for one user
type Account struct {
	username            string    // Username
	passwordHash        string    // Password hash
	passwordSetAt       time.Time // The last time this account's password was changed
	failedLoginAttempts int       // The number of times someone has tried to login to this account and failed consecutively
	locked              bool      // If true this account is locked for interactive login, probably due to too many consecutive failed login attempts
	suspendedUntil      time.Time // End of the most recent account suspension
	emailAddress        string    // Email address
	player              uo.Serial // Serial of the player's permanent mobile (not the currently controlled mobile)
	roles               Role      // The roles this account has been assigned
}

// NewAccount creates a new account object
func NewAccount(username, passwordHash string, roles Role) *Account {
	return &Account{
		username:     username,
		passwordHash: passwordHash,
		roles:        roles,
	}
}

// TemplateName returns the template name for accounts
func (a *Account) TemplateName() string { return "Account" }

// SetTemplateName does nothing for accounts
func (a *Account) SetTemplateName(name string) {}

// Marshal writes the account data to a segment
func (a *Account) Marshal(s *marshal.TagFileSegment) {
	s.PutInt(0) // Version
	s.PutInt(uint32(a.player))
	s.PutString(a.username)
	s.PutString(a.passwordHash)
	s.PutLong(uint64(a.passwordSetAt.Unix()))
	s.PutInt(uint32(a.failedLoginAttempts))
	s.PutBool(a.locked)
	s.PutLong(uint64(a.suspendedUntil.Unix()))
	s.PutString(a.emailAddress)
	s.PutByte(byte(a.roles))
}

// Deserialize does nothing
func (a *Account) Deserialize(t *template.Template, create bool) {}

// Unmarshal reads the account data from a segment
func (a *Account) Unmarshal(s *marshal.TagFileSegment) {
	_ = s.Int() // Version
	a.player = uo.Serial(s.Int())
	a.username = s.String()
	a.passwordHash = s.String()
	a.passwordSetAt = time.Unix(int64(s.Long()), 0)
	a.failedLoginAttempts = int(s.Int())
	a.locked = s.Bool()
	a.suspendedUntil = time.Unix(int64(s.Long()), 0)
	a.emailAddress = s.String()
	a.roles = Role(s.Byte())
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

// ToggleRole toggles the given role.
func (a *Account) ToggleRole(r Role) { a.roles ^= r }

// EmailAddress returns the email address for the account
func (a *Account) EmailAddress() string { return a.emailAddress }

// SetEmailAddress sets the email address for the account
func (a *Account) SetEmailAddress(e string) { a.emailAddress = e }

// UpdatePasswordByHash updates the account's password by hash value.
func (a *Account) UpdatePasswordByHash(hash string) {
	a.passwordHash = hash
	a.passwordSetAt = time.Now()
	a.failedLoginAttempts = 0
}

// IncrementFailedLoginCount increments the failed login count and returns it.
func (a *Account) IncrementFailedLoginCount() int {
	a.failedLoginAttempts++
	return a.failedLoginAttempts
}

// Locked returns true if the account is locked.
func (a *Account) Locked() bool { return a.locked }

// Lock locks the account.
func (a *Account) Lock() { a.locked = true }

// Unlock unlocks the account.
func (a *Account) Unlock() { a.locked = false }

// SuspendedUntil returns the time the latest suspension ends.
func (a *Account) SuspendedUntil() time.Time { return a.suspendedUntil }

// Suspend suspends the account for the given amount of time.
func (a *Account) Suspend(d time.Duration) {
	a.suspendedUntil = time.Now().Add(d)
}
