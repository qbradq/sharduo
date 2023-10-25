package commands

import (
	"encoding/csv"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

// Server callbacks
var globalChat func(uo.Hue, string, string)
var saveWorld func()
var broadcast func(string, ...any)
var shutdown func()
var latestSavePath func() string

// RegisterCallbacks registers the various server callbacks required to execute
// certain commands.
func RegisterCallbacks(
	lGlobalChat func(uo.Hue, string, string),
	lSaveWorld func(),
	lBroadcast func(string, ...any),
	lShutdown func(),
	lLatestSavePath func() string,
) {
	globalChat = lGlobalChat
	saveWorld = lSaveWorld
	broadcast = lBroadcast
	shutdown = lShutdown
	latestSavePath = lLatestSavePath
}

// regcmd registers a command description
func regcmd(d *cmdesc) {
	commands[d.name] = d
	for _, alt := range d.alts {
		commands[alt] = d
	}
}

// commandFunction is the signature of a command function
type commandFunction func(game.NetState, CommandArgs, string)

// cmdesc describes a command
type cmdesc struct {
	name        string
	alts        []string
	fn          commandFunction
	roles       game.Role
	usage       string
	description string
}

// commands is the mapping of command strings to commandFunction's
var commands = make(map[string]*cmdesc)

// Execute executes the command for the given command line
func Execute(n game.NetState, line string) {
	if n.Account() == nil {
		return
	}
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ' ' // Space
	c, err := r.Read()
	if err != nil {
		n.Speech(nil, "command not found")
		return
	}
	if len(c) == 0 {
		n.Speech(nil, "command not found")
		return
	}
	desc := commands[c[0]]
	if desc == nil {
		n.Speech(nil, "%s is not a command", c[0])
		return
	}
	if !n.Account().HasRole(desc.roles) {
		n.Speech(nil, "you do not have permission to use the %s command", c[0])
		return
	}
	desc.fn(n, c, line)
}
