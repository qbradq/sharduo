package game

import (
	"io"
	"time"

	"github.com/qbradq/sharduo/lib/util"
)

// Role describes the roles that an account may have.
type Role uint8

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

// Account holds a copy of authentication information for use in certain game
// mechanics.
type Account struct {
	Username     string    // Account username
	PasswordHash string    // Hashed password
	Roles        Role      // Roles of the account
	Created      time.Time // Time of account creation
}

// NewAccount creates a new account with the given properties.
func NewAccount(username, passwordHash string, roles Role) *Account {
	return &Account{
		Username:     username,
		PasswordHash: passwordHash,
		Roles:        roles,
		Created:      time.Now(),
	}
}

// Write writes the account information to the writer.
func (a *Account) Write(w io.Writer) {
	util.PutUInt32(w, 0)              // Version
	util.PutString(w, a.Username)     // Username
	util.PutString(w, a.PasswordHash) // Hash of the password as a hex string
	util.PutByte(w, byte(a.Roles))    // Roles bit mask
	util.PutTime(w, a.Created)        // Time of account creation
}

// Read reads the account information from the reader.
func (a *Account) Read(r io.Reader) {
	_ = util.GetUInt32(r)              // Version
	a.Username = util.GetString(r)     // Username
	a.PasswordHash = util.GetString(r) // Has of the password as a hex string
	a.Roles = Role(util.GetByte(r))    // Roles bit mask
	a.Created = util.GetTime(r)        // Time of account creation
}

// HasRole returns true if the account holds *all* roles given.
func (a *Account) HasRole(r Role) bool {
	return a.Roles&r == r
}
