package gumps

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("spawn", 0, func() GUMP {
		return &spawn{}
	})
}

// spawn implements a menu to manage region spawn entries
type spawn struct {
	StandardGUMP
	Region *game.Region // Region we are managing the spawn entries for
}

// Layout implements the game.GUMP interface.
func (g *spawn) Layout(target, param game.Object) {
	pages := len(g.Region.Entries) / 8
	if len(g.Region.Entries)%8 != 0 {
		pages++
	}
	if pages < 1 {
		pages = 1
	}
	g.Window(16, 10, "Spawn Entries: "+g.Region.Name, 0, uint32(pages))
	g.Text(0, 0, 2, uo.HueDefault, "MinZ")
	g.TextEntry(2, 0, 2, uo.HueDefault, strconv.FormatInt(int64(g.Region.SpawnMinZ), 10), 4, 1)
	g.Text(4, 0, 2, uo.HueDefault, "MaxZ")
	g.TextEntry(6, 0, 2, uo.HueDefault, strconv.FormatInt(int64(g.Region.SpawnMaxZ), 10), 4, 2)
	g.ReplyButton(8, 0, 2, 1, uo.HueDefault, "Auto", 6)
	g.ReplyButton(10, 0, 2, 1, uo.HueDefault, "Test", 3)
	g.ReplyButton(12, 0, 2, 1, uo.HueDefault, "Apply", 4)
	g.HorizontalBar(0, 1, 16)
	var i int
	for i = int(g.currentPage-1) * 8; i < len(g.Region.Entries) && i < int(g.currentPage)*8; i++ {
		ty := i%8 + 2
		e := g.Region.Entries[i]
		g.TextEntry(0, ty, 2, uo.HueDefault, strconv.FormatInt(int64(e.Amount), 10), 3, uint32(1001+i))
		g.TextEntry(2, ty, 8, uo.HueDefault, e.Template, 256, uint32(2001+i))
		g.Text(10, ty, 1, uo.HueDefault, "Every")
		g.TextEntry(11, ty, 2, uo.HueDefault, strconv.FormatInt(int64(e.Delay/uo.DurationMinute), 10), 5, uint32(3001+i))
		g.Text(13, ty, 1, uo.HueDefault, "Mins")
		g.GemButton(14, ty, SGGemButtonDelete, uint32(4001+i))
	}
	if i == len(g.Region.Entries) && g.currentPage == uint32(pages) {
		ty := i%8 + 2
		g.GemButton(0, ty, SGGemButtonAdd, 5)
	}
}

// HandleReply implements the GUMP interface.
func (g *spawn) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	fn := func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			return 0
		}
		return int(v)
	}
	// Data
	g.Region.SpawnMinZ = int8(fn(p.Text(1)))
	g.Region.SpawnMaxZ = int8(fn(p.Text(2)))
	for i, e := range g.Region.Entries {
		s := p.Text(uint16(1001 + i))
		if len(s) == 0 {
			continue
		}
		e.Amount = fn(s)
		e.Template = p.Text(uint16(2001 + i))
		e.Delay = uo.Time(fn(p.Text(uint16(3001+i)))) * uo.DurationMinute
	}
	// Standard reply
	if g.StandardReplyHandler(p) {
		return
	}
	// Handle tool buttons
	switch p.Button {
	case 3: // Test
		g.Region.FullRespawn()
		return
	case 4: // Apply
		// Do nothing and let the GUMP refresh
		return
	case 5: // Add entry
		g.Region.Entries = append(g.Region.Entries, &game.SpawnerEntry{
			Delay:  uo.DurationMinute * 5,
			Amount: 1,
		})
		return
	case 6: // Auto adjust Z limits
		if n.Mobile() == nil {
			return
		}
		g.Region.SpawnMinZ = n.Mobile().Location().Z - uo.PlayerHeight
		g.Region.SpawnMaxZ = g.Region.SpawnMinZ + uo.PlayerHeight*2
		return
	}
	// Delete buttons
	if p.Button >= 4001 {
		i := int(p.Button - 4001)
		if i >= len(g.Region.Entries) {
			return
		}
		e := g.Region.Entries[i]
		for _, o := range e.Objects {
			game.Remove(o.Object)
		}
		g.Region.Entries = append(g.Region.Entries[:i], g.Region.Entries[i+1:]...)
	}
}
