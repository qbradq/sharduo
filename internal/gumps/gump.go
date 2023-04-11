package gumps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

var reWhitespace = regexp.MustCompile(`\s+`)

// MungHTMLForGUMP manipulates the input line-by-line in the following way:
// 1. All leading and trailing whitespace is removed
// 2. The trailing newline is replaced with a single space
// 3. Consecutive whitespace is collapsed into a single space
func MungHTMLForGUMP(in string) string {
	in = strings.TrimSpace(in) + " "
	return reWhitespace.ReplaceAllLiteralString(in, " ")
}

// GUMP is the interface all GUMP objects implement.
type GUMP interface {
	// Layout executes all of the layout functions that comprise this GUMP.
	// Must be called before Packet().
	Layout(target, param game.Object)
	// InvalidateLayout resets the internal state so that Layout() may be called
	// again.
	InvalidateLayout()
	// Packet returns a newly created serverpacket.Packet for this GUMP.
	Packet(x, y int, sender, serial uo.Serial) serverpacket.Packet
	// HandleReply is called to process all replies for this GUMP. HandleReply
	// is not expected to handle GUMP close requests. The server keeping track
	// of open GUMPs should do that. Additionally the server needs to call
	// Layout again and send the new GUMP packet back to the client.
	HandleReply(n game.NetState, p *clientpacket.GUMPReply)
	// TypeCode returns the GUMP's type code.
	TypeCode() uo.Serial
	// SetTypeCode sets the GUMP's type code.
	SetTypeCode(uo.Serial)
}

// BaseGUMP represents a generic GUMP and is the basis for all other GUMPs.
type BaseGUMP struct {
	l        strings.Builder // Layout string
	lines    []string        // List of all text lines used
	typeCode uo.Serial
}

// BaseGUMP does not implement the Layout() function. This forces includers to
// define their own.

// BaseGUMP does not implement the HandleReply() function. This forces includers
// to define their own.

// TypeCode implements the GUMP interface.
func (g *BaseGUMP) TypeCode() uo.Serial { return g.typeCode }

// SetTypeCode implements the GUMP interface.
func (g *BaseGUMP) SetTypeCode(c uo.Serial) { g.typeCode = c }

// InvalidateLayout implements the GUMP interface.
func (g *BaseGUMP) InvalidateLayout() {
	g.l.Reset()
	g.lines = nil
}

// Packet implements the GUMP interface.
func (g *BaseGUMP) Packet(x, y int, sender, typeCode uo.Serial) serverpacket.Packet {
	return &serverpacket.GUMP{
		Sender:   sender,
		TypeCode: typeCode,
		Layout:   g.l.String(),
		Location: uo.Location{
			X: int16(x),
			Y: int16(y),
		},
		Lines: g.lines,
	}
}

// InsertLine adds l to the GUMP's list of text lines and returns the reference
// number.
func (g *BaseGUMP) InsertLine(l string) int {
	g.lines = append(g.lines, l)
	return len(g.lines) - 1
}

// NoClose flags the GUMP as not closable with right-click.
func (g *BaseGUMP) NoClose() {
	g.l.WriteString("{ noclose }")
}

// NoDispose flags the GUMP as not closable with the escape key. This is the
// default for most GUMPs.
func (g *BaseGUMP) NoDispose() {
	g.l.WriteString("{ nodispose }")
}

// NoMove disables the ability to close the GUMP to be moved on screen.
func (g *BaseGUMP) NoMove() {
	g.l.WriteString("{ nomove }")
}

// NoResize disables the ability to resize certain GUMPs.
func (g *BaseGUMP) NoResize() {
	g.l.WriteString("{ noresize }")
}

// AlphaRegion is supposed to add a checkered transparent region to the GUMP.
// However ClassicUO - at least as of version 0.1.11.39 - adds this region but
// also makes the entire GUMP 50% alpha.
func (g *BaseGUMP) AlphaRegion(x, y, w, h int) {
	g.l.WriteString(fmt.Sprintf("{ checkertrans %d %d %d %d }", x, y, w, h))
}

// Background creates a 3x3 background tiled to the given dimensions.
func (g *BaseGUMP) Background(x, y, w, h int, bg uo.GUMP) {
	g.l.WriteString(fmt.Sprintf("{ resizepic %d %d %d %d %d }", x, y, bg, w, h))
}

// ReplyButton creates a button that will generate a reply packet.
func (g *BaseGUMP) ReplyButton(x, y int, normal, pressed uo.GUMP, id uint32) {
	g.l.WriteString(fmt.Sprintf("{ button %d %d %d %d 1 0 %d }", x, y, normal, pressed, id))
}

// PageButton creates a button that will change pages.
func (g *BaseGUMP) PageButton(x, y int, normal, pressed uo.GUMP, page uint32) {
	g.l.WriteString(fmt.Sprintf("{ button %d %d %d %d 0 %d 0 }", x, y, normal, pressed, page))
}

