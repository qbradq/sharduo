package game

import (
	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// AccountManager manages all of the accounts on the server
type AccountManager struct {
	// Database of accounts
	ds *util.DataStore
}

// NewAccountManager creates and returns a new AccountManager object
func NewAccountManager(dbpath string) *AccountManager {
	adminUsername := "admin"
	adminPassword := "password"
	m := &AccountManager{
		ds: util.OpenOrCreateDataStore(dbpath, true),
	}
	m.GetOrCreate(adminUsername, HashPassword(adminPassword))
	return m
}

// GetOrCreate gets or creates a new account with the given details, or nil if
// the password hash did not match an existing account.
func (m *AccountManager) GetOrCreate(username, passwordHash string) *Account {
	var a *Account

	as := m.ds.GetByIndex(username)
	if as == nil {
		a = NewAccount(username, passwordHash)
		m.ds.Add(a, username, uo.SerialTypeUnbound)
	} else {
		a = as.(*Account)
	}
	if a.PasswordHash != passwordHash {
		return nil
	}
	return a
}

// Save writes the account manager to a JSON file.
func (m *AccountManager) Save(dbpath string) error {
	return m.ds.Save(dbpath)
}
