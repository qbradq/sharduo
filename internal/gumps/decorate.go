package gumps

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("decorate", GUMPIDDecorate, func() GUMP { return &decorate{} })
}

type zMode uint8

const (
	zModeOnTop   zMode = 0
	zModeOnLevel zMode = 1
	zModeFixed   zMode = 2
)

// decorate implements a menu that allows the user to open the various other
// decoration menus.
type decorate struct {
	StandardGUMP
	fixedZ int8   // The absolute Z value to use if zMode == zModeFixed
	zMode  zMode  // The Z fixing mode
	hue    uo.Hue // The selected hue
}

// Layout implements the game.GUMP interface.
func (g *decorate) Layout(target, param game.Object) {
	g.Window(5, 14, "Decoration Tools Menu", 0, 1)
	g.ReplyButton(0, 0, 5, 1, uo.HueDefault, "Single-Tile Statics", 1)
	g.ReplyButton(0, 1, 5, 1, uo.HueDefault, "Multi-Tile Statics", 2)
	g.ReplyButton(0, 2, 5, 1, uo.HueDefault, "Floors", 3)
	g.ReplyButton(0, 3, 5, 1, uo.HueDefault, "Doors", 4)
	g.ReplyButton(0, 4, 5, 1, uo.HueDefault, "Signs", 5)
	g.ReplyButton(0, 5, 5, 1, uo.HueDefault, "Walls", 6)
	g.HorizontalBar(0, 6, 5)
	if g.zMode == zModeFixed {
		g.CheckedReplyButton(0, 7, 3, 1, uo.HueDefault, "Fixed", 8)
	} else {
		g.ReplyButton(0, 7, 3, 1, uo.HueDefault, "Fixed", 8)
	}
	g.TextEntry(3, 7, 2, uo.HueDefault, strconv.FormatInt(int64(g.fixedZ), 10), 4, 9)
	if g.zMode == zModeOnLevel {
		g.CheckedReplyButton(0, 8, 5, 1, uo.HueDefault, "Same Z", 10)
	} else {
		g.ReplyButton(0, 8, 5, 1, uo.HueDefault, "Same Z", 10)
	}
	if g.zMode == zModeOnTop {
		g.CheckedReplyButton(0, 9, 5, 1, uo.HueDefault, "On Top", 11)
	} else {
		g.ReplyButton(0, 9, 5, 1, uo.HueDefault, "On Top", 11)
	}
	g.HorizontalBar(0, 10, 5)
	g.ReplyButton(0, 11, 3, 1, g.hue, "Apply Hue", 12)
	g.TextEntry(3, 11, 2, uo.HueDefault, strconv.FormatInt(int64(g.hue), 10), 4, 13)
	g.HorizontalBar(0, 12, 5)
	g.ReplyButton(0, 13, 5, 1, uo.HueDefault, "Save Everything", 7)
}

// HandleReply implements the GUMP interface.
func (g *decorate) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	// Data
	v, err := strconv.ParseInt(p.Text(9), 0, 32)
	if err == nil {
		g.fixedZ = int8(v)
	}
	v, err = strconv.ParseInt(p.Text(13), 0, 32)
	if err == nil {
		g.hue = uo.Hue(v)
	}
	// Standard reply
	if g.StandardReplyHandler(p) {
		return
	}
	// Tool buttons
	switch p.Button {
	case 1:
		n.GUMP(New("statics"), nil, nil)
	case 2:
		n.GUMP(New("objects"), nil, nil)
	case 3:
		n.GUMP(New("floors"), nil, nil)
	case 4:
		n.GUMP(New("doors"), nil, nil)
	case 5:
		n.GUMP(New("signs"), nil, nil)
	case 7:
		executeCommand(n, "savestatics")
		executeCommand(n, "savedoors")
		executeCommand(n, "savesigns")
	case 8:
		g.zMode = zModeFixed
	case 10:
		g.zMode = zModeOnLevel
	case 11:
		g.zMode = zModeOnTop
	}
}

// targetVolume executes a function with a bounding rect selected by the client.
func (g *decorate) targetVolume(n game.NetState, fn func(uo.Bounds)) {
	n.Speech(n.Mobile(), "Starting Point")
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		start := tr.Location
		so := game.Find[game.Item](tr.TargetObject)
		n.Speech(n.Mobile(), "Ending Point")
		n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
			end := tr.Location
			eo := game.Find[game.Item](tr.TargetObject)
			lowest := start.Z
			if end.Z < start.Z {
				lowest = end.Z
			}
			highest := start.Z
			if so != nil && so.Highest() > highest {
				highest = so.Highest()
			}
			if end.Z > highest {
				highest = end.Z
			}
			if eo != nil && eo.Highest() > highest {
				highest = eo.Highest()
			}
			b := uo.BoundsOf(start, end)
			switch g.zMode {
			case zModeFixed:
				b.Z = g.fixedZ
				b.D = 1
			case zModeOnLevel:
				b.Z = lowest
				b.D = int16(highest) - int16(lowest)
			case zModeOnTop:
				b.Z = highest
				b.D = 1
			}
			fn(b)
		})
	})
}

// place places a single static with regard to a reference item, if any.
func (g *decorate) place(l uo.Location, exp string, ref game.Item) bool {
	item := template.Create[*game.StaticItem]("StaticItem")
	if item == nil {
		// Something very wrong
		return false
	}
	item.SetBaseGraphic(uo.Graphic(util.RangeExpression(exp, game.GetWorld().Random())))
	if item.BaseGraphic() == uo.GraphicNone {
		// Refuse to place bad objects
		return false
	}
	item.SetHue(g.hue)
	switch g.zMode {
	case zModeFixed:
		l.Z = g.fixedZ
	case zModeOnLevel:
		if ref != nil {
			l.Z = ref.Z()
		}
	case zModeOnTop:
		if ref != nil {
			l.Z = ref.Highest()
		}
	}
	item.SetLocation(l)
	game.GetWorld().Map().ForceAddObject(item)
	return true
}
