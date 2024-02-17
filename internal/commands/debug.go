package commands

import (
	"fmt"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// This file just houses the multi-purpose debug command

func init() {
	reg(&cmDesc{"debug", nil, commandDebug, game.RoleDeveloper, "debug command [arguments]", "Executes debug commands"})
}

func commandDebug(n game.NetState, args CommandArgs, cl string) {
	if n == nil || n.Mobile() == nil {
		return
	}
	// Sanitize input
	if len(args) < 2 {
		n.Speech(nil, "debug command requires a command name: [debug command_name")
		return
	}
	switch args[1] {
	case "hue_field":
		fallthrough
	case "vendor_bag":
		fallthrough
	case "welcome":
		fallthrough
	case "gfx_test":
		fallthrough
	case "force_chunk_update":
		fallthrough
	case "memory_test":
		fallthrough
	case "delay_test":
		fallthrough
	case "mount":
		fallthrough
	case "panic":
		fallthrough
	case "shirt_bag":
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
	case "hue_field":
		var hue uo.Hue = 1
		for iy := 0; iy < 10; iy++ {
			for ix := 0; ix < 300; ix++ {
				r := game.NewItem("HueSelector")
				if r == nil {
					continue
				}
				r.Hue = hue
				hue++
				if hue >= 3000 {
					hue -= 3000
				}
				l := n.Mobile().Location
				l.X += ix
				l.Y += iy * 2
				r.Location = l
				game.World.Map().AddItem(r, false)
			}
		}
	case "vendor_bag":
		n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
			m, found := game.World.FindMobile(tr.TargetObject)
			if !found {
				return
			}
			w := m.Equipment[uo.LayerNPCBuyRestockContainer]
			if w == nil {
				return
			}
			if !w.HasFlags(game.ItemFlagsContainer) {
				return
			}
			w.Open(n.Mobile())
		})
	case "welcome":
		n.GUMP(gumps.New("welcome"), n.Mobile().Serial, 0)
	case "gfx_test":
		n.Send(&serverpacket.GraphicalEffect{
			GFXType:        uo.GFXTypeFixed,
			Source:         n.Mobile().Serial,
			Target:         n.Mobile().Serial,
			Graphic:        0x3728,
			SourceLocation: n.Mobile().Location,
			TargetLocation: n.Mobile().Location,
			Speed:          15,
			Duration:       10,
			Fixed:          true,
			Explodes:       false,
			Hue:            uo.HueDefault,
			GFXBlendMode:   uo.GFXBlendModeNormal,
		})
	case "force_chunk_update":
		n.Speech(n.Mobile(), "Target the chunk you wish to force-update")
		n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
			game.World.Map().GetChunk(tr.Location).Update(game.World.Time())
		})
	case "global_light":
		ll := uo.LightLevel(args.Int(2))
		n.Send(&serverpacket.GlobalLightLevel{
			LightLevel: ll,
		})
	case "personal_light":
		ll := uo.LightLevel(args.Int(2))
		n.Send(&serverpacket.PersonalLightLevel{
			Serial:     n.Mobile().Serial,
			LightLevel: ll,
		})
	case "animate":
		at := uo.AnimationType(args.Int(2))
		aa := uo.AnimationAction(args.Int(2))
		n.Animate(n.Mobile(), at, aa)
	case "music":
		which := uo.Music(args.Int(2))
		n.Music(which)
	case "sound":
		which := uo.Sound(args.Int(2))
		game.World.Map().PlaySound(which, n.Mobile().Location)
	case "memory_test":
		start := time.Now()
		for i := 0; i < 1_000_000; i++ {
			item := game.NewItem("FancyShirt")
			if item == nil {
				continue
			}
			nl := uo.Point{
				X: util.Random(100, uo.MapWidth-101),
				Y: util.Random(100, uo.MapHeight-101),
				Z: uo.MapMaxZ,
			}
			f, _ := game.World.Map().GetFloorAndCeiling(nl, false, false)
			if f != nil {
				nl.Z = f.Z()
			}
			item.Location = nl
			game.World.Map().AddItem(item, true)
		}
		for i := 0; i < 150_000; i++ {
			m := game.NewMobile("Banker")
			if m == nil {
				continue
			}
			nl := uo.Point{
				X: util.Random(100, uo.MapWidth-101),
				Y: util.Random(100, uo.MapHeight-101),
				Z: uo.MapMaxZ,
			}
			f, _ := game.World.Map().GetFloorAndCeiling(nl, false, false)
			if f != nil {
				nl.Z = f.Z()
			}
			m.Location = nl
			game.World.Map().AddMobile(m, true)
		}
		end := time.Now()
		n.Speech(n.Mobile(), fmt.Sprintf("operation completed in %s", end.Sub(start)))
	case "delay_test":
		game.NewTimer(uo.DurationSecond*5, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*10, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*15, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*20, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*25, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*30, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*35, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*40, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*45, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*50, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*55, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*60, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
		game.NewTimer(uo.DurationSecond*65, "WhisperTime", n.Mobile(), n.Mobile(), false, nil)
	case "mount":
		llama := game.NewMobile("Llama")
		llama.Location = n.Mobile().Location
		llama.Hue = 0x76 // Llama Energy Vortex hue
		llama.ControlMaster = n.Mobile()
		llama.SetAI("Follow", n.Mobile())
		game.World.Map().AddMobile(llama, true)
		game.ExecuteEventHandler("Mount", llama, n.Mobile(), nil)
	case "shirt_bag":
		backpack := game.NewItem("Backpack")
		backpack.Location = n.Mobile().Location
		game.World.Map().AddItem(backpack, true)
		for i := 0; i < 125; i++ {
			shirt := game.NewItem("FancyShirt")
			backpack.DropInto(shirt, true)
		}
	case "splat":
		start := time.Now()
		count := 0
		for iy := n.Mobile().Location.Y - 50; iy < n.Mobile().Location.Y+50; iy++ {
			for ix := n.Mobile().Location.X - 50; ix < n.Mobile().Location.X+50; ix++ {
				if m := game.NewMobile(args[2]); m != nil {
					m.Location = uo.Point{X: ix, Y: iy, Z: n.Mobile().Location.Z}
					game.World.Map().AddMobile(m, true)
				} else if i := game.NewItem(args[2]); i != nil {
					i.Location = uo.Point{X: ix, Y: iy, Z: n.Mobile().Location.Z}
					game.World.Map().AddItem(i, true)
				} else {
					n.Speech(n.Mobile(), "debug splat failed to create object with template %s", args[2])
					return
				}
				count++
			}
		}
		end := time.Now()
		n.Speech(n.Mobile(), "generated %d items in %d milliseconds\n", count, end.Sub(start).Milliseconds())
	case "panic":
		// Force a nil reference panic
		var m game.Mobile
		m.AdjustWeight(1)
	}
}
