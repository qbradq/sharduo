package uod

import (
	"log"

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
	worldHandlers.Add(0x34, handleStatusRequest)
	worldHandlers.Add(0x6C, handleTargetResponse)
	worldHandlers.Add(0x73, handlePing)
	worldHandlers.Add(0xAD, handleSpeech)
	worldHandlers.Add(0xBD, handleVersion)
	worldHandlers.Add(0xC8, handleViewRange)
}

// Registry of packet handler functions
var worldHandlers = util.NewRegistry[byte, func(*NetState, clientpacket.Packet)]("world-handlers")

func handlePing(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Ping)
	n.Send(&serverpacket.Ping{
		Key: p.Key,
	})
}

func handleSpeech(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Speech)
	if len(p.Text) == 0 || n.m == nil {
		return
	}
	switch p.Type {
	case uo.SpeechTypeWhisper:
		for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.SpeechWhisperRange) {
			mob.NetState().Speech(n.m, p.Text)
		}
	case uo.SpeechTypeNormal:
		if p.Text[0] == '[' {
			// Server command request
			cl := ""
			if len(p.Text) > 1 {
				cl = p.Text[1:]
			}
			ExecuteCommand(n, cl)
		} else {
			// Normal speech request
			for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.SpeechNormalRange) {
				if n.m.Location().XYDistance(mob.Location()) <= mob.ViewRange() {
					mob.NetState().Speech(n.m, p.Text)
				}
			}
		}
	case uo.SpeechTypeEmote:
		for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.SpeechEmoteRange) {
			if n.m.Location().XYDistance(mob.Location()) <= mob.ViewRange() {
				mob.NetState().Speech(n.m, p.Text)
			}
		}
	case uo.SpeechTypeYell:
		for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.SpeechYellRange) {
			if n.m.Location().XYDistance(mob.Location()) <= mob.ViewRange() {
				mob.NetState().Speech(n.m, p.Text)
			}
		}
	}
}

func handleVersion(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.Version)
	if p.String != "7.0.15.1" {
		log.Printf("error: bad client version %s: disconnecting client", p.String)
		n.Disconnect()
	}
}

func handleTargetResponse(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.TargetResponse)
	n.TargetResponse(p)
}

func handleStatusRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.PlayerStatusRequest)
	switch p.StatusRequestType {
	case uo.StatusRequestTypeBasic:
		n.UpdateObject(n.m)
	case uo.StatusRequestTypeSkills:
		n.SendAllSkills()
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
		n.Send(&serverpacket.MoveReject{
			Sequence: byte(p.Sequence),
			Location: n.m.Location(),
			Facing:   n.m.Facing(),
		})
	}
}

func handleSingleClickRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.SingleClick)
	o := world.Find(p.Object)
	if o == nil {
		return
	}
	o.SingleClick(n.m)
}

func handleDoubleClickRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.DoubleClick)
	if p.WantPaperDoll {
		// This is an explicit request for our own paper doll, just send it
		if n.m != nil {
			n.OpenPaperDoll(n.m)
		}
		return
	}
	o := world.Find(p.Object.StripSelfFlag())
	if o == nil {
		return
	}
	// If this is a mobile we can skip a lot of checks
	if o.Serial().IsMobile() {
		// Range check just to make sure the player can actually see this thing
		// on-screen
		targetLocation := game.RootParent(o).Location()
		if n.m.Location().XYDistance(targetLocation) > n.m.ViewRange() {
			return
		}
		game.DynamicDispatch("OnDoubleClick", o, n.m)
		return
	}
	if !n.m.CanAccess(o) {
		return
	}
	// Range check
	targetLocation := game.RootParent(o).Location()
	if n.m.Location().XYDistance(targetLocation) > uo.MaxUseRange {
		return
	}
	// TODO Line of sight check
	game.DynamicDispatch("OnDoubleClick", o, n.m)
}

func handleViewRange(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.ViewRange)
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
	if !item.Movable() {
		n.DropReject(uo.MoveItemRejectReasonCannotLift)
		return
	}
	if n.m.Location().XYDistance(game.RootParent(item).Location()) > uo.MaxLiftRange {
		n.DropReject(uo.MoveItemRejectReasonOutOfRange)
		return
	}
	if !n.m.CanAccess(item) {
		n.DropReject(uo.MoveItemRejectReasonBelongsToAnother)
		return
	}
	// TODO Line of sight check
	item.Split(p.Amount)
	if !n.m.PickUp(item) {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
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
	// Do not trust the serial coming from the client, only drop what we are
	// holding.
	item := n.m.ItemInCursor()
	n.m.RequestCursorState(game.CursorStateDrop)
	if p.Container == uo.SerialSystem {
		// Drop to map request
		newLocation := p.Location
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
			n.m.PickUp(nil)
			n.Send(&serverpacket.DropApproved{})
			for _, mob := range world.Map().GetNetStatesInRange(n.m.Location(), uo.MaxViewRange) {
				mob.NetState().DragItem(item, n.m, n.m.Location(), nil, newLocation)
			}
		}
	} else {
		target := world.Find(p.Container)
		if target == nil {
			n.m.DropItemInCursor()
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
		}
		newLocation := game.RootParent(target).Location()
		if n.m.Location().XYDistance(newLocation) > uo.MaxDropRange {
			n.m.DropItemInCursor()
			n.DropReject(uo.MoveItemRejectReasonOutOfRange)
			return
		}
		if !n.m.CanAccess(target) {
			n.m.DropItemInCursor()
			n.DropReject(uo.MoveItemRejectReasonOutOfRange)
			return
		}
		// TODO Line of sight check
		if !target.DropObject(item, p.Location, n.m) {
			n.m.PickUp(nil)
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
			return
		}
		n.m.PickUp(nil)
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
		n.m.DropItemInCursor()
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if item.Serial() != n.m.ItemInCursor().Serial() {
		n.m.DropItemInCursor()
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	wearable, ok := item.(game.Wearable)
	if !ok {
		n.m.DropItemInCursor()
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	// TODO Check if we are allowed to equip items to this mobile
	n.m.RequestCursorState(game.CursorStateEquip)
	if !n.m.Equip(wearable) {
		n.m.PickUp(nil)
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	} else {
		n.Send(&serverpacket.DropApproved{})
	}
	n.m.PickUp(nil)
	n.Send(&serverpacket.DropApproved{})
}
