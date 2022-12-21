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
	if world.Map().MoveMobile(n.m, p.Direction) {
		n.Send(&serverpacket.MoveAcknowledge{
			Sequence:  p.Sequence,
			Notoriety: uo.NotorietyInnocent,
		})
	} else {
		// TODO reject movement packet
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
		// TODO Handle double-click on other objects
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
	// TODO Update visible objects for the client
}

func handleLiftRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil || n.m.IsItemOnCursor() {
		return
	}
	p := cp.(*clientpacket.LiftRequest)
	o := world.Find(p.Item)
	if o == nil {
		return
	}
	item, ok := o.(game.Item)
	if !ok {
		return
	}
	// TODO Range check
	n.m.SetItemInCursor(item)
}

func handleDropRequest(n *NetState, cp clientpacket.Packet) {
	if n.m == nil || !n.m.IsItemOnCursor() {
		log.Println("drop request with no item on cursor")
		return
	}
	p := cp.(*clientpacket.DropRequest)
	if p.Item != n.m.ItemInCursor().Serial() {
		log.Println("drop request for an item that was not on the player's cursor")
		return
	}
	// TODO Range check
	item := world.Find(p.Item)
	item.SetLocation(uo.Location{X: p.X, Y: p.Y, Z: p.Z})
	if !world.Map().SetNewParent(item, nil) {
		// TODO Drop reject
	}
}
