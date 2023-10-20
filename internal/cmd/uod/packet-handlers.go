package uod

import (
	"log"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/internal/commands"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// These functions handle client packets within the world process with direct
// access to the memory model.
func init() {
	packetHandlers.Add(0x02, handleWalkRequest)
	packetHandlers.Add(0x06, handleDoubleClickRequest)
	packetHandlers.Add(0x07, handleLiftRequest)
	packetHandlers.Add(0x08, handleDropRequest)
	packetHandlers.Add(0x09, handleSingleClickRequest)
	packetHandlers.Add(0x13, handleWearItemRequest)
	packetHandlers.Add(0x34, handleStatusRequest)
	packetHandlers.Add(0x3B, handleBuyRequest)
	packetHandlers.Add(0x6C, handleTargetResponse)
	packetHandlers.Add(0x73, handlePing)
	packetHandlers.Add(0x98, handleNameRequest)
	packetHandlers.Add(0x9F, handleSellRequest)
	packetHandlers.Add(0xAD, handleSpeech)
	packetHandlers.Add(0xB1, handleGUMPReply)
	packetHandlers.Add(0xBD, handleVersion)
	packetHandlers.Add(0xBF, handleGeneralInformation)
	packetHandlers.Add(0xC8, handleViewRange)
	packetHandlers.Add(0xD6, handleOPLCacheMiss)
}

// Registry of packet handler functions
var packetHandlers = util.NewRegistry[byte, func(*NetState, clientpacket.Packet)]("packet-handlers")

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
		world.Map().SendSpeech(n.m, uo.SpeechWhisperRange, p.Text)
	case uo.SpeechTypeNormal:
		if p.Text[0] == '[' {
			// Server command request
			cl := ""
			if len(p.Text) > 1 {
				cl = p.Text[1:]
			}
			commands.Execute(n, cl)
		} else {
			// Normal speech request
			world.Map().SendSpeech(n.m, uo.SpeechNormalRange, p.Text)
		}
	case uo.SpeechTypeEmote:
		world.Map().SendSpeech(n.m, uo.SpeechEmoteRange, p.Text)
	case uo.SpeechTypeYell:
		world.Map().SendSpeech(n.m, uo.SpeechYellRange, p.Text)
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
	n.Speech(o, o.DisplayName())
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
		game.DynamicDispatch("DoubleClick", o, n.m, nil)
		return
	}
	if !n.m.CanAccess(o) {
		return
	}
	game.DynamicDispatch("DoubleClick", o, n.m, nil)
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
		item.SetDropLocation(p.Location)
		item.SetLocation(newLocation)
		if !game.DynamicDispatch("Drop", target, n.m, item) {
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

func handleGUMPReply(n *NetState, cp clientpacket.Packet) {
	if n == nil {
		return
	}
	p := cp.(*clientpacket.GUMPReply)
	n.GUMPReply(p.GUMPSerial, p)
}

func handleBuyRequest(n *NetState, cp clientpacket.Packet) {
	// Sanity checks
	if n == nil || n.m == nil {
		return
	}
	p := cp.(*clientpacket.BuyItems)
	vendor := game.Find[game.Mobile](p.Vendor)
	if vendor == nil || vendor.Location().XYDistance(n.m.Location()) > uo.MaxViewRange {
		return
	}
	// Calculate total cost
	total := 0
	for _, bi := range p.BoughtItems {
		i := game.Find[game.Item](bi.Item)
		// Sanity checks
		if i == nil || game.RootParent(i).Serial() != p.Vendor {
			return
		}
		total += bi.Amount * i.Value()
	}
	// Charge gold
	// TODO support bank gold, change cliloc to 1042556
	if total > n.m.Gold() {
		n.Cliloc(vendor, 1019022) // You do not have enough gold.
		return
	}
	n.m.RemoveGold(total)
	// Give items
	for _, bi := range p.BoughtItems {
		i := game.Find[game.Item](bi.Item)
		if i == nil {
			continue
		}
		// Dirty hack for stablemaster NPCs
		tn := i.TemplateName()
		if strings.HasPrefix(tn, "StablemasterPlaceholder") {
			tn = strings.TrimPrefix(tn, "StablemasterPlaceholder")
			m := template.Create[game.Mobile](tn)
			if m == nil {
				// Something very wrong
				continue
			}
			m.SetLocation(n.m.Location())
			m.SetControlMaster(n.m)
			// TODO ownership
			world.Map().AddObject(m)
		} else {
			ni := template.Create[game.Item](tn)
			if ni == nil {
				// Something very wrong
				continue
			}
			ni.SetAmount(bi.Amount)
			ni.SetDropLocation(uo.RandomContainerLocation)
			if !n.m.DropToBackpack(ni, false) {
				n.m.DropToFeet(ni)
			}
		}
	}
	world.Map().SendCliloc(vendor, uo.SpeechNormalRange, 1080013, strconv.Itoa(total)) // The total of thy purchase is ~1_VAL~ gold,
}

func handleSellRequest(n *NetState, cp clientpacket.Packet) {
	// Sanity checks
	if n == nil || n.m == nil {
		return
	}
	p := cp.(*clientpacket.SellResponse)
	vm := game.Find[game.Mobile](p.Vendor)
	if vm == nil || vm.Location().XYDistance(n.m.Location()) > uo.MaxViewRange {
		return
	}
	// Remove items and calculate total
	total := 0
	for _, si := range p.SellItems {
		i := game.Find[game.Item](si.Serial)
		if i == nil {
			continue
		}
		rp := game.RootParent(i)
		if rp == nil || rp.Serial() != n.m.Serial() {
			continue
		}
		sa := si.Amount
		if sa > i.Amount() {
			sa = i.Amount()
		} else if sa < 1 {
			sa = 1
		}
		total += (i.Value() / 2) * sa
		if sa == i.Amount() {
			game.Remove(i)
		} else {
			i.SetAmount(i.Amount() - sa)
			world.Update(i)
		}
	}
	// Payment
	// TODO check support
	gc := template.Create[game.Item]("GoldCoin")
	gc.SetAmount(total)
	if !n.m.DropToBackpack(gc, false) {
		n.m.DropToFeet(gc)
	}
}

func handleNameRequest(n *NetState, cp clientpacket.Packet) {
	if n == nil {
		return
	}
	p := cp.(*clientpacket.NameRequest)
	o := world.Find(p.Serial)
	if o == nil {
		return
	}
	n.Send(&serverpacket.NameResponse{
		Serial: p.Serial,
		Name:   o.DisplayName(),
	})
}

func handleOPLCacheMiss(n *NetState, cp clientpacket.Packet) {
	if n == nil {
		return
	}
	p := cp.(*clientpacket.OPLCacheMiss)
	for _, s := range p.Serials {
		o := world.Find(s)
		if o == nil {
			continue
		}
		opl, _ := o.OPLPackets(o)
		n.Send(opl)
	}
}
