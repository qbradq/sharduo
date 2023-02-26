package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("SmeltOre", SmeltOre)
}

var forgeItemSet = map[uo.Graphic]struct{}{
	0x0FB1: {},
	0x197A: {},
	0x197E: {},
	0x1982: {},
	0x1986: {},
	0x198A: {},
	0x198E: {},
	0x1992: {},
	0x1996: {},
	0x199A: {},
	0x199E: {},
	0x19A2: {},
	0x19A6: {},
}

func SmeltOre(receiver, source game.Object, v any) {
	smelter, ok := source.(game.Mobile)
	if !ok || smelter.NetState() == nil {
		// Something is very wrong
		return
	}
	// The only scenario in which we would reject the request is if the ore
	// belongs to a mobile other than the smelter.
	root := game.RootParent(receiver)
	if root.Serial().IsMobile() && root.Serial() != receiver.Serial() {
		if smelter.NetState() != nil {
			smelter.NetState().Speech(nil, "You cannot access that.")
		}
		return
	}
	if !game.GetWorld().Map().Query(source.Location(), 3, forgeItemSet) {
		if smelter.NetState() != nil {
			smelter.NetState().Speech(nil, "There is no forge nearby.")
		}
		return
	}
	ore, ok := receiver.(game.Item)
	if !ok {
		// Something is very wrong
		return
	}
	if !smelter.SkillCheck(uo.SkillMining, 0, 750) {
		// Skill check failed, burn some ore
		smelter.NetState().Cliloc(nil, 501989)
		ore.Consume(1)
		return
	}
	ingot := template.Create("IronIngot").(game.Item)
	ingot.SetAmount(ore.Amount() * 2)
	game.GetWorld().Remove(ore)
	if !smelter.DropToBackpack(ingot, false) {
		smelter.DropToFeet(ingot)
	}
	smelter.NetState().Cliloc(nil, 501988)
}
