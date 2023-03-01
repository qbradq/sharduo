package gump

import (
	"fmt"
	"strings"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// GUMP represents a generic GUMP.
type GUMP struct {
	l strings.Builder // Layout string
}

// AlphaRegion creates a checker-pattern alpha effect in the given area
func (g *GUMP) AlphaRegion(x, y, w, h int) {
	g.l.WriteString(fmt.Sprintf("{ checkertrans %d %d %d %d }", x, y, w, h))
}

// Background creates a 9-way background
func (g *GUMP) Background(x, y, w, h int, background uo.GUMP) {
	g.l.WriteString(fmt.Sprintf("{ resizepic %d %d %d %d %d }", x, y, background, w, h))
}

// Packet returns a newly created serverpacket.Packet for this GUMP
func (g *GUMP) Packet(x, y int, id, serial uo.Serial) serverpacket.Packet {
	return &serverpacket.GUMP{
		ProcessID: id,
		GUMPID:    serial,
		Layout:    g.l.String(),
		Location: uo.Location{
			X: int16(x),
			Y: int16(y),
		},
	}
}
