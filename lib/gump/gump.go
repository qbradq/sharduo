package gump

import (
	"fmt"
	"strings"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// GUMP is the interface all GUMP objects implement.
type GUMP interface {
	// InsertLine inserts line into the GUMP's list of text lines and returns
	// the new line's index number.
	InsertLine(line string) int
	// AlphaRegion creates a checker-pattern alpha effect in the given area.
	AlphaRegion(x, y, w, h int)
	// Background creates a 9-way background.
	Background(x, y, w, h int, bg uo.GUMP)
	// ReplyButton creates a normal button that causes a reply and closes the
	// GUMP.
	ReplyButton(x, y int, normal, pressed uo.GUMP, id uint32)
	// PageButton creates a button that hides all pages except page 0, then
	// shows the given page number. This all happens client-side and does not
	// generate a response packet.
	PageButton(x, y int, normal, pressed uo.GUMP, page int)
	// Group defines a button group.
	Group(n int)
	// Checkbox creates an on-off checkbox button.
	Checkbox(x, y int, normal, pressed uo.GUMP, id uint32, checked bool)
	// HTML creates an HTML view area with literal text.
	HTML(x, y, w, h int, html string, background, scrollbar bool)
	// Image places a GUMP image.
	Image(x, y int, gump uo.Gump, hue uo.Hue)
	// ReplyImageTileButton creates a button element with an item graphic over
	// it that generates a reply and closes the GUMP.
	ReplyImageTileButton(x, y, w, h, normal, pressed uo.GUMP, item uo.Graphic, hue uo.Hue, id uint32)
	// PageImageTileButton creates a button element with an item graphic over
	// it that hides all pages except page 0, then shows the indicated page.
	PageImageTileButton(x, y, w, h, normal, pressed uo.GUMP, item uo.Graphic, hue uo.Hue, page int)
	// Packet returns a newly created serverpacket.Packet for this GUMP.
	Packet(x, y int, id, serial uo.Serial) serverpacket.Packet
}

// BaseGUMP represents a generic GUMP and is the basis for all other GUMPs.
type BaseGUMP struct {
	l     strings.Builder // Layout string
	lines []string        // List of all text lines used
}

// Packet implements the GUMP interface.
func (g *BaseGUMP) Packet(x, y int, id, serial uo.Serial) serverpacket.Packet {
	return &serverpacket.GUMP{
		ProcessID: id,
		GUMPID:    serial,
		Layout:    g.l.String(),
		Location: uo.Location{
			X: int16(x),
			Y: int16(y),
		},
		Lines: g.lines,
	}
}

// InsertLine implements the GUMP interface.
func (g *BaseGUMP) InsertLine(l string) int {
	g.lines = append(g.lines, l)
	return len(g.lines) - 1
}

// AlphaRegion implements the GUMP interface.
func (g *BaseGUMP) AlphaRegion(x, y, w, h int) {
	g.l.WriteString(fmt.Sprintf("{ checkertrans %d %d %d %d }", x, y, w, h))
}

// Background implements the GUMP interface.
func (g *BaseGUMP) Background(x, y, w, h int, bg uo.GUMP) {
	g.l.WriteString(fmt.Sprintf("{ resizepic %d %d %d %d %d }", x, y, bg, w, h))
}

// ReplyButton implements the GUMP interface.
func (g *BaseGUMP) ReplyButton(x, y int, normal, pressed uo.GUMP, id uint32) {
	g.l.WriteString(fmt.Sprintf("{ button %d %d %d %d 1 0 %d }", x, y, normal, pressed, id))
}

// PageButton implements the GUMP interface.
func (g *BaseGUMP) PageButton(x, y int, normal, pressed uo.GUMP, page int) {
	g.l.WriteString(fmt.Sprintf("{ button %d %d %d %d 0 %d 0 }", x, y, normal, pressed, page))
}

// Checkbox implements the GUMP interface.
func (g *BaseGUMP) Checkbox(x, y int, normal, pressed uo.GUMP, id uint32, checked bool) {
	v := 0
	if checked {
		v = 1
	}
	g.l.WriteString(fmt.Sprintf("{ checkbox %d %d %d %d %d %d }", x, y, normal, pressed, v, id))
}

// Group implements the GUMP interface.
func (g *BaseGUMP) Group(n int) {
	g.l.WriteString(fmt.Sprintf("{ group %d }", n))
}

// HTML implements the GUMP interface.
func (g *BaseGUMP) HTML(x, y, w, h int, html string, background, scrollbar bool) {
	bg := 0
	if background {
		bg = 1
	}
	sb := 0
	if scrollbar {
		sb = 1
	}
	g.l.WriteString(fmt.Sprintf("{ htmlgump %d %d %d %d %d %d %d }", x, y, w, h,
		g.InsertLine(html), bg, sb))
}

// Image implements the GUMP interface.
func (g *BaseGUMP) Image(x, y int, gump uo.Gump, hue uo.Hue) {
	if hue == 0 {
		g.l.WriteString(fmt.Sprintf("{ gumppic %d %d %d }", x, y, gump))
	} else {
		g.l.WriteString(fmt.Sprintf("{ gumppic %d %d %d hue=%d }", x, y, gump, hue))
	}
}

// ReplyImageTileButton implements the GUMP interface.
func (g *BaseGUMP) ReplyImageTileButton(x, y, w, h, normal, pressed uo.GUMP, item uo.Graphic, hue uo.Hue, id uint32) {
	g.l.WriteString(fmt.Sprintf("{ buttontileart %d %d %d %d 1 0 %d %d %d %d %d }",
		x, y, normal, pressed, id, item, hue, w, h))
}

// PageImageTileButton implements the GUMP interface.
func (g *BaseGUMP) PageImageTileButton(x, y, w, h, normal, pressed uo.GUMP, item uo.Graphic, hue uo.Hue, page int) {
	g.l.WriteString(fmt.Sprintf("{ buttontileart %d %d %d %d 0 %d %d %d %d %d 0 }",
		x, y, normal, pressed, page, item, hue, w, h))
}
