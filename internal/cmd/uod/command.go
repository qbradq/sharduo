package uod

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/game/events"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	commands["bank"] = &cmdesc{commandBank, game.RoleGameMaster, "bank", "Opens the bank box of the targeted mobile, if any"}
	commands["broadcast"] = &cmdesc{commandBroadcast, game.RoleAdministrator, "broadcast text", "Broadcasts the given text to all connected players"}
	commands["c"] = &cmdesc{commandChat, game.RolePlayer, "c", "Sends global chat speech"}
	commands["chat"] = &cmdesc{commandChat, game.RolePlayer, "chat", "Sends global chat speech"}
	commands["debug"] = &cmdesc{commandDebug, game.RoleGameMaster, "debug command [arguments]", "Executes debug commands"}
	commands["g"] = &cmdesc{commandChat, game.RolePlayer, "g", "Sends global chat speech"}
	commands["global"] = &cmdesc{commandChat, game.RolePlayer, "global", "Sends global chat speech"}
	commands["location"] = &cmdesc{commandLocation, game.RoleAdministrator, "location", "Tells the absolute location of the targeted location or object"}
	commands["new"] = &cmdesc{commandNew, game.RoleGameMaster, "new template_name [stack_amount]", "Creates a new item with an optional stack amount"}
	commands["save"] = &cmdesc{commandSave, game.RoleAdministrator, "save", "Executes a world save immediately"}
	commands["shutdown"] = &cmdesc{commandShutdown, game.RoleAdministrator, "shutdown", "Shuts down the server immediately"}
	commands["snapshot_clean"] = &cmdesc{commandSnapshotClean, game.RoleAdministrator, "snapshot_clean", "internal command, please do not use"}
	commands["snapshot_daily"] = &cmdesc{commandSnapshotDaily, game.RoleAdministrator, "snapshot_daily", "internal command, please do not use"}
	commands["snapshot_weekly"] = &cmdesc{commandSnapshotWeekly, game.RoleAdministrator, "snapshot_weekly", "internal command, please do not use"}
	commands["static"] = &cmdesc{commandStatic, game.RoleGameMaster, "static graphic_number", "Creates a new static object with the given graphic number"}
	commands["teleport"] = &cmdesc{commandTeleport, game.RoleGameMaster, "teleport [x y|x y z|multi]", "Teleports you to the targeted location - optionally multiple times, or to the top Z of the given X/Y location, or to the absolute location"}
}

// commandFunction is the signature of a command function
type commandFunction func(*NetState, CommandArgs, string)

// cmdesc describes a command
type cmdesc struct {
	fn          commandFunction
	roles       game.Role
	usage       string
	description string
}

// commands is the mapping of command strings to commandFunction's
var commands = make(map[string]*cmdesc)

// ExecuteCommand executes the command for the given command line
func ExecuteCommand(n *NetState, line string) {
	if n.account == nil {
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
	if !n.account.HasRole(desc.roles) {
		n.Speech(nil, "you do not have permission to use the %s command", c[0])
		return
	}
	desc.fn(n, c, line)
}

func commandChat(n *NetState, args CommandArgs, cl string) {
	if n.m == nil {
		return
	}
	parts := strings.SplitN(cl, " ", 2)
	if len(parts) != 2 {
		return
	}
	GlobalChat(n.m.DisplayName(), parts[1])
}

func commandLocation(n *NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeLocation, func(r *clientpacket.TargetResponse) {
		n.Speech(nil, "Location X=%d Y=%d Z=%d", r.Location.X, r.Location.Y, r.Location.Z)
	})
}

