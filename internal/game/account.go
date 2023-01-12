package game

import (
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/google/uuid"
	"github.com/qbradq/sharduo/internal/marshal"
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
	// Serial of the player's permanent mobile (not the currently controlled
	// mobile)
	player uo.Serial
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
	f.WriteHex("Player", uint32(a.player))
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
	a.player = uo.Serial(f.GetHex("Player", uint32(uo.SerialMobileNil)))
}

// MarshalAccounts marshals all accounts in the map.
func MarshalAccounts(s *marshal.TagFileSegment, accounts map[uo.Serial]*Account) {
	for _, a := range accounts {
		s.PutInt(uint32(a.Serial()))
		s.PutString(a.username)
		s.PutString(a.passwordHash)
		s.PutInt(uint32(a.player))
		s.IncrementRecordCount()
	}
}

// UnmarshalAccounts unmarshals all accounts into a map.
func UnmarshalAccounts(s *marshal.TagFileSegment) map[uo.Serial]*Account {
	ret := make(map[uo.Serial]*Account)
	for i := uint32(0); i < s.RecordCount(); i++ {
		a := &Account{}
		a.SetSerial(uo.Serial(s.Int()))
		a.username = s.String()
		a.passwordHash = s.String()
		a.player = uo.Serial(s.Int())
		ret[a.Serial()] = a
	}
	return ret
}

// Username returns the username of the account
func (a *Account) Username() string {
	return a.username
}

// ComparePasswordHash returns true if the hash given matches this account's
func (a *Account) ComparePasswordHash(hash string) bool {
	return a.passwordHash == hash
}

// Player returns the player mobile serial, or uo.SerialMobileNil if none
func (a *Account) Player() uo.Serial { return a.player }

// SetPlayer sets the player mobile serial, or uo.SerialMobileNil if none
func (a *Account) SetPlayer(s uo.Serial) { a.player = s }
