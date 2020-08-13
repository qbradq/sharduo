package common

import (
	"crypto/sha256"
	"sync"
)

// An Account represents a player's standing with the shard
type Account struct {
	username string
	passhash [sha256.Size]byte
	lock     sync.Mutex
	roles    Role
}

// GetAccount returns a possibly new account with the proivded information, or
// nil if there was an issue. See err for error details.
func GetAccount(username, password string) (*Account, error) {
	return &Account{
		username: username,
		passhash: sha256.Sum256([]byte(password)),
	}, nil
}

// Roles returns a Role value representing all roles of the authenticated account
func (a *Account) Roles() Role {
	// Role is an atomic read
	return a.roles
}

// AddRole adds a Role to the account
func (a *Account) AddRole(r Role) {
	// Writes are never atomic, and the atomic package does not support bitwise ops
	a.lock.Lock()
	defer a.lock.Unlock()
	a.roles = a.roles | r
}

// RemoveRole removes a Role from the account
func (a *Account) RemoveRole(r Role) {
	// Writes are never atomic, and the atomic package does not support bitwise ops
	a.lock.Lock()
	defer a.lock.Unlock()
	a.roles = a.roles & (^r)
}
