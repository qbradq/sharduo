package game

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/qbradq/sharduo/internal/util"
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
