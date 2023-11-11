package commands

import (
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Player-facing commands live here

func init() {
	regcmd(&cmdesc{"chat", []string{"c", "global", "g"}, commandChat, game.RolePlayer, "chat", "Sends global chat speech"})
	regcmd(&cmdesc{"graphic", nil, commandGraphic, game.RolePlayer, "graphic", "Tells you the item graphic number of the object"})
	regcmd(&cmdesc{"hue", nil, commandHue, game.RolePlayer, "hue", "Tells you the hue number of the object"})
}

func commandGraphic(n game.NetState, args CommandArgs, cl string) {
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		var bg, ag uo.Graphic
		var speaker game.Object
		if tr.TargetObject != uo.SerialZero {
			i := game.Find[game.Item](tr.TargetObject)
			if i == nil {
				return
			}
			ag = i.Graphic()
			bg = i.BaseGraphic()
			speaker = i
		} else {
			bg = tr.Graphic
			ag = tr.Graphic
			speaker = n.Mobile()
		}
		if ag == bg {
			n.Speech(speaker, "0x%04X", bg)
		} else {
			n.Speech(speaker, "0x%04X (0x%04X)", bg, ag)
		}
	})
}

func commandChat(n game.NetState, args CommandArgs, cl string) {
	if n.Mobile() == nil {
		return
	}
	hue := uo.Hue(args.Int(1))
	line := ""
	if hue != uo.HueDefault {
		parts := strings.SplitN(cl, " ", 3)
		if len(parts) != 3 {
			return
		}
		line = parts[2]
	} else {
		parts := strings.SplitN(cl, " ", 2)
		if len(parts) != 2 {
			return
		}
		line = parts[1]
	}
	globalChat(hue, n.Mobile().DisplayName(), line)
}

func commandHue(n game.NetState, args CommandArgs, cl string) {
	if n == nil || n.Mobile() == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		o := game.GetWorld().Find(tr.TargetObject)
		if o == nil {
			return
		}
		n.Speech(n.Mobile(), "Hue %d", o.Hue())
	})
}
