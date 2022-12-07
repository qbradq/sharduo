package uod

import (
	"encoding/csv"
	"log"
	"strings"

	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	commandFactory.Add("location", newLocationCommand)
}

// Command is the interface all command objects implement
type Command interface {
	// Compile takes all appropriate steps that can be done in advance of
	// command execution.
	Compile() error
	// Execute executes the command and should only be called after a call
	// to Compile. Execute may be ran multiple times per object.
	Execute(*NetState) error
}

// commandFactory manages the available commands
var commandFactory = util.NewFactory[string, CommandArgs]("commands")

// ParseCommand returns a Command object parsed from a command line
func ParseCommand(line string) Command {
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ' ' // Space
	c, err := r.Read()
	if err != nil {
		log.Println(err)
		return nil
	}
	if len(c) == 0 {
		return nil
	}
	cmd := commandFactory.New(c[0], c)
	if cmd != nil {
		return cmd.(Command)
	}
	return nil
}

// BaseCommand implements some basic command functionality
type BaseCommand struct {
	args CommandArgs
}

// LocationCommand reports the location of a target
type LocationCommand struct {
	BaseCommand
}

// newLocationCommand constructs a new LocationCommand
func newLocationCommand(args CommandArgs) any {
	return &LocationCommand{
		BaseCommand: BaseCommand{
			args: args,
		},
	}
}

// Compile implements the Command interface
func (l *LocationCommand) Compile() error {
	return nil
}

// Execute implements the Command interface
func (l *LocationCommand) Execute(n *NetState) error {
	if n == nil {
		return nil
	}
	n.Send(&serverpacket.Target{

	})
}
