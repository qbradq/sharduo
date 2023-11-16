package gumps

import (
	"path"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("signs", 0, func() GUMP {
		return &signs{}
	})
	var lfr util.ListFileReader
	r, err := data.FS.Open(path.Join("misc", "signs.ini"))
	if err != nil {
		panic("error: reading file misc/signs.ini")
	}
	defer r.Close()
	segs := lfr.ReadSegments(r)
	if len(segs) != 2 || segs[0].Name != "Signs" || segs[1].Name != "Signposts" {
		panic("error: malformed misc/signs.ini")
	}
	for _, s := range segs[0].Contents {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			panic("error: malformed misc/signs.ini")
		}
		v, err := strconv.ParseInt(parts[1], 0, 32)
		if err != nil {
			panic(err)
		}
		signGraphics[parts[0]] = uo.Graphic(v)
		signNames = append(signNames, parts[0])
	}
	for _, s := range segs[1].Contents {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		signpostGraphics = append(signpostGraphics, uo.Graphic(v))
	}
}

var signNames []string
var signGraphics = map[string]uo.Graphic{}
var signpostGraphics []uo.Graphic

// signs implements a menu that allows smart sign placement.
type signs struct {
	StandardGUMP
	g uo.Graphic // Currently selected graphic
}

// Layout implements the game.GUMP interface.
func (g *signs) Layout(target, param game.Object) {
	pages := len(signGraphics) / 20
	if len(signGraphics)%20 != 0 {
		pages++
	}
	g.Window(10, 20, "Sign Placement", 0, uint32(pages))
	for i := int(g.currentPage-1) * 20; i < len(signGraphics) && i < int(g.currentPage)*20; i++ {
		s := signNames[i]
		sg := signGraphics[s]
		ty := i % 20
		g.Item(0, ty, 0, 0, uo.HueDefault, sg)
		g.ReplyButton(1, ty, 9, 1, uo.HueDefault, s, uint32(1001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *signs) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	if p.Button >= 1001 {
		i := int(p.Button - 1001)
		if i >= len(signNames) {
			return
		}
		s := signNames[i]
		sg := signGraphics[s]
		g.g = sg
		g.placeSingle(n)
	}
}

func (g *signs) placeSingle(n game.NetState) {
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		a := n.GetGUMPByID(GUMPIDDecorate)
		if a == nil {
			return
		}
		d, ok := a.(*decorate)
		if !ok {
			return
		}
		l := tr.Location
		for _, s := range game.GetWorld().Map().StaticsAt(l) {
			for i, p := range signpostGraphics {
				if s.BaseGraphic() != p {
					continue
				}
				l.Z = s.Z()
				sg := g.g
				if i%2 != 0 {
					sg++
				}
				d.place(l, strconv.FormatInt(int64(sg), 10), nil)
				return
			}
		}
	})
}