func commandNew(n *NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	if len(args) < 2 || len(args) > 3 {
		n.Speech(nil, "new command requires 2 or 3 arguments, got %d", len(args))
	}
	n.TargetSendCursor(uo.TargetTypeLocation, func(r *clientpacket.TargetResponse) {
		o := template.Create(args[1]).(game.Object)
		if o == nil {
			n.Speech(nil, "failed to create object with template %s", args[1])
			return
		}
		o.SetLocation(r.Location)
		if len(args) == 3 {
			item, ok := o.(game.Item)
			if !ok {
				n.Speech(nil, "amount specified for non-item %s", args[1])
				return
			}
			if !item.Stackable() {
				n.Speech(nil, "amount specified for non-stackable item %s", args[1])
				return
			}
			v := args.Int(2)
			if v < 1 {
				v = 1
			}
			item.SetAmount(v)
		}
		// Try to add the object to the map legit, but if that fails just force
		// it so we don't leak it.
		if !world.Map().AddObject(o) {
			world.Map().ForceAddObject(o)
		}
	})
}

func commandTeleport(n *NetState, args CommandArgs, cl string) {
	if n.m == nil {
		return
	}
	targeted := false
	multi := false
	l := uo.Location{}
	l.Z = uo.MapMaxZ
	if len(args) > 4 {
		n.Speech(nil, "teleport command expects a maximum of 3 arguments")
		return
	}
	if len(args) == 4 {
		l.Z = int8(args.Int(3))
	}
	if len(args) > 3 {
		l.Y = int16(args.Int(2))
		l.X = int16(args.Int(1))
	}
	if len(args) == 2 {
		if args[1] == "multi" {
			targeted = true
			multi = true
		} else {
			n.Speech(nil, "incorrect usage of teleport command. Use [teleport (multi|X Y|X Y Z)")
			return
		}
	}
	if len(args) == 1 {
		targeted = true
	}
	l = l.Bound()
	if !targeted {
		if l.Z == uo.MapMaxZ {
			floor, _ := world.Map().GetFloorAndCeiling(l, false)
			if floor == nil {
				n.Speech(nil, "location has no floor")
				return
			}
			l.Z = floor.Z()
		}
		if !world.Map().TeleportMobile(n.m, l) {
			n.Speech(nil, "something is blocking that location")
		}
		return
	}

	var fn func(*clientpacket.TargetResponse)
	fn = func(r *clientpacket.TargetResponse) {
		if n.m == nil {
			return
		}
		if !world.Map().TeleportMobile(n.m, r.Location) {
			n.Speech(nil, "something is blocking that location")
		}
		if multi {
			n.TargetSendCursor(uo.TargetTypeLocation, fn)
		}
	}
	n.TargetSendCursor(uo.TargetTypeLocation, fn)
}

