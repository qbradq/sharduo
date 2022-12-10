package uod

import (
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// These functions handle client packets within the world process with direct
// access to the memory model.
func init() {
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

	}
}
