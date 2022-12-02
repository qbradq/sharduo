package game

import (
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/google/uuid"
	"github.com/qbradq/sharduo/internal/util"
)

func init() {
	util.RegisterCtor(func() util.Serializeable { return &Account{} })
}

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

// GetTypeName implements the util.Serializeable interface.
func (a *Account) GetTypeName() string {
	return "Account"
}

// Serialize implements the util.Serializeable interface.
func (a *Account) Serialize(f *util.TagFileWriter) {
	a.BaseSerializeable.Serialize(f)
	f.WriteString("Username", a.Username)
	f.WriteString("PasswordHash", a.PasswordHash)
}

// Deserialize implements the util.Serializeable interface.
func (a *Account) Deserialize(f *util.TagFileObject) {
	a.BaseSerializeable.Deserialize(f)
	accountRecoveryString := "__bad_account__" + uuid.New().String()
	a.Username = f.GetString("Username", accountRecoveryString)
	if a.Username == accountRecoveryString {
		log.Println("account recovery required:", accountRecoveryString)
	}
	a.PasswordHash = f.GetString("PasswordHash", "")
}

// NewAccount creates a new account object
func NewAccount(username, passwordHash string) *Account {
	return &Account{
		Username:     username,
		PasswordHash: passwordHash,
	}
}