// Checkbox creates a checkbox button.
func (g *BaseGUMP) Checkbox(x, y int, normal, pressed uo.GUMP, id uint32, checked bool) {
	v := 0
	if checked {
		v = 1
	}
	g.l.WriteString(fmt.Sprintf("{ checkbox %d %d %d %d %d %d }", x, y, normal, pressed, v, id))
}

// Group starts a RadioButton group. RadioButton groups must be book-ended by
// the EndGroup function otherwise they will not work on pages after 1 according
// to the POL GUMP documentation. No idea if this is an issue in ClassicUO.
func (g *BaseGUMP) Group(n uint32) {
	g.l.WriteString(fmt.Sprintf("{ group %d }", n))
}

// EndGroup ends the current RadioButton group. RadioButton groups must be book-
// ended by the EndGroup function otherwise they will not work on pages after 1
// according to the POL GUMP documentation. No idea if this is an issue in
// ClassicUO.
func (g *BaseGUMP) EndGroup() {
	g.l.WriteString("{ endgroup }")
}

// HTML creates an HTML view.
//
// According to the POL GUMP documentation these are the HTML tags that are
// supported:
// <B></B>..................................... Bold
// <BIG></BIG>................................. Bigger font
// <SMALL></SMALL>............................. Smaller font
// <EM></EM>................................... Emphasis
// <I></I>..................................... Italicized
// <U></U>..................................... Underlined
// <H1></H1>................................... Heading 1 - largest
// <H2></H2>................................... Heading 2
// <H3></H3>................................... Heading 3
// <H4></H4>................................... Heading 4
// <H5></H5>................................... Heading 5
// <H6></H6>................................... Heading 6 - smallest
// <a href=""></a>............................. Hyperlink
// <div align="right"></DIV>................... Division on the right
// <div align="left"></DIV>.................... Division on the left
// <left></left>............................... Left-align text
// <P>......................................... Paragraph block
// <CENTER></CENTER>........................... Center-align text
// <BR></BR>................................... Line break
// <BASEFONT color=#ffffff size=1-7></BASEFONT> Set default font?
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

// Image places a GUMP image.
func (g *BaseGUMP) Image(x, y int, gump uo.GUMP, hue uo.Hue) {
	if hue == uo.HueDefault {
		g.l.WriteString(fmt.Sprintf("{ gumppic %d %d %d }", x, y, gump))
	} else {
		g.l.WriteString(fmt.Sprintf("{ gumppic %d %d %d hue=%d }", x, y, gump, hue))
	}
}

// TiledImage places a tiled GUMP image.
func (g *BaseGUMP) TiledImage(x, y, w, h int, gump uo.GUMP) {
	g.l.WriteString(fmt.Sprintf("{ gumppictiled %d %d %d %d %d }", x, y, w, h, gump))
}

// Item places an item graphic.
func (g *BaseGUMP) Item(x, y int, item uo.Graphic, hue uo.Hue) {
	if hue == uo.HueDefault {
		g.l.WriteString(fmt.Sprintf("{ tilepic %d %d %d }", x, y, item))
	} else {
		g.l.WriteString(fmt.Sprintf("{ tilepichue %d %d %d %d }", x, y, item, hue))
	}
}

// Label places text on the GUMP. NOTE: Newline characters are not supported.
func (g *BaseGUMP) Label(x, y int, hue uo.Hue, text string) {
	g.l.WriteString(fmt.Sprintf("{ text %d %d %d %d }", x, y, hue, g.InsertLine(text)))
}

// CroppedLabel like Label but within a cropped area.
func (g *BaseGUMP) CroppedLabel(x, y, w, h int, hue uo.Hue, text string) {
	g.l.WriteString(fmt.Sprintf("{ croppedtext %d %d %d %d %d %d }", x, y, w, h, hue, g.InsertLine(text)))
}

// Page begins the numbered page.
func (g *BaseGUMP) Page(page uint32) {
	g.l.WriteString(fmt.Sprintf("{ page %d }", page))
}

// RadioButton creates a radio button switch. See Group and EndGroup.
func (g *BaseGUMP) RadioButton(x, y int, normal, pressed uo.GUMP, id uint32, on bool) {
	v := 0
	if on {
		v = 1
	}
	g.l.WriteString(fmt.Sprintf("{ radio %d %d %d %d %d %d }", x, y, normal, pressed, v, id))
}

// TextEntry creates a text entry area. If limit is less than 1 no limit will be
// enforced.
func (g *BaseGUMP) TextEntry(x, y, w, h int, hue uo.Hue, text string, limit int, id uint32) {
	if limit < 1 {
		g.l.WriteString(fmt.Sprintf("{ textentry %d %d %d %d %d %d %d }", x, y, w, h, hue, id, g.InsertLine(text)))
	} else {
		g.l.WriteString(fmt.Sprintf("{ textentrylimited %d %d %d %d %d %d %d %d }", x, y, w, h, hue, id, g.InsertLine(text), limit))
	}
}
