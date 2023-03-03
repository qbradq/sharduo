package gumps

import (
	"fmt"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// GUMP is the interface all GUMP objects implement.
type GUMP interface {
	// Layout executes all of the layout functions that comprise this GUMP.
	// Must be called before Packet().
	Layout(target, param game.Object)
	// Packet returns a newly created serverpacket.Packet for this GUMP.
	Packet(x, y int, id, serial uo.Serial) serverpacket.Packet
}

// BaseGUMP represents a generic GUMP and is the basis for all other GUMPs.
type BaseGUMP struct {
	l     strings.Builder // Layout string
	lines []string        // List of all text lines used
}

// BaseGUMP does not implement the Layout() function. This forces includers to
// define their own to satisfy the GUMP interface.

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
func (g *BaseGUMP) Image(x, y int, gump uo.GUMP, hue uo.Hue) {
	if hue == uo.HueDefault {
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

// TiledImage implements the GUMP interface.
func (g *BaseGUMP) TiledImage(x, y, w, h int, gump uo.GUMP) {
	g.l.WriteString(fmt.Sprintf("{ gumppictiled %d %d %d %d %d }", x, y, w, h, gump))
}

// Item implements the GUMP interface.
func (g *BaseGUMP) Item(x, y int, item uo.Graphic, hue uo.Hue) {
	if hue == uo.HueDefault {
		g.l.WriteString(fmt.Sprintf("{ tilepic %d %d %d }", x, y, item))
	} else {
		g.l.WriteString(fmt.Sprintf("{ tilepichue %d %d %d %d }", x, y, item, hue))
	}
}

// Label implements the GUMP interface.
func (g *BaseGUMP) Label(x, y int, hue uo.Hue, text string) {
	g.l.WriteString(fmt.Sprintf("{ text %d %d %d %d }", x, y, hue, g.InsertLine(text)))
}

// CroppedLabel implements the GUMP interface.
func (g *BaseGUMP) CroppedLabel(x, y, w, h int, hue uo.Hue, text string) {
	g.l.WriteString(fmt.Sprintf("{ croppedtext %d %d %d %d %d %d }", x, y, w, h, hue, g.InsertLine(text)))
}

// Page implements the GUMP interface.
func (g *BaseGUMP) Page(page int) {
	g.l.WriteString(fmt.Sprintf("{ page %d }", page))
}

// RadioButton implements the GUMP interface.
func (g *BaseGUMP) RadioButton(x, y int, normal, pressed uo.GUMP, id uint32, on bool) {
	v := 0
	if on {
		v = 1
	}
	g.l.WriteString(fmt.Sprintf("{ radio %d %d %d %d %d %d }", x, y, normal, pressed, v, id))
}

// Sprite implements the GUMP interface.
func (g *BaseGUMP) Sprite(x, y int, gump uo.GUMP, sx, sy, w, h int) {
	g.l.WriteString(fmt.Sprintf("{ picinpic %d %d %d %d %d %d %d }", x, y, gump, w, h, sx, sy))
}

// TextEntry implements the GUMP interface.
func (g *BaseGUMP) TextEntry(x, y, w, h int, hue uo.Hue, id uint32, text string, limit int) {
	if limit < 1 {
		g.l.WriteString(fmt.Sprintf("{ textentry %d %d %d %d %d %d %d }", x, y, w, h, hue, id, g.InsertLine(text)))
	} else {
		g.l.WriteString(fmt.Sprintf("{ textentrylimited %d %d %d %d %d %d %d %d }", x, y, w, h, hue, id, g.InsertLine(text), limit))
	}
}
