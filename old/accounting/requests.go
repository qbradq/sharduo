package accounting

import (
	"net"

	"github.com/qbradq/sharduo/pkg/uo"

	"github.com/qbradq/sharduo/internal/common"
)

// LoginRequest account authentication request (login screen)
type LoginRequest struct {
	State    common.NetState
	Username string
	Password string
}

func doLogin(r *LoginRequest) {
	// Send fixed server list for now
	ip := net.ParseIP(common.Config.GetString("network.externalIP", "127.0.0.1"))
	addr := net.IPAddr{
		IP: ip,
	}
	p := uo.NewServerPacketServerList(make([]byte, 0, 1024))
	p.AddServer("ShardUO TC", 0, 0, addr)
	p.Finish()
	r.State.SendPacket(p)
}

// SelectServerRequest game server connection request
type SelectServerRequest struct {
	State common.NetState
	Slot  uint
}

func doSelectServer(r *SelectServerRequest) {
	// Send fixed connection information for now
	ip := net.ParseIP(common.Config.GetString("network.externalIP", "127.0.0.1"))
	addr := net.IPAddr{
		IP: ip,
	}
	p := uo.NewServerPacketConnectToServer(make([]byte, 0, 128), addr, 2593)
	r.State.SendPacket(p)
}

// GameServerLoginRequest game server login request (after shard selection)
type GameServerLoginRequest struct {
	State    common.NetState
	Username string
	Password string
}

func doGameServerLogin(r *GameServerLoginRequest) {
	p := uo.NewServerPacketCharacterList(make([]byte, 0, 2048))
	p.AddCharacter(r.Username)
	p.FinishCharacterList()
	p.AddStartingLocation("a", "a")
	p.Finish(uo.FeatureFlagSiege)
	r.State.SendPacket(p)
}
