package game

import (
	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// Account holds all of the account information for one user
type Account struct {
	util.BaseSerializeable
	// Username
	Username string
	// Password hash
	PasswordHash string
}

// NewAccount creates a new account object
func NewAccount(username, passwordHash string) *Account {
	return &Account{
		Username:     username,
		PasswordHash: passwordHash,
	}
}

// AccountManager manages all of the accounts on the server
type AccountManager struct {
	// Database of accounts
	ds *util.DataStore
	// Serial manager for accounts
	sm *uo.SerialManager
}

// NewAccountManager creates and returns a new AccountManager object
func NewAccountManager(dbpath string) *AccountManager {
	return &AccountManager{
		ds: util.NewDataStore(dbpath, true),
		sm: uo.NewSerialManager(),
	}
}

// GetOrCreate gets or creates a new account with the given details, or nil if
// the password hash did not match an existing account.
func (m *AccountManager) GetOrCreate(username, passwordHash string) *Account {
	var a *Account

	as := m.ds.GetByIndex(username)
	if as == nil {
		a = NewAccount(username, passwordHash)
		m.ds.Set(a, username)
	} else {
		a = as.(*Account)
	}
	if a.PasswordHash != passwordHash {
		return nil
	}
	return a
}
