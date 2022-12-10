package game

import (
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/google/uuid"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &Account{} })
}

// Hashes a password suitable for the accounts database.
func HashPassword(password string) string {
	hd := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hd[:])
}

// NewAccount creates a new account object
func NewAccount(username, passwordHash string) *Account {
	return &Account{
		username:     username,
		passwordHash: passwordHash,
	}
}

// Account holds all of the account information for one user
type Account struct {
	util.BaseSerializeable
	// Username
	username string
	// Password hash
	passwordHash string
}

// GetTypeName implements the util.Serializeable interface.
func (a *Account) TypeName() string {
	return "Account"
}

// GetSerialType implements the util.Serializeable interface.
func (a *Account) SerialType() uo.SerialType {
	return uo.SerialTypeUnbound
}

// Serialize implements the util.Serializeable interface.
func (a *Account) Serialize(f *util.TagFileWriter) {
	a.BaseSerializeable.Serialize(f)
	f.WriteString("Username", a.username)
	f.WriteString("PasswordHash", a.passwordHash)
}

// Deserialize implements the util.Serializeable interface.
func (a *Account) Deserialize(f *util.TagFileObject) {
	a.BaseSerializeable.Deserialize(f)
	accountRecoveryString := "__bad_account__" + uuid.New().String()
	a.username = f.GetString("Username", accountRecoveryString)
	if a.username == accountRecoveryString {
		log.Println("account recovery required:", accountRecoveryString)
	}
	a.passwordHash = f.GetString("PasswordHash", "")
}

// Username returns the username of the account
func (a *Account) Username() string {
	return a.username
}

// ComparePasswordHash returns true if the hash given matches this account's
func (a *Account) ComparePasswordHash(hash string) bool {
	return a.passwordHash == hash
}
