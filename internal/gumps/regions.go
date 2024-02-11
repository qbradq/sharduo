package gumps

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("regions", GUMPIDRegions, func() GUMP {
		return &regions{}
	})
}

// regions implements a menu to allow region management from within the game.
type regions struct {
	StandardGUMP
	regions []*game.Region
	tr      *game.Mobile
}

// Layout implements the game.GUMP interface.
func (g *regions) Layout(target, param any) {
	g.tr = target.(*game.Mobile)
	g.regions = game.World.Map().RegionsAt(g.tr.Location)
	pages := len(g.regions) / 5
	if len(g.regions)%5 != 0 {
		pages++
	}
	if pages < 1 {
		pages = 1
	}
	g.Window(12, 7, "Regions at Location", 0, uint32(pages))
	g.ReplyButton(0, 0, 2, 1, uo.HueDefault, "New", 1)
	g.ReplyButton(2, 0, 3, 1, uo.HueDefault, "Refresh", 2)
	g.ReplyButton(5, 0, 2, 1, uo.HueDefault, "Show", 3)
	g.ReplyButton(7, 0, 3, 1, uo.HueDefault, "Save All", 4)
	g.ReplyButton(10, 0, 3, 1, uo.HueDefault, "Load All", 5)
	g.HorizontalBar(0, 1, 12)
	for i := int(g.currentPage-1) * 5; i < len(g.regions) && i < int(g.currentPage)*5; i++ {
		ty := i % 10
		r := g.regions[i]
		g.ReplyButton(0, ty+2, 7, 1, uo.HueDefault, r.Name, uint32(1001+i))
		g.ReplyButton(7, ty+2, 3, 1, uo.HueDefault, "Copy", uint32(3001+i))
		g.GemButton(10, ty+2, SGGemButtonDelete, uint32(2001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *regions) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	switch p.Button {
	case 1: // New region button
		a := New("region-edit")
		re := a.(*regionEdit)
		re.Region = &game.Region{
			Name:      "New Region",
			SpawnMinZ: uo.MapMinZ,
			SpawnMaxZ: uo.MapMaxZ,
		}
		re.Region.AddRect(n.Mobile().Location.BoundsByRadius(0))
		game.World.Map().AddRegion(re.Region)
		n.GUMP(a, 0, 0)
		n.RefreshGUMP(g)
		return
	case 2: // Refresh view button
		// Do nothing and let the GUMP refresh
		return
	case 3: // Show all regions button
		b := g.tr.Location.BoundsByRadius(g.tr.ViewRange)
		for _, r := range game.World.Map().RegionsWithin(b) {
			hue := uo.Hue(util.Random(0, 199)*5 + 1)
			l := g.tr.Location
			for l.Y = b.Y; l.Y <= b.South(); l.Y++ {
				for l.X = b.X; l.X <= b.East(); l.X++ {
					if !r.Contains(l) {
						continue
					}
					l.Z = g.tr.Location.Z
					f, _ := game.World.Map().GetFloorAndCeiling(l, false, false)
					if f == nil {
						continue
					}
					l.Z = f.StandingHeight()
					n.Send(&serverpacket.GraphicalEffect{
						GFXType:        uo.GFXTypeFixed,
						Graphic:        0x0495,
						SourceLocation: l,
						TargetLocation: l,
						Speed:          15,
						Duration:       75,
						Hue:            hue,
						GFXBlendMode:   uo.GFXBlendModeNormal,
					})
				}
			}
		}
	case 4:
		executeCommand(n, "save_regions")
	case 5:
		executeCommand(n, "load_regions")
	}
	// Region copy button
	if p.Button >= 3001 {
		i := int(p.Button - 3001)
		if i >= len(g.regions) {
			return
		}
		r := g.regions[i]
		nr := &game.Region{
			Name:     r.Name + " - Copy",
			Rects:    make([]uo.Bounds, len(r.Rects)),
			Features: r.Features,
			Music:    r.Music,
		}
		copy(nr.Rects, r.Rects)
		a := New("region-edit")
		re := a.(*regionEdit)
		re.Region = nr
		game.World.Map().AddRegion(nr)
		n.GUMP(a, 0, 0)
		n.RefreshGUMP(g)
	}
	// Region delete button
	if p.Button >= 2001 {
		i := int(p.Button - 2001)
		if i >= len(g.regions) {
			return
		}
		r := g.regions[i]
		game.World.Map().RemoveRegion(r)
		return
	}
	// Region button
	if p.Button >= 1001 {
		i := int(p.Button - 1001)
		if i >= len(g.regions) {
			return
		}
		a := New("region-edit")
		re := a.(*regionEdit)
		re.Region = g.regions[i]
		n.GUMP(a, 0, 0)
		return
	}
}