func commandDebug(n *NetState, args CommandArgs, cl string) {
	if n == nil || n.m == nil {
		return
	}
	// Sanitize input
	if len(args) < 2 {
		n.Speech(nil, "debug command requires a command name: [debug command_name")
		return
	}
	switch args[1] {
	case "memory_test":
		fallthrough
	case "delay_test":
		fallthrough
	case "mount":
		fallthrough
	case "shirtbag":
		if len(args) != 2 {
			n.Speech(nil, "debug %s command requires 0 arguments", args[1])
			return
		}
	case "global_light":
		fallthrough
	case "music":
		fallthrough
	case "personal_light":
		fallthrough
	case "sound":
		fallthrough
	case "splat":
		if len(args) != 3 {
			n.Speech(nil, "debug %s command requires 1 arguments", args[1])
			return
		}
	case "animate":
		if len(args) != 4 {
			n.Speech(nil, "debug %s command requires 2 arguments", args[1])
			return
		}
	default:
		n.Speech(nil, "unknown debug command %s", args[1])
		return
	}
	// Execute command
	switch args[1] {
	case "global_light":
		ll := uo.LightLevel(args.Int(2))
		n.Send(&serverpacket.GlobalLightLevel{
			LightLevel: ll,
		})
	case "personal_light":
		ll := uo.LightLevel(args.Int(2))
		n.Send(&serverpacket.PersonalLightLevel{
			Serial:     n.m.Serial(),
			LightLevel: ll,
		})
	case "animate":
		at := uo.AnimationType(args.Int(2))
		aa := uo.AnimationAction(args.Int(2))
		n.Animate(n.m, at, aa)
	case "music":
		which := uo.Song(args.Int(2))
		n.Music(which)
	case "sound":
		which := uo.Sound(args.Int(2))
		world.Map().PlaySound(which, n.m.Location())
	case "memory_test":
		start := time.Now()
		for i := 0; i < 1_000_000; i++ {
			o := template.Create("FancyShirt").(game.Object)
			o.SetLocation(uo.Location{
				X: int16(world.Random().Random(100, uo.MapWidth-101)),
				Y: int16(world.Random().Random(100, uo.MapHeight-101)),
				Z: 65,
			})
			world.Map().ForceAddObject(o)
		}
		for i := 0; i < 150_000; i++ {
			o := template.Create("Banker").(game.Object)
			o.SetLocation(uo.Location{
				X: int16(world.Random().Random(100, uo.MapWidth-101)),
				Y: int16(world.Random().Random(100, uo.MapHeight-101)),
				Z: 65,
			})
			world.Map().ForceAddObject(o)
		}
		end := time.Now()
		log.Printf("operation completed in %s", end.Sub(start))
	case "delay_test":
		game.NewTimer(uo.DurationSecond*5, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*10, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*15, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*20, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*25, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*30, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*35, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*40, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*45, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*50, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*55, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*60, "WhisperTime", n.m, n.m)
		game.NewTimer(uo.DurationSecond*65, "WhisperTime", n.m, n.m)
	case "mount":
		mi := template.Create("HorseMountItem")
		n.m.ForceEquip(mi.(*game.MountItem))
	case "shirtbag":
		backpack := template.Create("Backpack").(game.Container)
		backpack.SetLocation(n.m.Location())
		world.m.SetNewParent(backpack, nil)
		for i := 0; i < 125; i++ {
			shirt := template.Create("FancyShirt").(game.Object)
			shirt.SetLocation(uo.RandomContainerLocation)
			if !world.Map().SetNewParent(shirt, backpack) {
				n.Speech(nil, "failed to add an item to the backpack")
				break
			}
		}
	case "splat":
		start := time.Now()
		count := 0
		for iy := n.m.Location().Y - 50; iy < n.m.Location().Y+50; iy++ {
			for ix := n.m.Location().X - 50; ix < n.m.Location().X+50; ix++ {
				o := template.Create(args[2]).(game.Object)
				if o == nil {
					n.Speech(nil, "debug splat failed to create object with template %s", args[2])
				}
				o.SetLocation(uo.Location{
					X: ix,
					Y: iy,
					Z: n.m.Location().Z,
				})
				world.Map().SetNewParent(o, nil)
				count++
			}
		}
		end := time.Now()
		n.Speech(nil, "generated %d items in %d milliseconds\n", count, end.Sub(start).Milliseconds())
	}
}

func commandBank(n *NetState, args CommandArgs, cl string) {
	if n == nil || n.m == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(r *clientpacket.TargetResponse) {
		events.OpenBankBox(nil, n.m, nil)
	})
}

func commandStatic(n *NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	if len(args) != 2 {
		n.Speech(nil, "usage: static item_id")
		return
	}
	g := uo.Graphic(args.Int(1))
	if g.IsNoDraw() {
		n.Speech(nil, "refusing to create no-draw static 0x%04X", g)
	}
	n.TargetSendCursor(uo.TargetTypeLocation, func(r *clientpacket.TargetResponse) {
		o := template.Create("StaticItem")
		if o == nil {
			n.Speech(nil, "StaticItem template not found")
			return
		}
		i, ok := o.(game.Item)
		if !ok {
			n.Speech(nil, "StaticItem template was not an item")
			return
		}
		i.SetBaseGraphic(g)
		i.SetLocation(r.Location)
		world.Map().ForceAddObject(i)
	})
}

