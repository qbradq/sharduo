package gumps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

type decorItem struct {
	name       string
	expression string
	hue        uo.Hue
}

type decorGroup struct {
	name  string
	items []decorItem
}

var rootDecorGroup decorGroup
var decorCatagories []decorGroup
var decorGroups []decorGroup

func init() {
	reg("statics", 0, func() GUMP {
		return &statics{}
	})
	var f = func(p string) []decorGroup {
		var lfr util.ListFileReader
		r, err := data.FS.Open(p)
		if err != nil {
			panic(fmt.Sprintf("error: reading file \"%s\"", p))
		}
		segs := lfr.ReadSegments(r)
		var ret []decorGroup
		for _, seg := range segs {
			g := decorGroup{
				name: seg.Name,
			}
			for _, s := range seg.Contents {
				parts := strings.Split(s, "=")
				if len(parts) != 2 {
					panic(fmt.Sprintf("error: processing decor file at line \"%s\"", s))
				}
				g.items = append(g.items, decorItem{
					name:       parts[0],
					expression: parts[1],
				})
			}
			ret = append(ret, g)
		}
		return ret
	}
	decorCatagories = f("misc/deco-catagories.ini")
	decorGroups = f("misc/deco-groups.ini")
	rootDecorGroup.name = "ROOT"
	for _, c := range decorCatagories {
		if len(c.items) < 1 {
			continue
		}
		rootDecorGroup.items = append(rootDecorGroup.items, decorItem{
			name:       c.name,
			expression: c.items[0].expression,
		})
	}
}

// statics implements a server-side building and decor system.
type statics struct {
	StandardGUMP
	depth    int       // Depth into the GUMP, 0=Catagories, 1=Groups, 2=Tiles
	category int       // Currently selected category index
	group    int       // Currently selected group index
	item     decorItem // Currently selected decor item
	tool     int       // Currently selected tool
}

// Layout implements the game.GUMP interface.
func (g *statics) Layout(target, param any) {
	fn := func(dg decorGroup) {
		for i := int(g.currentPage-1) * 12; i < len(dg.items) && i < int(g.currentPage)*12; i++ {
			item := dg.items[i]
			tx := i % 3
			ty := i / 3
			ty %= 4
			g.ReplyButton(tx*6+0, ty*5+0+7, 6, 1, 0, item.name, uint32(1001+i))
			g.Item(tx*6+2, ty*5+1+8, 0, 0, 0, uo.Graphic(util.RangeExpression(item.expression)))
		}
	}
	// Display grid
	switch g.depth {
	case 0:
		pages := len(rootDecorGroup.items) / 12
		if len(rootDecorGroup.items)%12 != 0 {
			pages++
		}
		g.Window(18, 27, "Statics", 0, uint32(pages))
		g.layoutCommonControls()
		fn(rootDecorGroup)
	case 1:
		c := decorCatagories[g.category]
		pages := len(c.items) / 12
		if len(c.items)%12 != 0 {
			pages++
		}
		g.Window(18, 27, "Statics", 0, uint32(pages))
		g.layoutCommonControls()
		g.ReplyButton(10, 0, 3, 1, 0, "Back", 101)
		fn(c)
	case 2:
		c := decorGroups[g.group]
		pages := len(c.items) / 12
		if len(c.items)%12 != 0 {
			pages++
		}
		g.Window(18, 27, "Statics", 0, uint32(pages))
		g.layoutCommonControls()
		g.ReplyButton(10, 0, 3, 1, 0, "Back", 101)
		fn(c)
	}
}

func (g *statics) layoutCommonControls() {
	g.Text(0, 0, 4, 0, "Current")
	g.Item(1, 1, 0, 0, g.item.hue,
		uo.Graphic(util.RangeExpression(g.item.expression)))
	g.ReplyButton(5, 0, 5, 1, 0, "Single Placement", 1)
	g.ReplyButton(5, 1, 5, 1, 0, "Fill Area", 2)
	g.ReplyButton(5, 2, 5, 1, 0, "Erase", 3)
	g.ReplyButton(5, 3, 5, 1, 0, "Erase Area", 4)
	g.ReplyButton(5, 4, 5, 1, 0, "Copy", 5)
}

