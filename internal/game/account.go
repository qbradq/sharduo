package game

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
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
	RoleAll           Role = 0b11111111 // All roles current and future
)

// Hashes a password suitable for the accounts database.
func HashPassword(password string) string {
	hd := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hd[:])
}

// Account holds all of the account information for one user
type Account struct {
	// Username
	username string
	// Password hash
	passwordHash string
	// Email address
	emailAddress string
	// Serial of the player's permanent mobile (not the currently controlled
	// mobile)
	player uo.Serial
	// List of serials of the player's pets in stable
	stabledPets util.Slice[Mobile]
	// The roles this account has been assigned
	roles Role
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
	s.PutInt(uint32(a.player))
	s.PutString(a.username)
	s.PutString(a.passwordHash)
	s.PutString(a.emailAddress)
	s.PutByte(byte(a.roles))
	s.PutByte(byte(len(a.stabledPets)))
	for _, pm := range a.stabledPets {
		s.PutObject(pm)
	}
}

// Deserialize does nothing
func (a *Account) Deserialize(t *template.Template, create bool) {}

// Unmarshal reads the account data from a segment
func (a *Account) Unmarshal(s *marshal.TagFileSegment) {
	a.player = uo.Serial(s.Int())
	a.username = s.String()
	a.passwordHash = s.String()
	a.emailAddress = s.String()
	a.roles = Role(s.Byte())
	n := int(s.Byte())
	a.stabledPets = make(util.Slice[Mobile], n)
	for i := 0; i < n; i++ {
		a.stabledPets[i] = s.Object().(Mobile)
	}
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

// EmailAddress returns the email address for the account
func (a *Account) EmailAddress() string { return a.emailAddress }

// SetEmailAddress sets the email address for the account
func (a *Account) SetEmailAddress(e string) { a.emailAddress = e }

// AddStabledPet attempts to add a pet to the account's stable list, returning
// an Error object describing why the action could not be performed on error
func (a *Account) AddStabledPet(p Mobile) *Error {
	if len(a.stabledPets) >= MaxStabledPets {
		return &Error{
			String: "You have too many animals in the stables already!",
		}
	}
	a.stabledPets = a.stabledPets.Append(p)
	return nil
}

// RemoveStabledPet attempts to remove the given pet from the account's stable
// list, returning true on success
func (a *Account) RemoveStabledPet(p Mobile) bool {
	a.stabledPets = a.stabledPets.Remove(p)
	return true
}

// StabledPets returns a slice of the all of the pets in this account's stable
func (a *Account) StabledPets() []Mobile { return a.stabledPets }

// UpdatePasswordByHash updates the account's password by hash value.
func (a *Account) UpdatePasswordByHash(hash string) {
	a.passwordHash = hash
}
