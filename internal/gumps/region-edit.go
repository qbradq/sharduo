package gumps

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("region-edit", 0, func() GUMP {
		return &regionEdit{}
	})
}

// region-edit implements a menu to edit a single region in detail.
type regionEdit struct {
	StandardGUMP
	Region *game.Region // The region being edited
}

// Layout implements the game.GUMP interface.
func (g *regionEdit) Layout(target, param game.Object) {
	fn := func(x, y int, flag game.RegionFeature, name string, id uint32) {
		if g.Region.Features&flag != 0 {
			g.CheckedReplyButton(x*4, y+7, 4, 1, uo.HueDefault, name, id)
		} else {
			g.ReplyButton(x*4, y+7, 4, 1, uo.HueDefault, name, id)
		}
	}
	fn2 := func(i int, b uo.Bounds) {
		ty := (i % 4) + 2
		g.ReplyButton(0, ty, 2, 1, uo.HueDefault, "TL", uint32(1001+i))
		g.TextEntry(2, ty, 2, uo.HueDefault, strconv.FormatInt(int64(b.X), 10), 4, uint32(2001+i))
		g.TextEntry(4, ty, 2, uo.HueDefault, strconv.FormatInt(int64(b.Y), 10), 4, uint32(3001+i))
		g.ReplyButton(6, ty, 2, 1, uo.HueDefault, "GO", uint32(4001+i))
		g.ReplyButton(8, ty, 2, 1, uo.HueDefault, "BR", uint32(5001+i))
		g.TextEntry(10, ty, 2, uo.HueDefault, strconv.FormatInt(int64(b.East()), 10), 4, uint32(6001+i))
		g.TextEntry(12, ty, 2, uo.HueDefault, strconv.FormatInt(int64(b.South()), 10), 4, uint32(7001+i))
		g.ReplyButton(14, ty, 2, 1, uo.HueDefault, "GO", uint32(8001+i))
		g.GemButton(16, ty, SGGemButtonDelete, uint32(10001+i))
	}
	pages := (len(g.Region.Rects) + 1) / 4
	if (len(g.Region.Rects)+1)%4 != 0 {
		pages++
	}
	g.Window(18, 8, "Region Editor", 0, uint32(pages))
	g.Text(0, 0, 1, uo.HueDefault, "Name")
	g.TextEntry(1, 0, 5, uo.HueDefault, g.Region.Name, 64, 1)
	g.Text(6, 0, 1, uo.HueDefault, "Song")
	g.TextEntry(7, 0, 4, uo.HueDefault, g.Region.Music, 64, 2)
	g.ReplyButton(11, 0, 2, 1, uo.HueDefault, "Test", 3)
	g.ReplyButton(13, 0, 2, 1, uo.HueDefault, "Show", 4)
	g.ReplyButton(15, 0, 3, 1, uo.HueDefault, "Spawn", 6)
	g.HorizontalBar(0, 1, 18)
	var i int
	for i = int(g.currentPage-1) * 4; i < len(g.Region.Rects) && i < int(g.currentPage)*4; i++ {
		r := g.Region.Rects[i]
		fn2(i, r)
	}
	if i == len(g.Region.Rects) && g.currentPage == uint32(pages) {
		ty := (i % 4) + 2
		g.GemButton(0, ty, SGGemButtonAdd, 5)
	}
	g.HorizontalBar(0, 6, 18)
	fn(0, 0, game.RegionFeatureSafeLogout, "Safe Logout", 9001+0)
	fn(1, 0, game.RegionFeatureGuarded, "Guard Zone", 9001+1)
	fn(2, 0, game.RegionFeatureNoTeleport, "No Recall", 9001+2)
	fn(3, 0, game.RegionFeatureSpawnOnGround, "Ground Spawn", 9001+3)
}

