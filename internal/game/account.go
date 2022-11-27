package game

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// Hashes a password suitable for the accounts database.
func HashPassword(password string) string {
	hd := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hd[:])
}

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
		a.SetSerial(m.ds.UniqueSerial(uo.SerialTypeUnbound))
		m.ds.Set(a, username)
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
