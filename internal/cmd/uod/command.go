package uod

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	commandFactory.Add("location", newLocationCommand)
	commandFactory.Add("new", newNewCommand)
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
var commandFactory = util.NewFactory[string, CommandArgs, Command]("commands")

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
	return commandFactory.New(c[0], c)
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
func newLocationCommand(args CommandArgs) Command {
	return &LocationCommand{
		BaseCommand: BaseCommand{
			args: args,
		},
	}
}

// Compile implements the Command interface
func (c *LocationCommand) Compile() error {
	return nil
}

// Execute implements the Command interface
func (c *LocationCommand) Execute(n *NetState) error {
	if n == nil {
		return nil
	}
	world.SendTarget(n, uo.TargetTypeLocation, nil, func(r *clientpacket.TargetResponse, ctx interface{}) {
		n.SystemMessage("Location X=%d Y=%d Z=%d", r.X, r.Y, r.Z)
	})
	return nil
}

// NewCommand creates a new object from the named template
type NewCommand struct {
	BaseCommand
}

// newNewCommand constructs a new NewCommand
func newNewCommand(args CommandArgs) Command {
	return &NewCommand{
		BaseCommand: BaseCommand{
			args: args,
		},
	}
}

// Compile implements the Command interface
func (c *NewCommand) Compile() error {
	if len(c.args) != 2 {
		return fmt.Errorf("new command requires exactly 2 arguments, go %d", len(c.args))
	}
	return nil
}

// Execute implements the Command interface
func (c *NewCommand) Execute(n *NetState) error {
	if n == nil {
		return nil
	}
	world.SendTarget(n, uo.TargetTypeLocation, nil, func(r *clientpacket.TargetResponse, ctx interface{}) {
		o := templateManager.NewObject(c.args[1])
		if o == nil {
			return
		}
		o.SetLocation(uo.Location{
			X: r.X,
			Y: r.Y,
			Z: r.Z,
		})
		world.Map().AddNewObject(o)
	})
	return nil
}