// HandleReply implements the GUMP interface.
func (g *regionEdit) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	fn := func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(v)
	}
	fn2 := func(x, y int16) {
		l := uo.Location{
			X: x,
			Y: y,
			Z: uo.MapMaxZ,
		}
		f, _ := game.GetWorld().Map().GetFloorAndCeiling(l, false, true)
		if f == nil {
			return
		}
		l.Z = f.StandingHeight()
		game.GetWorld().Map().TeleportMobile(n.Mobile(), l)
	}
	// Data
	g.Region.Name = p.Text(1)
	g.Region.Music = p.Text(2)
	game.GetWorld().Map().RemoveRegion(g.Region)
	for i := range g.Region.Rects {
		s := p.Text(uint16(2001 + i))
		if s == "" {
			continue
		}
		g.Region.Rects[i] = uo.BoundsOf(uo.Location{
			X: int16(fn(s)),
			Y: int16(fn(p.Text(uint16(3001 + i)))),
			Z: uo.MapMinZ,
		}, uo.Location{
			X: int16(fn(p.Text(uint16(6001 + i)))),
			Y: int16(fn(p.Text(uint16(7001 + i)))),
			Z: uo.MapMaxZ,
		})
	}
	game.GetWorld().Map().AddRegion(g.Region)
	defer n.RefreshGUMP(n.GetGUMPByID(GUMPIDRegions))
	// Standard reply
	if g.StandardReplyHandler(p) {
		return
	}
	// Handle replies
	switch p.Button {
	case 3: // Test music
		n.Music(uo.Music(util.RangeExpression(g.Region.Music, game.GetWorld().Random())))
		return
	case 4: // Show regions
		m := n.Mobile()
		if m == nil {
			break
		}
		b := m.Location().BoundsByRadius(int(m.ViewRange()))
		l := m.Location()
		for l.Y = b.Y; l.Y <= b.South(); l.Y++ {
			for l.X = b.X; l.X <= b.East(); l.X++ {
				if l.X == m.Location().X && l.Y == m.Location().Y {
					print("debug")
				}
				if !g.Region.Contains(l) {
					continue
				}
				l.Z = m.Location().Z
				f, _ := game.GetWorld().Map().GetFloorAndCeiling(l, false, false)
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
					Hue:            62,
					GFXBlendMode:   uo.GFXBlendModeNormal,
				})
			}
		}
		return
	case 5: // Add rect
		m := n.Mobile()
		if m == nil {
			break
		}
		game.GetWorld().Map().RemoveRegion(g.Region)
		g.Region.AddRect(m.Location().BoundsByRadius(0))
		game.GetWorld().Map().AddRegion(g.Region)
		return
	case 6: // Spawn button
		a := New("spawn")
		sg := a.(*spawn)
		sg.Region = g.Region
		n.GUMP(a, nil, nil)
		return
	}
	// Delete buttons
	if p.Button >= 10001 {
		i := int(p.Button - 10001)
		if i >= len(g.Region.Rects) {
			return
		}
		game.GetWorld().Map().RemoveRegion(g.Region)
		g.Region.Rects = append(g.Region.Rects[:i], g.Region.Rects[i+1:]...)
		g.Region.ForceRecalculateBounds()
		game.GetWorld().Map().AddRegion(g.Region)
		return
	}
	// Flag buttons
	if p.Button >= 9001 {
		if p.Button == 9001 {
			g.Region.Features ^= game.RegionFeatureSafeLogout
		}
		if p.Button == 9002 {
			g.Region.Features ^= game.RegionFeatureGuarded
		}
		if p.Button == 9003 {
			g.Region.Features ^= game.RegionFeatureNoTeleport
		}
		if p.Button == 9004 {
			g.Region.Features ^= game.RegionFeatureSpawnOnGround
		}
		return
	}
	// GO BR
	if p.Button >= 8001 {
		i := int(p.Button - 8001)
		if i >= len(g.Region.Rects) {
			return
		}
		b := g.Region.Rects[i]
		fn2(b.East(), b.South())
		return
	}
	// Set BR
	if p.Button >= 5001 {
		i := int(p.Button - 5001)
		if i >= len(g.Region.Rects) {
			return
		}
		r := g.Region.Rects[i]
		n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
			l := tr.Location
			l.Z = uo.MapMaxZ
			game.GetWorld().Map().RemoveRegion(g.Region)
			g.Region.Rects[i] = uo.BoundsOf(uo.Location{
				X: r.X,
				Y: r.Y,
				Z: uo.MapMinZ,
			}, l)
			g.Region.ForceRecalculateBounds()
			game.GetWorld().Map().AddRegion(g.Region)
			n.RefreshGUMP(g)
		})
		return
	}
	// GO TL
	if p.Button >= 4001 {
		i := int(p.Button - 4001)
		if i >= len(g.Region.Rects) {
			return
		}
		b := g.Region.Rects[i]
		fn2(b.X, b.Y)
		return
	}
	// Set TL
	if p.Button >= 1001 {
		i := int(p.Button - 1001)
		if i >= len(g.Region.Rects) {
			return
		}
		r := g.Region.Rects[i]
		n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
			l := tr.Location
			l.Z = uo.MapMinZ
			game.GetWorld().Map().RemoveRegion(g.Region)
			g.Region.Rects[i] = uo.BoundsOf(l, uo.Location{
				X: r.East(),
				Y: r.South(),
				Z: uo.MapMaxZ,
			})
			g.Region.ForceRecalculateBounds()
			game.GetWorld().Map().AddRegion(g.Region)
			n.RefreshGUMP(g)
		})
		return
	}
}
