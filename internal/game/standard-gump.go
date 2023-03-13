package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// SGGemButton is an enumeration of the available gem buttons available to the
// GemButton() function.
type SGGemButton uint8

// All valid values for SGGemButton
const (
	SGGemButtonOptions   SGGemButton = 0
	SGGemButtonLogOut    SGGemButton = 1
	SGGemButtonJournal   SGGemButton = 2
	SGGemButtonSkills    SGGemButton = 3
	SGGemButtonChat      SGGemButton = 4
	SGGemButtonPeace     SGGemButton = 5
	SGGemButtonWar       SGGemButton = 6
	SGGemButtonStatus    SGGemButton = 7
	SGGemButtonCharacter SGGemButton = 8
	SGGemButtonHelp      SGGemButton = 9
	SGGemButtonAuto      SGGemButton = 10
	SGGemButtonManual    SGGemButton = 11
	SGGemButtonCancel    SGGemButton = 12
	SGGemButtonApply     SGGemButton = 13
	SGGemButtonDefault   SGGemButton = 14
	SGGemButtonOK        SGGemButton = 15
	SGGemButtonStrategy  SGGemButton = 16
	sgGemButtonCount     int         = 17
)

// sgGemDef defines the graphics and offsets for a gem button for the
// GemButton() function.
type sgGemDef struct {
	// Button normal
	n uo.GUMP
	// Button pressed
	p uo.GUMP
	// Button horizontal offset
	h int
	// Button vertical offset
	v int
}

// sgGemDefs defines all of the buttons for the GemButton() function.
var sgGemDefs [sgGemButtonCount]sgGemDef = [sgGemButtonCount]sgGemDef{
	{n: 2006, p: 2007, h: 0, v: 5},
	{n: 2009, p: 2010, h: 0, v: 5},
	{n: 2012, p: 2013, h: 0, v: 5},
	{n: 2015, p: 2016, h: 0, v: 5},
	{n: 2018, p: 2019, h: 0, v: 5},
	{n: 2021, p: 2022, h: 0, v: 5},
	{n: 2024, p: 2025, h: 0, v: 5},
	{n: 2027, p: 2028, h: 0, v: 5},
	{n: 2030, p: 2030, h: -1, v: 5},
	{n: 2031, p: 2032, h: 0, v: 5},
	{n: 2111, p: 2112, h: 4, v: 5},
	{n: 2114, p: 2115, h: 4, v: 5},
	{n: 2119, p: 2120, h: 4, v: 5},
	{n: 2122, p: 2123, h: 4, v: 5},
	{n: 2125, p: 2126, h: 4, v: 5},
	{n: 2128, p: 2129, h: 4, v: 5},
	{n: 2131, p: 2131, h: -1, v: 5},
}

// SGFlag describes the flags that a StandardGUMP object understands.
type SGFlag uint8

const (
	SGFlagNoClose       SGFlag = 0b00000001 // Prevent the GUMP from closing
	SGFlagNoMinimize    SGFlag = 0b00000010 // Does not support minimizing
	SGFlagNoPageButtons SGFlag = 0b00000100 // Do not generate page buttons
	SGFlagNoMove        SGFlag = 0b00001000 // Do not allow GUMP to be moved
)

