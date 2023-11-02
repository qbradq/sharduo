package gumps

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("spawner", func() GUMP {
		return &spawner{}
	})
}

// spawner edits a Spawner object
type spawner struct {
	StandardGUMP
	Spawner *game.Spawner // The spawner we are editing
}

// Layout implements the GUMP interface.
func (g *spawner) Layout(target, param game.Object) {
	var ok bool
	g.Window(23, 12, "Spawner Editor", SGFlagNoPageButtons)
	g.Page(1)
	g.Spawner, ok = param.(*game.Spawner)
	if !ok {
		return
	}
	g.Text(0, 0, 3, uo.HueDefault, "Radius")
	// Radius field is 9999
	g.TextEntry(3, 0, 3, uo.HueDefault, strconv.Itoa(g.Spawner.Radius), 5, 9999)
	for i, e := range g.Spawner.Entries {
		// Delete buttons start at 2000
		g.GemButton(0, 2+i, SGGemButtonDelete, uint32(2000+i))
		g.Image(2, 2+i, 6, 6, uo.HueDefault, uo.GUMP(2225+i))
		// Template names start at 0
		g.TextEntry(3, 2+i, 8, uo.HueDefault, e.Template, 64, uint32(i))
		g.Text(11, 2+i, 2, uo.HueDefault, "Amount")
		// Amount starts at 4000
		g.TextEntry(13, 2+i, 3, uo.HueDefault, strconv.Itoa(e.Amount), 3, uint32(4000+i))
		g.Text(16, 2+i, 1, uo.HueDefault, "Delay")
		// Delay starts at 5000
		g.TextEntry(17, 2+i, 3, uo.HueDefault, strconv.Itoa(int(e.Delay/uo.DurationMinute)), 6, uint32(5000+i))
		g.Text(20, 2+i, 3, uo.HueDefault, "Minutes")
	}
	if len(g.Spawner.Entries) < 8 {
		i := len(g.Spawner.Entries)
		// The add button is 1000
		g.GemButton(0, 2+i, SGGemButtonAdd, 1000)
	}
	// Apply button is 1001
	g.GemButton(20, 11, SGGemButtonApply, 1001)
	// Total respawn button is 1002
	g.ReplyButton(16, 11, 4, 1, uo.HueDefault, "Full Respawn", 1002)
}

// HandleReply implements the GUMP interface.
func (g *spawner) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	// Update all data
	g.Spawner.Radius, _ = strconv.Atoi(p.Text(9999))
	for i := 0; i < 8; i++ {
		g.updateRowData(i, p)
	}
	game.GetWorld().Update(g.Spawner)
	// Have to do a full respawn to sync spawner with changes
	g.Spawner.FullRespawn()
	// Standard behavior handling
	if g.StandardReplyHandler(p) {
		return
	}
	// Add button
	if p.Button == 1000 {
		if len(g.Spawner.Entries) >= 8 {
			// Sanity check
			return
		}
		g.Spawner.Entries = append(g.Spawner.Entries, &game.SpawnerEntry{
			Delay:  uo.DurationMinute * 5,
			Amount: 1,
		})
		return
	}
	// Apply button, does nothing
	if p.Button == 1001 {
		return
	}
	// Full test button
	if p.Button == 1002 {
		g.Spawner.FullRespawn()
		return
	}
	// Delete button
	if p.Button >= 2000 && p.Button < 3000 {
		i := p.Button - 2000
		g.Spawner.Entries[i] = nil
		g.Spawner.Entries = append(g.Spawner.Entries[:i], g.Spawner.Entries[i+1:]...)
		return
	}
}

func (g *spawner) updateRowData(i int, p *clientpacket.GUMPReply) {
	if i < 0 || i >= len(g.Spawner.Entries) {
		return
	}
	e := g.Spawner.Entries[i]
	e.Template = p.Text(uint16(i))
	e.Amount, _ = strconv.Atoi(p.Text(uint16(4000 + i)))
	mins, _ := strconv.Atoi(p.Text(uint16(5000 + i)))
	if mins < 1 {
		mins = 1
	}
	e.Delay = uo.DurationMinute * uo.Time(mins)
}
