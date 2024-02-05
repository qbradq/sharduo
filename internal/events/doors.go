package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("UseDoor", useDoor)
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

func doUseDoor(ri *game.Item, sm *game.Mobile, force bool) bool {
	// Range check
	if !force && ri.Location.XYDistance(sm.Location) > uo.MaxUseRange {
		sm.NetState.Cliloc(nil, 502803) // It's too far away.
		return false
	}
	// Line of sight check
	if !force && !sm.HasLineOfSight(ri) {
		sm.NetState.Cliloc(nil, 500950) // You cannot see that.
		return false
	}
	// Select door offsets, sounds, ect
	l := ri.Location
	ofs := uo.DoorOffsets[int(ri.Facing)]
	var s uo.Sound
	if ri.Flipped {
		l.X -= ofs.X
		l.Y -= ofs.Y
		s = doorCloseSounds[ri.TemplateName]
	} else {
		l.X += ofs.X
		l.Y += ofs.Y
		s = doorOpenSounds[ri.TemplateName]
	}
	// Skip map fit check, not all doors on the retail map have clearance with
	// the terrain to properly open
	game.World.Map().RemoveItem(ri)
	ri.Flipped = !ri.Flipped
	ri.Location = l
	ri.Def = game.World.ItemDefinition(ri.CurrentGraphic())
	game.World.Map().AddItem(ri, true)
	game.World.Map().PlaySound(s, l)
	// Auto-close functionality
	if ri.Flipped {
		// Door is now open, setup a timer to close it
		doorCloseTimers[ri.Serial] = game.NewTimer(uo.DurationSecond*20, "UseDoor", ri, sm, true, true)
	} else {
		// Door is now closed, make sure there are no pending timers
		s, found := doorCloseTimers[ri.Serial]
		if found {
			game.CancelTimer(s)
			delete(doorCloseTimers, ri.Serial)
		}
	}
	return true
}

func useDoor(receiver, source, v any) bool {
	force, ok := v.(bool)
	if !ok {
		force = false
	}
	ri := receiver.(*game.Item)
	sm := source.(*game.Mobile)
	doors := game.World.Map().ItemBaseQuery("BaseDoor", ri.Location, 1)
	for _, d := range doors {
		if !doUseDoor(d, sm, force) {
			return false
		}
	}
	return true
}