// Standard theme constants
const (
	//
	// Configurable constants
	//

	sgCellWidth  int = 32
	sgCellHeight int = 32

	//
	// Scroll background
	//

	sgBackground           uo.GUMP = 5170
	sgBoarderWidth         int     = 38
	sgBoarderHeight        int     = 38
	sgBottomBoarderOverlap int     = 10
	sgTitleTextOffset      int     = 3

	//
	// Text
	//

	sgTextHue     uo.Hue = 1
	sgTextHOffset int    = 4
	sgTextVOffset int    = 4

	//
	// Scroll button skins
	//

	sgCloseButton    uo.GUMP = 5156
	sgMinimizeButton uo.GUMP = 5150
	sgPageUpButton   uo.GUMP = 5152
	sgPageDownButton uo.GUMP = 5158

	//
	// Reply and page buttons skin
	//

	sgButtonNormal  uo.GUMP = 2151
	sgButtonPressed uo.GUMP = 2152
	sgButtonHOffset int     = 1
	sgButtonVOffset int     = 1

	//
	// Checkbox and radio buttons skin
	//

	sgSwitchOff     uo.GUMP = 2151
	sgSwitchOn      uo.GUMP = 2154
	sgSwitchHOffset int     = 1
	sgSwitchVOffset int     = 1

	//
	// Text entry skin
	//

	sgTEBarLeft             uo.GUMP = 57
	sgTEBarMid              uo.GUMP = 58
	sgTEBarRight            uo.GUMP = 59
	sgTEBarWidth            int     = 16
	sgTEVOffset             int     = 10
	sgTELeftEndCapHOffset   int     = 2
	sgTEBarMidHeight        int     = 16
	sgTEEndCapApparentWidth int     = 15
	sgTEHeight              int     = 16

	//
	// Calculated constants
	//

	sgTotalHFill int = sgBoarderWidth * 2
	sgTotalVFill int = sgBoarderHeight*2 + sgBottomBoarderOverlap
	sgTop        int = sgBoarderHeight
	sgLeft       int = sgBoarderWidth

	//
	// Reply constants
	//

	sgReplyClose        uint32 = 0
	sgReplyMinimize     uint32 = 1
	sgReplyLastReserved uint32 = 1000 // User codes begin at 1001
)

// StandardGUMP provides a set of layout methods built on top of BaseGUMP's
// layout methods that make it a little more convenient to build GUMPs with the
// standard theme. The methods of this struct accept coordinates and dimensions
// in cells, not pixels. A cell is 32x32 pixels.
type StandardGUMP struct {
	// BaseGUMP we are managing
	g BaseGUMP
	// Width of the GUMP in cells
	w int
	// Height of the GUMP in cells
	h int
	// If true generate page buttons at the right
	pageButtons bool
	// Number of the last page that was created, 0 means none
	lastPage uint32
	// Number of the last radio group that was created, 0 means none
	lastGroup uint32
}

// InvalidateLayout implements the GUMP interface.
func (g *StandardGUMP) InvalidateLayout() {
	g.g.InvalidateLayout()
	g.lastPage = 0
	g.lastGroup = 0
}

// Packet implements the GUMP interface.
func (g *StandardGUMP) Packet(x, y int, sender, serial uo.Serial) serverpacket.Packet {
	return g.g.Packet(x, y, sender, serial)
}

// StandardReplyHandler returns true if the reply was totally handled by this
// function and should be ignored. This function alters the packet in ways
// required for the proper function of the GUMP, therefore this function must be
// called first thing in any HandleReply function.
func (g *StandardGUMP) StandardReplyHandler(p *clientpacket.GUMPReply) bool {
	if p.Button == sgReplyClose {
		// Should never reach this as the server is responsible for GUMP close
		// requests
		return true
	}
	if p.Button == sgReplyMinimize {
		// TODO Minimize functionality
		return true
	}
	if p.Button > sgReplyLastReserved {
		p.Button -= sgReplyLastReserved
		return false
	} else {
		log.Printf("Unknown standard reply button ID %d", p.Button)
		return true
	}
}

