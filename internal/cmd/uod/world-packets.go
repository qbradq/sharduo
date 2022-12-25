package uod

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// These functions handle client packets within the world process with direct
// access to the memory model.
func init() {
	worldHandlers.Add(0x02, handleWalkRequest)
	worldHandlers.Add(0x06, handleDoubleClickRequest)
	worldHandlers.Add(0x07, handleLiftRequest)
	worldHandlers.Add(0x08, handleDropRequest)
	worldHandlers.Add(0x09, handleSingleClickRequest)
	worldHandlers.Add(0x13, handleWearItemRequest)
	worldHandlers.Add(0x6C, handleTargetResponse)
	worldHandlers.Add(0x34, handleStatusRequest)
	worldHandlers.Add(0xC8, handleClientViewRange)
}

// Registry of packet handler functions
var worldHandlers = util.NewRegistry[uo.Serial, func(*NetState, clientpacket.Packet)]("world-handlers")

func handleTargetResponse(n *NetState, cp clientpacket.Packet) {
	world.ExecuteTarget(cp.(*clientpacket.TargetResponse))
}

func handleStatusRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.PlayerStatusRequest)
	switch p.StatusRequestType {
	case uo.StatusRequestTypeBasic:
		n.Send(&serverpacket.StatusBarInfo{
			Mobile:         n.m.Serial(),
			Name:           n.m.DisplayName(),
			Female:         n.m.IsFemale(),
			HP:             n.m.HitPoints(),
			MaxHP:          n.m.MaxHitPoints(),
			NameChangeFlag: false,
			Strength:       n.m.Strength(),
			Dexterity:      n.m.Dexterity(),
			Intelligence:   n.m.Intelligence(),
			Stamina:        n.m.Stamina(),
			MaxStamina:     n.m.MaxStamina(),
			Mana:           n.m.Mana(),
			MaxMana:        n.m.MaxMana(),
			Gold:           0,
			ArmorRating:    0,
			Weight:         0,
			StatsCap:       uo.StatsCapDefault,
			Followers:      0,
			MaxFollowers:   uo.MaxFollowers,
		})
	case uo.StatusRequestTypeSkills:
		// TODO Respond to skills request
	}
}

func handleWalkRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.WalkRequest)
	if n.m == nil {
		return
	}
	n.m.SetRunning(p.IsRunning)
	if world.Map().MoveMobile(n.m, p.Direction) {
		n.Send(&serverpacket.MoveAcknowledge{
			Sequence:  p.Sequence,
			Notoriety: uo.NotorietyInnocent,
		})
	} else {
		// TODO reject movement packet
	}
}

func handleSingleClickRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.SingleClick)
	o := world.Find(p.ID)
	if o != nil {
		// TODO Line of sight check
		o.SingleClick(n.m)
	}
}

func handleDoubleClickRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.DoubleClick)
	if p.IsSelf || p.ID == n.m.Serial() {
		// This is a double-click on the player's mobile. Ignore the rest of
		// the serial and directly access the mobile. NEVER trust a random
		// serial from the client :)
		n.Send(&serverpacket.OpenPaperDoll{
			Serial:    n.m.Serial(),
			Text:      n.m.DisplayName(),
			WarMode:   false,
			Alterable: true,
		})
	} else {
		o := world.Find(p.ID)
		if o == nil {
			return
		}
		// TODO Range check
		// TODO Line of sight check
		o.DoubleClick(n.m)
	}
}

func handleClientViewRange(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.ClientViewRange)
	world.Map().UpdateViewRangeForMobile(n.m, p.Range)
	n.Send(&serverpacket.ClientViewRange{
		Range: byte(n.m.ViewRange()),
	})
}

func handleLiftRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if n.m.IsItemOnCursor() {
		n.m.DropItemInCursor()
		n.DropReject(uo.MoveItemRejectReasonAlreadyHoldingItem)
		return
	}
	p := cp.(*clientpacket.LiftRequest)
	o := world.Find(p.Item)
	if o == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	item, ok := o.(game.Item)
	if !ok {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if n.m.Location().XYDistance(item.RootParent().Location()) > uo.MaxLiftRange {
		n.DropReject(uo.MoveItemRejectReasonOutOfRange)
		return
	}
	// TODO Line of sight check
	if !n.m.SetItemInCursor(item) {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
}

func handleDropRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if !n.m.IsItemOnCursor() {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	p := cp.(*clientpacket.DropRequest)
	if p.Item != n.m.ItemInCursor().Serial() {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	itemObj := world.Find(p.Item)
	if itemObj == nil {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	item, ok := itemObj.(game.Item)
	if !ok {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if p.Container == uo.SerialSystem {
		// Drop to map request
		newLocation := uo.Location{X: p.X, Y: p.Y, Z: p.Z}
		if n.m.Location().XYDistance(newLocation) > uo.MaxDropRange {
			n.m.DropItemInCursor()
			n.DropReject(uo.MoveItemRejectReasonOutOfRange)
			return
		}
		// TODO Line of sight check
		item.SetLocation(newLocation)
		if !world.Map().SetNewParent(item, nil) {
			n.m.DropItemInCursor()
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
			return
		} else {
			n.m.SetItemInCursor(nil)
			n.Send(&serverpacket.DropApproved{})
			for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.MaxViewRange) {
				mob.NetState().DragItem(item, n.m, n.m.Location(), nil, newLocation)
			}
		}
	} else {
		target := world.Find(p.Container)
		if target == nil {
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
		}
		newLocation := target.RootParent().Location()
		if n.m.Location().XYDistance(newLocation) > uo.MaxDropRange {
			n.m.DropItemInCursor()
			n.DropReject(uo.MoveItemRejectReasonOutOfRange)
			return
		}
		// TODO Line of sight check
		item.SetLocation(uo.Location{
			X: p.X,
			Y: p.Y,
		})
		if !target.DropObject(item, n.m) {
			n.m.SetItemInCursor(nil)
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
			return
		}
		n.m.SetItemInCursor(nil)
		n.Send(&serverpacket.DropApproved{})
		for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.MaxViewRange) {
			mob.NetState().DragItem(item, n.m, n.m.Location(), nil, newLocation)
		}
	}
}

func handleWearItemRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	p := cp.(*clientpacket.WearItemRequest)
	item := world.Find(p.Item)
	wearer := world.Find(p.Wearer)
	if item == nil || wearer == nil {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if item != n.m.ItemInCursor() {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	wearable, ok := item.(game.Wearable)
	if !ok {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	// TODO Check if we are allowed to equip items to this mobile
	// This will remove the object from it's parent (the mobile's cursor) and
	// add it to the other mobile's equipment, or not.
	if !n.m.Equip(wearable) {
		n.m.SetItemInCursor(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	} else {
		n.Send(&serverpacket.DropApproved{})
	}
	n.m.SetItemInCursor(nil)
	n.Send(&serverpacket.DropApproved{})
}
