package commands

import (
	"fmt"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

// This file just houses the multi-purpose debug command

func init() {
	regcmd(&cmdesc{"debug", nil, commandDebug, game.RoleDeveloper, "debug command [arguments]", "Executes debug commands"})
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
	case "hue_field":
		hue := 1
		for iy := 0; iy < 10; iy++ {
			for ix := 0; ix < 300; ix++ {
				r := template.Create[game.Item]("HueSelector")
				if r == nil {
					continue
				}
				r.SetHue(uo.Hue(hue))
				hue++
				if hue >= 3000 {
					hue -= 3000
				}
				l := n.Mobile().Location()
				l.X += int16(ix)
				l.Y += int16(iy * 2)
				r.SetLocation(l)
				game.GetWorld().Map().SetNewParent(r, nil)
			}
		}
	case "vendor_bag":
		n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
			m := game.Find[game.Mobile](tr.TargetObject)
			if m == nil {
				return
			}
			w := m.EquipmentInSlot(uo.LayerNPCBuyRestockContainer)
			if w == nil {
				return
			}
			c, ok := w.(game.Container)
			if !ok {
				return
			}
			c.Open(n.Mobile())
		})
	case "welcome":
		n.GUMP(gumps.New("welcome"), n.Mobile(), nil)
	case "gfx_test":
		n.Send(&serverpacket.GraphicalEffect{
			GFXType:        uo.GFXTypeFixed,
			Source:         n.Mobile().Serial(),
			Target:         n.Mobile().Serial(),
			Graphic:        0x3728,
			SourceLocation: n.Mobile().Location(),
			TargetLocation: n.Mobile().Location(),
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
			game.GetWorld().Map().GetChunk(tr.Location).Update(game.GetWorld().Time())
		})
	case "global_light":
		ll := uo.LightLevel(args.Int(2))
		n.Send(&serverpacket.GlobalLightLevel{
			LightLevel: ll,
		})
	case "personal_light":
		ll := uo.LightLevel(args.Int(2))
		n.Send(&serverpacket.PersonalLightLevel{
			Serial:     n.Mobile().Serial(),
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
		game.GetWorld().Map().PlaySound(which, n.Mobile().Location())
	case "memory_test":
		start := time.Now()
		for i := 0; i < 1_000_000; i++ {
			o := template.Create[game.Object]("FancyShirt")
			if o == nil {
				continue
			}
			nl := uo.Location{
				X: int16(game.GetWorld().Random().Random(100, uo.MapWidth-101)),
				Y: int16(game.GetWorld().Random().Random(100, uo.MapHeight-101)),
				Z: uo.MapMaxZ,
			}
			f, _ := game.GetWorld().Map().GetFloorAndCeiling(nl, false, false)
			if f != nil {
				nl.Z = f.Z()
			}
			o.SetLocation(nl)
			game.GetWorld().Map().ForceAddObject(o)
		}
		for i := 0; i < 150_000; i++ {
			o := template.Create[game.Object]("Banker")
			if o == nil {
				continue
			}
			nl := uo.Location{
				X: int16(game.GetWorld().Random().Random(100, uo.MapWidth-101)),
				Y: int16(game.GetWorld().Random().Random(100, uo.MapHeight-101)),
				Z: uo.MapMaxZ,
			}
			f, _ := game.GetWorld().Map().GetFloorAndCeiling(nl, false, false)
			if f != nil {
				nl.Z = f.Z()
			}
			o.SetLocation(nl)
			game.GetWorld().Map().ForceAddObject(o)
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
		llama := template.Create[game.Mobile]("Llama")
		if llama == nil {
			break
		}
		llama.SetLocation(n.Mobile().Location())
		llama.SetHue(0x76) // Llama Energy Vortex hue
		llama.SetControlMaster(n.Mobile())
		llama.SetAI("Follow")
		llama.SetAIGoal(n.Mobile())
		game.GetWorld().Map().AddObject(llama)
		game.ExecuteEventHandler("Mount", llama, n.Mobile(), nil)
	case "shirtbag":
		backpack := template.Create[game.Container]("Backpack")
		if backpack == nil {
			break
		}
		backpack.SetLocation(n.Mobile().Location())
		game.GetWorld().Map().SetNewParent(backpack, nil)
		for i := 0; i < 125; i++ {
			shirt := template.Create[game.Object]("FancyShirt")
			if shirt == nil {
				continue
			}
			shirt.SetLocation(uo.RandomContainerLocation)
			if !game.GetWorld().Map().SetNewParent(shirt, backpack) {
				n.Speech(n.Mobile(), "failed to add an item to the backpack")
				break
			}
		}
	case "splat":
		start := time.Now()
		count := 0
		for iy := n.Mobile().Location().Y - 50; iy < n.Mobile().Location().Y+50; iy++ {
			for ix := n.Mobile().Location().X - 50; ix < n.Mobile().Location().X+50; ix++ {
				o := template.Create[game.Object](args[2])
				if o == nil {
					n.Speech(n.Mobile(), "debug splat failed to create object with template %s", args[2])
				}
				o.SetLocation(uo.Location{
					X: ix,
					Y: iy,
					Z: n.Mobile().Location().Z,
				})
				game.GetWorld().Map().SetNewParent(o, nil)
				count++
			}
		}
		end := time.Now()
		n.Speech(n.Mobile(), "generated %d items in %d milliseconds\n", count, end.Sub(start).Milliseconds())
	case "panic":
		// Force a nil reference panic
		var m game.Mobile
		m.AdjustGold(1)
	}
}