// Window creates a new Window for the GUMP. Window dimensions are given in
// cells and specifies the inner size of the window. If closable is true a close
// button is generated and the GUMP may be closed with right-clicking and - in
// some situations - the escape key. If minimizeable is true a minimize button
// is generated and the GUMP will have minimize behavior. If pageButtons is true
// page buttons are generated for every page.
func (g *StandardGUMP) Window(w, h int, title string, flags SGFlag) {
	g.w = w
	g.h = h
	// Flag handling
	g.pageButtons = true
	if flags&SGFlagNoPageButtons != 0 {
		g.pageButtons = false
	}
	if flags&SGFlagNoClose != 0 {
		g.g.NoClose()
		g.g.NoDispose()
	}
	if flags&SGFlagNoMove != 0 {
		g.g.NoMove()
	}
	g.g.NoResize()
	// Convert dimensions from cells
	if w < 1 || h < 1 {
		return
	}
	w *= sgCellWidth
	h *= sgCellHeight
	w += sgTotalHFill
	h += sgTotalVFill
	// Scroll background
	g.g.Background(0, 0, w, h, sgBackground)
	// Title area
	g.g.Label(sgLeft, sgTitleTextOffset, sgTextHue-1, title)
	// Standard UI elements
	if flags&SGFlagNoClose == 0 {
		y := sgTop + g.h*sgCellHeight
		g.g.ReplyButton(0, y, sgCloseButton, sgCloseButton, sgReplyClose)
	}
	if flags&SGFlagNoMinimize == 0 {
		g.g.ReplyButton(0, 0, sgMinimizeButton, sgMinimizeButton, sgReplyMinimize)
	}
}

// DebugCheckBackground creates a checked background to illustrate the size and
// position of each cell.
func (g *StandardGUMP) DebugCheckBackground() {
	for iy := 0; iy < g.h; iy++ {
		y := sgTop + iy*sgCellHeight
		dark := iy%2 == 0
		for ix := 0; ix < g.w; ix++ {
			x := sgLeft + ix*sgCellWidth
			if dark {
				g.g.TiledImage(x, y, sgCellWidth, sgCellHeight, 9354)
			} else {
				g.g.TiledImage(x, y, sgCellWidth, sgCellHeight, 9304)
			}
			dark = !dark
		}
	}
}

// Text creates a cropped label element. Text elements are always one line tall
// and do not respect newlines. If more functionality is needed see HTML().
func (g *StandardGUMP) Text(x, y, w int, hue uo.Hue, text string) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || w < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w *= sgCellWidth
	h := sgCellHeight
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	g.g.CroppedLabel(x+sgTextHOffset, y+sgTextVOffset, w, h, hue, text)
}

// ReplyButton creates a small gem button with the text to the right. The
// dimensions given are for the total width including the button. This means
// that the text is cropped to an area w-1xh. This button will generate a GUMP
// reply packet and close the GUMP.
func (g *StandardGUMP) ReplyButton(x, y, w, h int, hue uo.Hue, text string, id uint32) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || w < 2 || h < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w = (w - 1) * sgCellWidth
	h *= sgCellHeight
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	if id != 0 {
		id += sgReplyLastReserved
	}
	g.g.ReplyButton(x+sgButtonHOffset, y+sgButtonVOffset, sgButtonNormal, sgButtonPressed, id)
	x += sgCellWidth
	g.g.CroppedLabel(x+sgTextHOffset, y+sgTextVOffset, w, h, hue, text)
}

// PageButton creates a small gem button with the text to the right. The
// dimensions given are for the total width including the button. This means
// that the text is cropped to an area w-1xh. This button will hide the
// currently visible page and show the indicated page. No packets are generated.
// The standard scroll interface has page buttons included on the right-hand
// side of the scroll that can be used to sequentially cycle through the pages
// forward and backward. Only use this function if a different use case is
// required.
func (g *StandardGUMP) PageButton(x, y, w, h int, hue uo.Hue, text string, page uint32) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || w < 2 || h < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w = (w - 1) * sgCellWidth
	h *= sgCellHeight
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	g.g.PageButton(x+sgButtonHOffset, y+sgButtonVOffset, sgButtonNormal, sgButtonPressed, page)
	x += sgCellWidth
	g.g.CroppedLabel(x+sgTextHOffset, y+sgTextVOffset, w, h, hue, text)
}

