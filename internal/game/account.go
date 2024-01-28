package game

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
	Username     string // Account username
	PasswordHash string // Hashed password
	Roles        Role   // Roles of the account
}

// HasRole returns true if the account holds *all* roles given.
func (a *Account) HasRole(r Role) bool {
	return a.Roles&r == r
}
