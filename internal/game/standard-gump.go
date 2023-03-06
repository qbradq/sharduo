package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// StandardGUMP provides a set of layout methods built on top of BaseGUMP's
// layout methods that make it a little more convenient to build GUMPs with the
// standard theme. The methods of this struct accept coordinates and dimensions
// in cells, not pixels. A cell is 32x32 pixels.

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
	sgTextHue              uo.Hue  = 1
	sgBoarderWidth         int     = 38
	sgBoarderHeight        int     = 38
	sgBottomBoarderOverlap int     = 10
	sgTitleTextOffset      int     = 3

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
	// Calculated constants
	//

	sgTotalHFill int = sgBoarderWidth * 2
	sgTotalVFill int = sgBoarderHeight*2 + sgBottomBoarderOverlap
	sgTop        int = sgBoarderHeight
	sgLeft       int = sgBoarderWidth

	//
	// Reply constants
	//

	SGReplyClose        uint32 = 0
	SGReplyMinimize     uint32 = 1
	SGReplyLastReserved uint32 = 1000 // User codes begin at 1001
)

type StandardGUMP struct {
	// BaseGUMP we are managing
	g BaseGUMP
	// Width of the GUMP in cells
	w int
	// Height of the GUMP in cells
	h int
	// If true generate page buttons at the right
	pageButtons bool
	// Number of the last page that was created
	lastPage uint32
}

// Packet implements the GUMP interface.
func (g *StandardGUMP) Packet(x, y int, id, serial uo.Serial) serverpacket.Packet {
	return g.g.Packet(x, y, id, serial)
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
		g.g.ReplyButton(0, y, sgCloseButton, sgCloseButton, SGReplyClose)
	}
	if flags&SGFlagNoMinimize == 0 {
		g.g.ReplyButton(0, 0, sgMinimizeButton, sgMinimizeButton, SGReplyMinimize)
	}
}

// Text creates a cropped label element.
func (g *StandardGUMP) Text(x, y, w, h int, hue uo.Hue, text string) {
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
	if hue == uo.HueDefault {
		hue = sgTextHue
	}
	hue -= 1
	g.g.CroppedLabel(x, y, w, h, hue, text)
}

// ReplyButton creates a small gem button with the text to the right. The
// dimensions given are for the total width including the button. This means
// that the text is cropped to an area w-1xh. This button will generate a GUMP
// reply packet and close the GUMP.
func (g *StandardGUMP) ReplyButton(x, y, w, h int, hue uo.Hue, text string, id uint32) {
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
	g.g.ReplyButton(x, y, sgButtonNormal, sgButtonPressed, id)
	x += sgCellWidth
	g.g.CroppedLabel(x, y, w, h, hue, text)
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
	g.g.PageButton(x, y, sgButtonNormal, sgButtonPressed, page)
	x += sgCellWidth
	g.g.CroppedLabel(x, y, w, h, hue, text)
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
