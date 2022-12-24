package uod

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	commandFactory.Add("debug", newDebugCommand)
	commandFactory.Add("location", newLocationCommand)
	commandFactory.Add("new", newNewCommand)
	commandFactory.Add("teleport", newTeleportCommand)
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
		return fmt.Errorf("new command requires exactly 2 arguments, got %d", len(c.args))
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
			n.SystemMessage("template %s not found", c.args[1])
			return
		}
		o.SetLocation(uo.Location{
			X: r.X,
			Y: r.Y,
			Z: r.Z,
		})
		world.Map().AddObject(world.AddNewObjectToDataStores(o))
	})
	return nil
}

// TeleportCommand teleports the user either by target to an absolute location
type TeleportCommand struct {
	BaseCommand
	Targeted      bool
	MultiTargeted bool
	Location      uo.Location
}

// newTeleportCommand constructs a new TeleportCommand
func newTeleportCommand(args CommandArgs) Command {
	return &TeleportCommand{
		BaseCommand: BaseCommand{
			args: args,
		},
	}
}

// Compile implements the Command interface
func (c *TeleportCommand) Compile() error {
	c.Location.Z = uo.MapMaxZ
	if len(c.args) > 4 {
		return errors.New("teleport command expects a maximum of 3 arguments")
	}
	if len(c.args) == 4 {
		z, err := strconv.ParseInt(c.args[3], 0, 32)
		if err != nil {
			return err
		}
		c.Location.Z = int(z)
	}
	if len(c.args) > 3 {
		y, err := strconv.ParseInt(c.args[2], 0, 32)
		if err != nil {
			return err
		}
		c.Location.Y = int(y)
		x, err := strconv.ParseInt(c.args[1], 0, 32)
		if err != nil {
			return err
		}
		c.Location.X = int(x)
	}
	if len(c.args) == 2 {
		if c.args[1] == "multi" {
			c.Targeted = true
			c.MultiTargeted = true
		} else {
			return errors.New("incorrect usage of teleport command. Use [teleport (multi|X Y|X Y Z)")
		}
	}
	if len(c.args) == 1 {
		c.Targeted = true
	}
	return nil
}

// Execute implements the Command interface
func (c *TeleportCommand) Execute(n *NetState) error {
	if n.m == nil {
		return nil
	}
	if !c.Targeted {
		if c.Location.Z == uo.MapMaxZ {
			surface := world.Map().GetTopSurface(c.Location, uo.MapMaxZ)
			c.Location.Z = surface.Z()
		}
		if !world.Map().TeleportMobile(n.m, c.Location) {
			return errors.New("something is blocking that location")
		}
		return nil
	}

	var fn func(*clientpacket.TargetResponse, interface{})
	fn = func(r *clientpacket.TargetResponse, ctx interface{}) {
		if !world.Map().TeleportMobile(n.m, uo.Location{
			X: r.X,
			Y: r.Y,
			Z: r.Z,
		}) {
			n.SystemMessage("something is blocking that location")
		}
		if c.MultiTargeted {
			world.SendTarget(n, uo.TargetTypeLocation, nil, fn)
		}
	}
	world.SendTarget(n, uo.TargetTypeLocation, nil, fn)

	return nil
}

// DebugCommand is where I shove random test commands for development
type DebugCommand struct {
	BaseCommand
}

// newDebugCommand constructs a new DebugCommand
func newDebugCommand(args CommandArgs) Command {
	return &DebugCommand{
		BaseCommand: BaseCommand{
			args: args,
		},
	}
}

// Compile implements the Command interface
func (c *DebugCommand) Compile() error {
	if len(c.args) < 2 {
		return errors.New("debug command requires at least 2 arguments")
	}
	switch c.args[1] {
	case "splat":
		if len(c.args) != 3 {
			return errors.New("debug splat command requires 3 arguments")
		}
	default:
		return fmt.Errorf("unknown debug command %s", c.args[1])
	}
	return nil
}

// Execute implements the Command interface
func (c *DebugCommand) Execute(n *NetState) error {
	if n == nil || n.m == nil {
		return nil
	}
	if c.args[1] == "splat" {
		start := time.Now().UnixMilli()
		count := 0
		for iy := n.m.Location().Y - 50; iy < n.m.Location().Y+50; iy++ {
			for ix := n.m.Location().X - 50; ix < n.m.Location().X+50; ix++ {
				o := templateManager.NewObject(c.args[2])
				if o == nil {
					return fmt.Errorf("debug splat failed to find template %s", c.args[2])
				}
				o.SetLocation(uo.Location{
					X: ix,
					Y: iy,
					Z: n.m.Location().Z,
				})
				world.Map().AddObject(world.AddNewObjectToDataStores((o)))
				count++
			}
		}
		end := time.Now().UnixMilli()
		n.SystemMessage("generated %d items in %d milliseconds\n", count, end-start)
	}
	return nil
}
