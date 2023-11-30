package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("UseDoor", UseDoor)
}

var doorOpenSounds = map[string]uo.Sound{
	"MetalDoor":        0xEC,
	"BarredMetalDoor":  0xEC,
	"RattanDoor":       0xEB,
	"WoodenDoor":       0xEA,
	"LightWoodenDoor":  0xEA,
	"StrongWoodenDoor": 0xEA,
	"IronGateShort":    0xEC,
	"IronGate":         0xEC,
	"LightWoodenGate":  0xEB,
	"DarkWoodenGate":   0xEB,
}

var doorCloseSounds = map[string]uo.Sound{
	"MetalDoor":        0xF3,
	"BarredMetalDoor":  0xF3,
	"RattanDoor":       0xF2,
	"WoodenDoor":       0xF1,
	"LightWoodenDoor":  0xF1,
	"StrongWoodenDoor": 0xF1,
	"IronGateShort":    0xF3,
	"IronGate":         0xF3,
	"LightWoodenGate":  0xF2,
	"DarkWoodenGate":   0xF2,
}

var doorCloseTimers = map[uo.Serial]uo.Serial{}

func doUseDoor(receiver, source game.Object, force bool) bool {
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	// Range check
	if !force && receiver.Location().XYDistance(source.Location()) > uo.MaxUseRange {
		sm.NetState().Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// TODO Line of sight check
	// Select door offsets, sounds, ect
	ri, ok := receiver.(game.Item)
	if !ok {
		return false
	}
	l := ri.Location()
	ofs := uo.DoorOffsets[int(ri.Facing())]
	var s uo.Sound
	if ri.Flipped() {
		l.X -= ofs.X
		l.Y -= ofs.Y
		s = doorCloseSounds[ri.TemplateName()]
	} else {
		l.X += ofs.X
		l.Y += ofs.Y
		s = doorOpenSounds[ri.TemplateName()]
	}
	// Skip map fit check, not all doors on the retail map have clearance with
	// the terrain to properly open
	game.GetWorld().Map().ForceRemoveObject(ri)
	ri.Flip()
	ri.SetLocation(l)
	ri.SetDefForGraphic(ri.Graphic())
	game.GetWorld().Map().ForceAddObject(ri)
	game.GetWorld().Map().PlaySound(s, l)
	// Auto-close functionality
	if ri.Flipped() {
		// Door is now open, setup a timer to close it
		doorCloseTimers[ri.Serial()] = game.NewTimer(uo.DurationSecond*20, "UseDoor", ri, source, true, true)
	} else {
		// Door is now closed, make sure there are no pending timers
		s, found := doorCloseTimers[ri.Serial()]
		if found {
			game.CancelTimer(s)
			delete(doorCloseTimers, ri.Serial())
		}
	}
	return true
}

func UseDoor(receiver, source game.Object, v any) bool {
	force, ok := v.(bool)
	if !ok {
		force = false
	}
	doors := game.GetWorld().Map().ItemBaseQuery("BaseDoor", receiver.Location().BoundsByRadius(1))
	for _, d := range doors {
		if !doUseDoor(d, source, force) {
			return false
		}
	}
	return true
}
