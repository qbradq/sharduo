package commands

import (
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

// Contains commands every account is intended to have access to

func init() {
	reg(&cmDesc{"password", nil, commandPassword, game.RoleAll, "password new_password", "Changes the password for this account to the new one provided"})
}

func commandPassword(n game.NetState, args CommandArgs, cl string) {
	parts := strings.SplitN(cl, " ", 2)
	if len(parts) != 2 {
		n.Speech(nil, "A password is required.")
		return
	}
	n.Mobile().Account.PasswordHash = util.HashPassword(parts[1])
}
