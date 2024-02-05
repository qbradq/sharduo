package gumps

import (
	"fmt"
	"path"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

type floorPatch struct {
	Name, C, F, N, E, S, W, NE, SE, SW, NW string
}

var floorPatches []*floorPatch

func init() {
	reg("floors", 0, func() GUMP {
		return &floors{}
	})
	var lfr util.ListFileReader
	r, err := data.FS.Open(path.Join("misc", "floors.ini"))
	if err != nil {
		panic("error: reading file misc/floors.ini")
	}
	defer r.Close()
	for _, seg := range lfr.ReadSegments(r) {
		p := &floorPatch{
			Name: seg.Name,
		}
		for _, line := range seg.Contents {
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				panic("error: malformed misc/floors.ini")
			}
			switch parts[0] {
			case "C":
				p.C = parts[1]
			case "F":
				p.F = parts[1]
			case "N":
				p.N = parts[1]
			case "E":
				p.E = parts[1]
			case "S":
				p.S = parts[1]
			case "W":
				p.W = parts[1]
			case "NE":
				p.NE = parts[1]
			case "SE":
				p.SE = parts[1]
			case "SW":
				p.SW = parts[1]
			case "NW":
				p.NW = parts[1]
			default:
				panic(fmt.Sprintf("error: unknown key \"%s\" in misc/floors.ini", parts[0]))
			}
		}
		floorPatches = append(floorPatches, p)
	}
}

// floors implements a menu to generate rectangular 9-ways like flooring,
// carpeting and large tables.
type floors struct {
	StandardGUMP
	f int // Selected floor index
}

// Layout implements the game.GUMP interface.
func (g *floors) Layout(target, param any) {
	pages := len(floorPatches) / 12
	if len(floorPatches)%12 != 0 {
		pages++
	}
	g.Window(18, 24, "9-way Floor Generator", 0, uint32(pages))
	for i := int(g.currentPage-1) * 12; i < len(floorPatches) && i < int(g.currentPage)*12; i++ {
		p := floorPatches[i]
		tx := i % 3
		ty := (i / 3) % 4
		g.ReplyButton(tx*6+0, ty*6+0, 6, 1, uo.HueDefault, p.Name, uint32(1001+i))
		g.Item(tx*6+0, ty*6+1, 0, 44, uo.HueDefault, uo.Graphic(util.RangeExpression(p.NW)))
		g.Item(tx*6+0, ty*6+1, 22, 66, uo.HueDefault, uo.Graphic(util.RangeExpression(p.N)))
		g.Item(tx*6+0, ty*6+1, 44, 88, uo.HueDefault, uo.Graphic(util.RangeExpression(p.NE)))
		g.Item(tx*6+0, ty*6+1, 22, 22, uo.HueDefault, uo.Graphic(util.RangeExpression(p.W)))
		g.Item(tx*6+0, ty*6+1, 44, 44, uo.HueDefault, uo.Graphic(util.RangeExpression(p.C)))
		g.Item(tx*6+0, ty*6+1, 66, 66, uo.HueDefault, uo.Graphic(util.RangeExpression(p.E)))
		g.Item(tx*6+0, ty*6+1, 44, 0, uo.HueDefault, uo.Graphic(util.RangeExpression(p.SW)))
		g.Item(tx*6+0, ty*6+1, 66, 22, uo.HueDefault, uo.Graphic(util.RangeExpression(p.S)))
		g.Item(tx*6+0, ty*6+1, 88, 44, uo.HueDefault, uo.Graphic(util.RangeExpression(p.SE)))

	}
}

// HandleReply implements the GUMP interface.
func (g *floors) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	if p.Button > 1001 {
		f := int(p.Button - 1001)
		if f > len(floorPatches) {
			return
		}
		g.f = f
		g.placeFloor(n)
	}
}

func (g *floors) placeFloor(n game.NetState) {
	a := n.GetGUMPByID(GUMPIDDecorate)
	if a == nil {
		return
	}
	d, ok := a.(*decorate)
	if !ok {
		return
	}
	d.targetVolume(n, func(b uo.Bounds) {
		p := floorPatches[g.f]
		l := uo.Point{
			Z: b.Top(),
		}
		for l.Y = b.Y; l.Y <= b.South(); l.Y++ {
			for l.X = b.X; l.X <= b.East(); l.X++ {
				// Tile selection
				var exp string
				if l.Y == b.Y {
					if l.X == b.X {
						exp = p.NW
					} else if l.X == b.East() {
						exp = p.NE
					} else {
						exp = p.N
					}
				} else if l.Y == b.South() {
					if l.X == b.X {
						exp = p.SW
					} else if l.X == b.East() {
						exp = p.SE
					} else {
						exp = p.S
					}
				} else {
					if l.X == b.X {
						exp = p.W
					} else if l.X == b.East() {
						exp = p.E
					} else {
						if p.F != "" {
							if l.X == b.X+1 || l.X == b.East()-1 ||
								l.Y == b.Y+1 || l.Y == b.South()-1 {
								exp = p.C
							} else {
								exp = p.F
							}
						} else {
							exp = p.C
						}
					}
				}
				d.place(l, exp, nil)
			}
		}
	})
}
