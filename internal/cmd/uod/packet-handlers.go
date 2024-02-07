package uod

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/qbradq/sharduo/internal/commands"
	"github.com/qbradq/sharduo/internal/configuration"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
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
	packetHandlers.Add(0x12, handleMacroRequest)
	packetHandlers.Add(0x13, handleWearItemRequest)
	packetHandlers.Add(0x34, handleStatusRequest)
	packetHandlers.Add(0x3B, handleBuyRequest)
	packetHandlers.Add(0x6C, handleTargetResponse)
	packetHandlers.Add(0x73, handlePing)
	packetHandlers.Add(0x75, handleRenameRequest)
	packetHandlers.Add(0x98, handleNameRequest)
	packetHandlers.Add(0x9F, handleSellRequest)
	packetHandlers.Add(0xAC, handleTextGUMPReply)
	packetHandlers.Add(0xAD, handleSpeech)
	packetHandlers.Add(0xB1, handleGUMPReply)
	packetHandlers.Add(0xBD, handleVersion)
	packetHandlers.Add(0xBF, handleGeneralInformation)
	packetHandlers.Add(0xC8, handleViewRange)
	packetHandlers.Add(0xD6, handleOPLCacheMiss)
	packetHandlers.Add(0xFE, handleCharacterLogout)
	packetHandlers.Add(0xFF, handleCharacterLogin)
}

// Registry of packet handler functions
var packetHandlers = util.NewRegistry[byte, func(*NetState, clientpacket.Packet)]("packet-handlers")

func handleCharacterLogin(n *NetState, cp clientpacket.Packet) {
	var player *game.Mobile
	p := cp.(*CharacterLogin)
	// Attempt to load the player
	if p.CharacterIndex >= len(n.account.Characters) {
		log.Printf("warning: account %s sent invalid character login packet",
			n.account.Username)
		n.Disconnect()
		return
	}
	ps := n.account.Characters[p.CharacterIndex]
	m := world.m.RetrieveObject(ps).(*game.Mobile)
	if m != nil {
		if m.NetState != nil {
			// Connecting to an already connected player, disconnect the
			// existing connection.
			m.NetState.Disconnect()
			m.NetState = nil
		}
		player = m
	} else {
		log.Printf("error: account %s character slot %d mobile not found",
			n.account.Username, p.CharacterIndex)
		n.Disconnect()
		return
	}
	// In case the player mobile was in deep storage we try to remove it
	world.m.RetrieveObject(player.Serial)
	world.m.RemoveMobile(m)
	world.m.AddMobile(m, true)
	world.UpdateMobile(player)
	n.m = player
	n.m.NetState = n
	n.m.Account = n.account
	Broadcast("Welcome %s to %s!", n.m.DisplayName(),
		configuration.GameServerName)
	// Send the EnterWorld packet
	facing := n.m.Facing
	if n.m.Running {
		facing = facing.SetRunningFlag()
	} else {
		facing = facing.StripRunningFlag()
	}
	n.Send(&serverpacket.EnterWorld{
		Player:   n.m.Serial,
		Body:     n.m.Body,
		Location: n.m.Location,
		Facing:   facing,
		Width:    uo.MapWidth,
		Height:   uo.MapHeight,
	})
	n.Send(&serverpacket.LoginComplete{})
	n.Send(&serverpacket.Time{
		Time: time.Now(),
	})
	world.m.SendEverything(n.m)
	n.SendMobile(n.m)
	n.GUMP(gumps.New("welcome"), n.m.Serial, 0)
}

func handleCharacterLogout(n *NetState, cp clientpacket.Packet) {
	game.NewTimer(uo.DurationMinute*10, "PlayerLogout", n.m, nil, false, nil)
}

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
		m := game.World.FindMobile(p.PlayerMobileID)
		n.UpdateMobile(m)
	case uo.StatusRequestTypeSkills:
		n.SendAllSkills()
	}
}

func handleWalkRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.WalkRequest)
	if n.m == nil {
		return
	}
	n.m.Running = p.Running
	if world.m.MoveMobile(n.m, p.Direction) {
		n.Send(&serverpacket.MoveAcknowledge{
			Sequence:  p.Sequence,
			Notoriety: uo.NotorietyInnocent,
		})
	} else {
		n.Send(&serverpacket.MoveReject{
			Sequence: byte(p.Sequence),
			Location: n.m.Location,
			Facing:   n.m.Facing,
		})
	}
}

func handleSingleClickRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.SingleClick)
	switch o := world.Find(p.Object).(type) {
	case *game.Mobile:
		n.Speech(o, o.DisplayName())
	case *game.Item:
		n.Speech(o, o.DisplayName())
	}
}

func handleDoubleClickRequest(n *NetState, cp clientpacket.Packet) {
	if !n.TakeAction() {
		n.Cliloc(nil, 500119) // You must wait to perform another action.
		return
	}
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
	switch o := world.Find(p.Object.StripSelfFlag()).(type) {
	case *game.Mobile:
		// Range check just to make sure the player can actually see this thing
		// on-screen
		if n.m.Location.XYDistance(o.Location) > n.m.ViewRange {
			return
		}
		o.ExecuteEvent("DoubleClick", n.m, nil)
	case *game.Item:
		if !n.m.CanAccess(o) {
			return
		}
		o.ExecuteEvent("DoubleClick", n.m, nil)
	}
}

func handleViewRange(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		return
	}
	p := cp.(*clientpacket.ViewRange)
	world.Map().UpdateViewRangeForMobile(n.m, p.Range)
	n.Send(&serverpacket.ClientViewRange{
		Range: byte(n.m.ViewRange),
	})
}

func handleLiftRequest(n *NetState, cp clientpacket.Packet) {
	if !n.TakeAction() {
		n.Cliloc(nil, 500119) // You must wait to perform another action.
		return
	}
	if n.m == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if n.m.Cursor != nil {
		n.m.DropToFeet(n.m.Cursor)
		n.m.Cursor = nil
		n.DropReject(uo.MoveItemRejectReasonAlreadyHoldingItem)
		return
	}
	p := cp.(*clientpacket.LiftRequest)
	item := world.FindItem(p.Item)
	if item == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if item.HasFlags(game.ItemFlagsFixed) {
		n.DropReject(uo.MoveItemRejectReasonCannotLift)
		return
	}
	if n.m.Location.XYDistance(game.MapLocation(item)) > uo.MaxLiftRange {
		n.DropReject(uo.MoveItemRejectReasonOutOfRange)
		return
	}
	if !n.m.CanAccess(item) {
		n.DropReject(uo.MoveItemRejectReasonBelongsToAnother)
		return
	}
	// Line of sight check
	if !n.m.HasLineOfSight(item) {
		n.DropReject(uo.MoveItemRejectReasonOutOfSight)
		return
	}
	item.Split(p.Amount)
	n.m.Cursor = item
	// Play lift sound
	n.Sound(item.LiftSound, game.MapLocation(item))
}

func handleDropRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if n.m.Cursor != nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	p := cp.(*clientpacket.DropRequest)
	// Do not trust the serial coming from the client, only drop what we are
	// holding.
	item := n.m.Cursor
	if p.Container == uo.SerialSystem {
		// Drop to map request
		newLocation := p.Location
		if n.m.Location.XYDistance(newLocation) > uo.MaxDropRange {
			n.m.DropToFeet(n.m.Cursor)
			n.m.Cursor = nil
			n.DropReject(uo.MoveItemRejectReasonOutOfRange)
			return
		}
		// Line of sight check
		oldLocation := item.Location
		item.Location = newLocation
		if !n.m.HasLineOfSight(item) {
			item.Location = oldLocation
			n.m.DropToFeet(n.m.Cursor)
			n.m.Cursor = nil
			n.DropReject(uo.MoveItemRejectReasonOutOfSight)
			return
		}
		if !world.Map().AddItem(item, false) {
			n.m.DropToFeet(n.m.Cursor)
			n.m.Cursor = nil
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
			return
		} else {
			n.m.Cursor = nil
			n.Send(&serverpacket.DropApproved{})
			n.Sound(item.GetDropSoundOverride(uo.SoundDefaultDrop), newLocation)
			// Distribute drag packets
			for _, mob := range world.Map().NetStatesInRange(n.m.Location, 0) {
				mob.NetState.DragItem(item, n.m, n.m.Location, nil, newLocation)
			}
		}
	} else {
		target := world.Find(p.Container)
		if target == nil {
			n.m.DropToFeet(n.m.Cursor)
			n.m.Cursor = nil
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
		}
		var nl uo.Point
		if item, ok := target.(*game.Item); ok {
			nl = game.MapLocation(item)
		} else {
			nl = target.(*game.Mobile).Location
		}
		item.Location = nl
		if !game.DynamicDispatch("Drop", target, n.m, item) {
			n.m.DropToFeet(n.m.Cursor)
			n.m.Cursor = nil
			n.DropReject(uo.MoveItemRejectReasonUnspecified)
			return
		}
		n.m.Cursor = nil
		n.Send(&serverpacket.DropApproved{})
		// Play drop sound
		if di, ok := target.(*game.Item); ok {
			if di.HasFlags(game.ItemFlagsContainer) {
				n.Sound(item.GetDropSoundOverride(di.DropSound), nl)
			} else {
				n.Sound(item.GetDropSoundOverride(uo.SoundDefaultDrop), nl)
			}
		} else {
			n.Sound(uo.SoundBagDrop, nl)
		}
		// Distribute drag packets
		for _, mob := range world.Map().NetStatesInRange(n.m.Location, 0) {
			mob.NetState.DragItem(item, n.m, n.m.Location, nil, nl)
		}
	}
}

func handleWearItemRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil {
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	p := cp.(*clientpacket.WearItemRequest)
	item := world.FindItem(p.Item)
	wearer := world.FindMobile(p.Wearer)
	if item == nil || wearer == nil {
		n.m.DropToFeet(n.m.Cursor)
		n.m.Cursor = nil
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	// Check if we are allowed to equip items to this mobile
	if wearer != n.m {
		n.m.DropToFeet(n.m.Cursor)
		n.m.Cursor = nil
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if item != n.m.Cursor {
		n.m.DropToFeet(n.m.Cursor)
		n.m.Cursor = nil
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if item.Wearable() {
		n.m.DropToFeet(n.m.Cursor)
		n.m.Cursor = nil
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	}
	if !n.m.Equip(item) {
		n.m.Cursor = nil
		n.DropReject(uo.MoveItemRejectReasonUnspecified)
		return
	} else {
		n.Send(&serverpacket.DropApproved{})
	}
	n.m.Cursor = nil
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
	vendor := world.FindMobile(p.Vendor)
	if vendor == nil || vendor.Location.XYDistance(n.m.Location) > uo.MaxViewRange {
		return
	}
	// Calculate total cost
	total := 0
	for _, bi := range p.BoughtItems {
		i := world.FindItem(bi.Item)
		// Sanity checks
		if i == nil || i.Wearer.Serial != p.Vendor {
			return
		}
		total += bi.Amount * i.Value
	}
	// Charge gold
	if !n.m.ChargeGold(total) {
		n.Cliloc(vendor, 1042556) // Thou dost not have enough gold, not even in thy bank account.
		return
	}
	// Give items
	for _, bi := range p.BoughtItems {
		i := world.FindItem(bi.Item)
		if i == nil {
			continue
		}
		// Dirty hack for stable master NPCs
		tn := i.TemplateName
		if strings.HasPrefix(tn, "StableMasterPlaceholder") {
			tn = strings.TrimPrefix(tn, "StableMasterPlaceholder")
			m := game.NewMobile(tn)
			if m == nil {
				// Something very wrong
				continue
			}
			m.Location = n.m.Location
			m.ControlMaster = n.m
			m.AI = "Follow"
			m.AIGoal = n.m
			world.Map().AddMobile(m, true)
		} else {
			ni := game.NewItem(tn)
			if ni == nil {
				// Something very wrong
				continue
			}
			ni.Amount = bi.Amount
			if !n.m.DropToBackpack(ni, false) {
				n.m.DropToFeet(ni)
			}
		}
	}
	world.Map().SendCliloc(vendor, 1080013, strconv.Itoa(total)) // The total of thy purchase is ~1_VAL~ gold,
	n.Sound(0x02E6, n.m.Location)
}

func handleSellRequest(n *NetState, cp clientpacket.Packet) {
	// Sanity checks
	if n == nil || n.m == nil {
		return
	}
	p := cp.(*clientpacket.SellResponse)
	vm := world.FindMobile(p.Vendor)
	if vm == nil || vm.Location.XYDistance(n.m.Location) > uo.MaxViewRange {
		return
	}
	// Remove items and calculate total
	total := 0
	for _, si := range p.SellItems {
		i := world.FindItem(si.Serial)
		if i == nil {
			continue
		}
		rp := game.Owner(i)
		if rp == nil || rp != n.m {
			continue
		}
		sa := si.Amount
		if sa > i.Amount {
			sa = i.Amount
		} else if sa < 1 {
			sa = 1
		}
		total += (i.Value / 2) * sa
		if sa == i.Amount {
			i.Remove()
		} else {
			i.Amount -= sa
			world.UpdateItem(i)
		}
	}
	// Payment
	gc := game.NewItem("GoldCoin")
	gc.Amount = total
	if !n.m.DropToBackpack(gc, false) {
		// Try a check instead
		gc.Remove()
		check := game.NewItem("Check")
		check.IArg = total
		if !n.m.DropToBackpack(check, false) {
			// Don't over-stuff the backpack, just let the check fall to their
			// feet.
			n.m.DropToFeet(check)
		}
	}
	n.Sound(0x02E6, n.m.Location)
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
	if m, ok := o.(*game.Mobile); ok {
		n.Send(&serverpacket.NameResponse{
			Serial: p.Serial,
			Name:   m.DisplayName(),
		})
	} else {
		n.Send(&serverpacket.NameResponse{
			Serial: p.Serial,
			Name:   o.(*game.Item).DisplayName(),
		})

	}
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
		var opl *serverpacket.OPLPacket
		switch t := o.(type) {
		case *game.Mobile:
			opl, _ = t.OPLPackets()
		case *game.Item:
			opl, _ = t.OPLPackets()
		}
		n.Send(opl)
	}
}

func handleRenameRequest(n *NetState, cp clientpacket.Packet) {
	if n == nil || n.m == nil {
		return
	}
	p := cp.(*clientpacket.RenameRequest)
	m := world.FindMobile(p.Serial)
	if m == nil || m.ControlMaster != n.m {
		return
	}
	m.Name = p.Name
	m.ArticleA = false
	m.ArticleAn = false
	game.World.UpdateMobile(m)
}

func handleMacroRequest(n *NetState, cp clientpacket.Packet) {
	if n == nil || n.m == nil {
		return
	}
	p := cp.(*clientpacket.MacroRequest)
	switch p.MacroType {
	case uo.MacroTypeOpenDoor:
		l := n.m.Location.Forward(n.m.Facing)
		doors := world.Map().ItemBaseQuery("BaseDoor", l, 0)
		if len(doors) > 0 {
			if !n.TakeAction() {
				n.Cliloc(nil, 500119) // You must wait to perform another action.
				return
			}
			game.DynamicDispatch("DoubleClick", doors[0], n.m, nil)
		}
	default:
		log.Printf("warning: unsupported macro type %d", p.MacroType)
	}
}

func handleTextGUMPReply(n *NetState, cp clientpacket.Packet) {
	if n == nil || n.m == nil {
		return
	}
	p := cp.(*clientpacket.TextGUMPReply)
	if p.Serial != uo.SerialTextGUMP {
		return
	}
	n.HandleGUMPTextReply(p.Text)
}