// Page begins the numbered page. If g.pageButtons is true then this page will
// be linked to the current as the next.
func (g *StandardGUMP) Page(page uint32) {
	// PageDown button
	if g.pageButtons && g.lastPage != 0 {
		// Page 1 does not need any buttons to make it appear
		g.g.PageButton(sgLeft+g.w*sgCellWidth, sgTop+g.h*sgCellHeight,
			sgPageDownButton, sgPageDownButton, page)
	}
	// Page marker
	g.g.Page(page)
	// PageUp button
	if g.pageButtons && g.lastPage != 0 {
		// The first page does not need a page button back to 0
		g.g.PageButton(sgLeft+g.w*sgCellWidth, 0, sgPageUpButton,
			sgPageUpButton, g.lastPage)
	}
	g.lastPage = page
}

// HTML creates an HTML view without background.
func (g *StandardGUMP) HTML(x, y, w, h int, html string, scrollbar bool) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || w < 1 || h < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w *= sgCellWidth
	h *= sgCellHeight
	g.g.HTML(x, y, w, h, html, false, scrollbar)
}

// CheckSwitch creates a checkbox element with a numbered switch and a label.
func (g *StandardGUMP) CheckSwitch(x, y, w, h int, hue uo.Hue, text string, id uint32, checked bool) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || w < 2 || h < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w = (w - 1) * sgCellWidth
	h *= sgCellHeight
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	g.g.Checkbox(x+sgSwitchHOffset, y+sgSwitchVOffset, sgSwitchOff, sgSwitchOn, id, checked)
	x += sgCellWidth
	g.g.CroppedLabel(x+sgTextHOffset, y+sgTextVOffset, w, h, hue, text)
}

// Group starts a new RadioSwitch group.
func (g *StandardGUMP) Group() {
	if g.lastGroup > 0 {
		g.g.EndGroup()
	}
	g.lastGroup++
	g.g.Group(g.lastGroup)
}

// RadioSwitch creates a radio button element with a numbered switch and a label.
func (g *StandardGUMP) RadioSwitch(x, y, w, h int, hue uo.Hue, text string, id uint32, checked bool) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || w < 2 || h < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w = (w - 1) * sgCellWidth
	h *= sgCellHeight
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	g.g.RadioButton(x+sgSwitchHOffset, y+sgSwitchVOffset, sgSwitchOff, sgSwitchOn, id, checked)
	x += sgCellWidth
	g.g.CroppedLabel(x+sgTextHOffset, y+sgTextVOffset, w, h, hue, text)
}

// GemButton creates a standardized button from the gem theme with a word baked
// into the button image. The button is always two cells wide by one cell tall.
func (g *StandardGUMP) GemButton(x, y int, which SGGemButton, id uint32) {
	// Convert dimensions from cells
	if x < 0 || y < 0 || which >= SGGemButton(sgGemButtonCount) {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	if id != 0 {
		id += sgReplyLastReserved
	}
	d := sgGemDefs[which]
	x += d.h
	y += d.v
	g.g.ReplyButton(x, y, d.n, d.p, id)
}

// TextEntry creates a text entry element. Text entries are always one line
// tall and does not support newlines.
func (g *StandardGUMP) TextEntry(x, y, w int, hue uo.Hue, text string, limit int, id uint32) {
	bw := w - 2
	// Convert dimensions from cells
	if x < 0 || y < 0 || bw < 1 {
		return
	}
	x *= sgCellWidth
	x += sgLeft
	y *= sgCellHeight
	y += sgTop
	w *= sgCellWidth
	h := sgTEHeight
	// Create the bar
	bx := x
	bw *= sgCellWidth
	g.g.Image(bx+sgTELeftEndCapHOffset, y+sgTEVOffset, sgTEBarLeft, uo.HueDefault)
	bx += sgCellWidth
	g.g.TiledImage(bx, y+sgTEVOffset, bw, sgTEBarMidHeight, sgTEBarMid)
	bx += bw
	g.g.Image(bx, y+sgTEVOffset, sgTEBarRight, uo.HueDefault)
	// Create the text entry
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	x += sgTEEndCapApparentWidth
	w -= sgTEEndCapApparentWidth * 2
	g.g.TextEntry(x, y, w, h, hue, text, limit, id)
}
