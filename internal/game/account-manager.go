package game

import (
	"io"
	"sync"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// AccountManager manages all of the accounts on the server
type AccountManager struct {
	// Database of accounts
	ds *util.DataStore
	// Read write lock for accounts manager
	lock sync.RWMutex
}

// NewAccount`Manager creates and returns a new AccountManager object
func NewAccountManager(rng uo.RandomSource) *AccountManager {
	return &AccountManager{
		ds: util.OpenOrCreateDataStore(rng, true),
	}
}

// Get gets the existing account by serial, or nil if it does not exist.
func (m *AccountManager) Get(id uo.Serial) *Account {
	m.lock.RLock()
	defer m.lock.RUnlock()
	as := m.ds.Get(id)
	if as == nil {
		return nil
	}
	return as.(*Account)
}

// GetOrCreate gets or creates a new account with the given details, or nil if
// the password hash did not match an existing account. If there are zero
// accounts in the data store, GetOrCreate creates the default admin account
// before attempting the lookup.
func (m *AccountManager) GetOrCreate(username, passwordHash string) *Account {
	var a *Account

	m.lock.Lock()
	defer m.lock.Unlock()
	if m.ds.Length() == 0 {
		adminUsername := "admin"
		adminPassword := "password"
		a := NewAccount(adminUsername, HashPassword(adminPassword))
		m.ds.Add(a, a.Username, uo.SerialTypeUnbound)
	}
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

// Save writes the account manager in tag file format.
func (m *AccountManager) Save(w io.Writer) []error {
	return m.ds.Save("accounts", w)
}

// Load loads the account manager in tag file format.
func (m *AccountManager) Load(r io.Reader) []error {
	errs := m.ds.Load(r)
	if errs != nil {
		return errs
	}

	// Rebuild the index
	return m.ds.Map(func(o util.Serializeable) error {
		a := o.(*Account)
		m.ds.AddIndex(a.Username, o)
		return nil
	})
}
