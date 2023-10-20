package game

import (
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// NetState is the interface the server client's network state object must
// implement to be compatible with this library of game objects.
type NetState interface {
	ContainerObserver

	//
	// Life cycle management
	//

	// Disconnect disconnects the underlying network connection
	Disconnect()
	// Account returns the account attached to this NetState, which will never
	// be nil.
	Account() *Account
	// TakeAction returns true if an action is allowed at this time. Examples of
	// actions are double-clicking anything basically and moving and
	// equipping items. This method assumes that the action will be taken after
	// this call and sets internal states to limit action speed.
	TakeAction() bool
	// Mobile returns the mobile associated with this state if any.
	Mobile() Mobile

	//
	// Speech and messaging
	//

	// Speech sends a speech packet
	Speech(Object, string, ...interface{})
	// Cliloc sends a localized client message packet
	Cliloc(Object, uo.Cliloc, ...string)

	//
	// Effects and random stuff
	//

	// Animate
	Animate(Mobile, uo.AnimationType, uo.AnimationAction)
	// Send sends a custom packet to the client
	Send(serverpacket.Packet) bool
	// Sound sends a sound to the client from the specified location
	Sound(uo.Sound, uo.Location)
	// Music sends a song to the client
	Music(uo.Song)
	// TargetSendCursor sends a targeting request to the client
	TargetSendCursor(uo.TargetType, func(*clientpacket.TargetResponse))

	//
	// Item management and updates
	//

	// SendItem sends an object information packet to the client
	SendObject(Object)
	// RemoveObject sends a delete object packet to the client
	RemoveObject(Object)
	// UpdateObject sends packets to update the stats of an object
	UpdateObject(Object)
	// UpdateMobile sends a
	// WordItem sends the WornItem packet to the given mobile
	WornItem(Wearable, Mobile)
	// DragItem sends a DragItem packet
	DragItem(Item, Mobile, uo.Location, Mobile, uo.Location)
	// DropReject sends the MoveItemReject packet with the given reason code
	DropReject(uo.MoveItemRejectReason)

	//
	// Mobile updates
	//

	// DrawPlayer sends a DrawPlayer packet for the attached mobile if any
	DrawPlayer()
	// MoveMobile sends an MoveMobile packet for the given mobile
	MoveMobile(Mobile)
	// UpdateSkill updates a single skill on the client side
	UpdateSkill(uo.Skill, uo.SkillLock, int)

	//
	// Gumps
	//

	// GUMP sends a generic GUMP to the client.
	GUMP(g interface{}, target, param Object)
	// CloseGump closes the named gump on the client
	CloseGump(gump uo.Serial)
	// OpenPaperDoll opens the paper doll of the given mobile
	OpenPaperDoll(m Mobile)
}
