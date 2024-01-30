package game

import (
	"io"
	"time"

	"github.com/qbradq/sharduo/lib/uo"
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
	Username       string      // Account username
	PasswordHash   string      // Hashed password
	Roles          Role        // Roles of the account
	Created        time.Time   // Time of account creation
	SuspendedUntil time.Time   // End of most recent suspension
	Locked         bool        // If true the account may not be logged into
	Characters     []uo.Serial // Serials of all character mobiles
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
	util.PutUInt32(w, 0)                     // Version
	util.PutString(w, a.Username)            // Username
	util.PutString(w, a.PasswordHash)        // Hash of the password as a hex string
	util.PutByte(w, byte(a.Roles))           // Roles bit mask
	util.PutTime(w, a.Created)               // Time of account creation
	util.PutTime(w, a.SuspendedUntil)        // End of most recent suspension
	util.PutBool(w, a.Locked)                // Account lock status
	util.PutByte(w, byte(len(a.Characters))) // Character count
	for _, s := range a.Characters {         // Character serials
		util.PutUInt32(w, uint32(s))
	}
}

// Read reads the account information from the reader.
func (a *Account) Read(r io.Reader) {
	_ = util.GetUInt32(r)               // Version
	a.Username = util.GetString(r)      // Username
	a.PasswordHash = util.GetString(r)  // Has of the password as a hex string
	a.Roles = Role(util.GetByte(r))     // Roles bit mask
	a.Created = util.GetTime(r)         // Time of account creation
	a.SuspendedUntil = util.GetTime(r)  // End of most recent suspension
	a.Locked = util.GetBool(r)          // Account lock status
	n := int(util.GetByte(r))           // Character count
	a.Characters = make([]uo.Serial, n) // Character serials
	for i := 0; i < n; i++ {
		a.Characters[i] = uo.Serial(util.GetUInt32(r))
	}
}

// HasRole returns true if the account holds *all* roles given.
func (a *Account) HasRole(r Role) bool {
	return a.Roles&r == r
}
