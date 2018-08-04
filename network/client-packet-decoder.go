package network

import (
	"github.com/qbradq/sharduo/accounting"
)

type clientPacketDecoder func(*PacketBuffer)

func decodePacket80(buf *PacketBuffer) {
	buf.Seek(1)
	username := buf.GetASCII(30)
	password := buf.GetASCII(30)
	accounting.Login(username, password)
}