func commandSave(n *NetState, args CommandArgs, cl string) {
	world.Marshal()
}

func commandShutdown(n *NetState, args CommandArgs, cl string) {
	gracefulShutdown()
}

func commandBroadcast(n *NetState, args CommandArgs, cl string) {
	parts := strings.SplitN(cl, " ", 2)
	if len(parts) != 2 {
		return
	}
	Broadcast(parts[1])
}

func commandSnapshotDaily(n *NetState, args CommandArgs, cl string) {
	// Make sure the archive directory exists
	os.MkdirAll(configuration.ArchiveDirectory, 0777)
	// Create an archive copy of the save file
	p := world.LatestSavePath()
	src, err := os.Open(p)
	if err != nil {
		n.Speech(nil, "error: failed to create daily archive: %s", err)
		return
	}
	defer src.Close()
	dest, err := os.Create(path.Join(configuration.ArchiveDirectory, "daily.sav.gz"))
	if err != nil {
		n.Speech(nil, "error: failed to create daily archive: %s", err)
		return
	}
	_, err = io.Copy(dest, src)
	dest.Close()
	if err != nil {
		n.Speech(nil, "error: failed to create daily archive: %s", err)
		return
	}
	// Remove the oldest save
	os.Remove(path.Join(configuration.ArchiveDirectory, "daily7.sav.gz"))
	// Rotate daily saves
	for i := 6; i > 0; i-- {
		op := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("daily%d.sav.gz", i))
		np := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("daily%d.sav.gz", i+1))
		os.Rename(op, np)
	}
	// Move the new save file into the archives
	os.Rename(path.Join(configuration.ArchiveDirectory, "daily.sav.gz"),
		path.Join(configuration.ArchiveDirectory, "daily1.sav.gz"))
	n.Speech(nil, "daily archive complete")
}

func commandSnapshotWeekly(n *NetState, args CommandArgs, cl string) {
	// Make sure the archive directory exists
	os.MkdirAll(configuration.ArchiveDirectory, 0777)
	// Create an archive copy of the save file
	p := world.LatestSavePath()
	src, err := os.Open(p)
	if err != nil {
		n.Speech(nil, "error: failed to create weekly archive: %s", err)
		return
	}
	defer src.Close()
	dest, err := os.Create(path.Join(configuration.ArchiveDirectory, "weekly.sav.gz"))
	if err != nil {
		n.Speech(nil, "error: failed to create weekly archive: %s", err)
		return
	}
	_, err = io.Copy(dest, src)
	dest.Close()
	if err != nil {
		n.Speech(nil, "error: failed to create weekly archive: %s", err)
		return
	}
	// Remove the oldest save
	os.Remove(path.Join(configuration.ArchiveDirectory, "weekly52.sav.gz"))
	// Rotate daily saves
	for i := 51; i > 0; i-- {
		op := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("weekly%d.sav.gz", i))
		np := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("weekly%d.sav.gz", i+1))
		os.Rename(op, np)
	}
	// Move the new save file into the archives
	os.Rename(path.Join(configuration.ArchiveDirectory, "weekly.sav.gz"),
		path.Join(configuration.ArchiveDirectory, "weekly1.sav.gz"))
	n.Speech(nil, "weekly archive complete")
}

func commandSnapshotClean(n *NetState, args CommandArgs, cl string) {
	t := time.Now().Add(time.Hour * 72 * -1)
	filepath.WalkDir(configuration.SaveDirectory, func(p string, d fs.DirEntry, e error) error {
		log.Println(p)
		if d.IsDir() {
			return nil
		}
		di, err := d.Info()
		if err != nil {
			return err
		}
		if di.ModTime().Before(t) {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
		return nil
	})
	n.Speech(nil, "old saves cleaned")
}