// HandleReply implements the GUMP interface.
func (g *statics) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	// Misc reply buttons
	if p.Button == 101 {
		g.depth--
		g.currentPage = 1
		return
	}
	// Tool buttons
	if p.Button < 1001 {
		// Handle tool buttons
		g.tool = int(p.Button - 1)
		switch g.tool {
		case 0:
			g.placeSingle(n)
		case 1:
			g.areaFill(n)
		case 2:
			g.eraseSingle(n)
		case 3:
			g.eraseArea(n)
		case 4:
			g.lift(n)
		}
		return
	}
	selection := int(p.Button - 1001)
	switch g.depth {
	case 0:
		if selection >= len(decorCatagories) {
			return
		}
		g.category = selection
		g.depth++
		g.currentPage = 1
	case 1:
		c := decorCatagories[g.category]
		if selection >= len(c.items) {
			return
		}
		item := c.items[selection]
		for i, group := range decorGroups {
			if group.name == item.name {
				g.group = i
				g.depth++
				g.currentPage = 1
				return
			}
		}
	case 2:
		group := decorGroups[g.group]
		if selection >= len(group.items) {
			return
		}
		g.item.name = group.items[selection].name
		g.item.expression = group.items[selection].expression
		switch g.tool {
		case 0:
			g.placeSingle(n)
		case 1:
			g.areaFill(n)
		default:
			g.tool = 0
			g.placeSingle(n)
		}
	}
}

func (g *statics) lift(n game.NetState) {
	n.Speech(n.Mobile(), "Select the static you wish to copy")
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		a := n.GetGUMPByID(GUMPIDDecorate)
		if a == nil {
			return
		}
		d := a.(*decorate)
		if d == nil {
			return
		}
		if tr.TargetObject == uo.SerialZero {
			if tr.Graphic != uo.GraphicNone {
				g.item.expression = strconv.FormatInt(int64(tr.Graphic), 10)
				g.item.name = "static #" + strconv.FormatInt(int64(tr.Graphic), 10)
			}
			n.RefreshGUMP(g)
			g.lift(n)
			return
		}
		s := game.World.FindItem(tr.TargetObject)
		if s == nil {
			// Something wrong
			return
		}
		g.item.name = s.DisplayName()
		g.item.expression = strconv.FormatInt(int64(s.BaseGraphic()), 10)
		d.hue = s.Hue
		n.RefreshGUMP(g)
		n.RefreshGUMP(d)
		g.lift(n)
	})
}

func (g *statics) eraseArea(n game.NetState) {
	a := n.GetGUMPByID(GUMPIDDecorate)
	if a == nil {
		return
	}
	d, ok := a.(*decorate)
	if !ok {
		return
	}
	d.targetVolume(n, func(b uo.Bounds) {
		items := game.World.Map().ItemQueryByBounds("StaticItem", b, b.TopLeft())
		for _, item := range items {
			item.Remove()
		}
		g.eraseArea(n)
	})
}

func (g *statics) eraseSingle(n game.NetState) {
	n.Speech(n.Mobile(), "Select object to erase")
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if tr.TargetObject == uo.SerialZero {
			return
		}
		s := game.World.FindItem(tr.TargetObject)
		if s == nil {
			// Something wrong
			return
		}
		s.Remove()
		g.eraseSingle(n)
	})
}

func (g *statics) placeSingle(n game.NetState) {
	n.Speech(n.Mobile(), "Select destination")
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		a := n.GetGUMPByID(GUMPIDDecorate)
		if a == nil {
			return
		}
		d, ok := a.(*decorate)
		if !ok {
			return
		}
		d.place(tr.Location, g.item.expression, game.World.FindItem(tr.TargetObject))
		g.placeSingle(n)
	})
}

func (g *statics) areaFill(n game.NetState) {
	a := n.GetGUMPByID(GUMPIDDecorate)
	if a == nil {
		return
	}
	d, ok := a.(*decorate)
	if !ok {
		return
	}
	d.targetVolume(n, func(b uo.Bounds) {
		l := uo.Point{Z: b.Z}
		for l.Y = b.Y; l.Y <= b.South(); l.Y++ {
			for l.X = b.X; l.X <= b.East(); l.X++ {
				d.place(l, g.item.expression, nil)
			}
		}
		g.areaFill(n)

	})
}
