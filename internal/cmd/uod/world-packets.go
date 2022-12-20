package uod

import (
	"log"
	"time"

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
	worldHandlers.Add(0x6C, handleTargetResponse)
	worldHandlers.Add(0x34, handleStatusRequest)
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
		// TODO Status update response
	}
}

func handleWalkRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.WalkRequest)
	if time.Now().UnixMilli()-n.lastWalkRequestTime < uo.FastWalkDelayMS {
		log.Printf("fast walk prevention triggered for account %s\n", n.id)
		return
	}
	if n.m == nil {
		return
	}
	if world.Map().MoveObject(n.m, p.Direction.Bound()) {
		n.Send(&serverpacket.MoveAcknowledge{
			Sequence:  p.Sequence,
			Notoriety: uo.NotorietyInnocent,
		})
	} else {
		// TODO reject movement packet
	}
}

func handleDoubleClickRequest(n *NetState, cp clientpacket.Packet) {
	p := cp.(*clientpacket.DoubleClick)
	if p.IsSelf {
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
		// TODO Movement reject
		log.Println(p)
	}
}
