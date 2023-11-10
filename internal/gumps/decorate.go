package gumps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

type decorItem struct {
	name       string
	expression string
}

type decorGroup struct {
	name  string
	items []decorItem
}

var rootDecorGroup decorGroup
var decorCatagories []decorGroup
var decorGroups []decorGroup

func init() {
	reg("decorate", func() GUMP {
		return &decorate{}
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

// decorate implements a server-side building and decor system.
type decorate struct {
	StandardGUMP
	depth     int       // Depth into the GUMP, 0=Catagories, 1=Groups, 2=Tiles
	category  int       // Currently selected category index
	group     int       // Currently selected group index
	item      decorItem // Currently selected decor item
	tool      int       // Currently selected tool
	useFixedZ bool      // If true use the absolute Z value in the entry field
	fixedZ    int8      // The absolute Z value to use if useFixedZ == true
}

// Layout implements the game.GUMP interface.
func (g *decorate) Layout(target, param game.Object) {
	var f = func(dg decorGroup) {
		page := 0
		for idx, item := range dg.items {
			if idx%15 == 0 {
				page++
				g.Page(uint32(page))
			}
			x := idx % 3
			y := idx / 3
			y %= 5
			g.ReplyButton(x*6+0, y*5+0+5, 6, 1, 0, item.name, uint32(1001+idx))
			g.Item(x*6+2, y*5+1+5, 0, 0, 0, uo.Graphic(util.RangeExpression(item.expression, game.GetWorld().Random())))
		}
	}
	g.Window(18, 30, "Decoration", 0)
	g.Page(0)
	// Controls and status stuff goes here
	g.Text(0, 0, 4, 0, "Current")
	g.Item(1, 1, 0, 0, 0, uo.Graphic(util.RangeExpression(g.item.expression, game.GetWorld().Random())))
	g.ReplyButton(5, 0, 5, 1, 0, "Single Placement", 1)
	g.ReplyButton(5, 1, 5, 1, 0, "Fill Area", 2)
	g.ReplyButton(5, 2, 5, 1, 0, "Erase", 3)
	g.ReplyButton(5, 3, 5, 1, 0, "Erase Area", 4)
	g.ReplyButton(5, 4, 5, 1, 0, "Eyedropper", 5)
	g.CheckSwitch(10, 1, 5, 1, uo.HueDefault, "Fixed-Z:", 6, g.useFixedZ)
	g.TextEntry(15, 1, 3, uo.HueDefault, strconv.Itoa(int(g.fixedZ)), 4, 7)
	g.ReplyButton(10, 2, 5, 1, uo.HueDefault, "Door Placement", 8)
	g.ReplyButton(10, 3, 5, 1, uo.HueDefault, "Sign Placement", 9)
	// Display grid
	switch g.depth {
	case 0:
		f(rootDecorGroup)
	case 1:
		g.ReplyButton(10, 0, 5, 1, 0, "Back", 101)
		f(decorCatagories[g.category])
	case 2:
		g.ReplyButton(10, 0, 5, 1, 0, "Back", 101)
		f(decorGroups[g.group])
	}
}

// HandleReply implements the GUMP interface.
func (g *decorate) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	v, err := strconv.ParseInt(p.Text(7), 0, 32)
	if err == nil {
		g.fixedZ = int8(v)
	}
	g.useFixedZ = p.Switch(6)
	if g.StandardReplyHandler(p) {
		return
	}
	// Misc reply buttons
	switch p.Button {
	case 8:
		n.GUMP(New("doors"), nil, nil)
	case 9:
		n.GUMP(New("signs"), nil, nil)
	case 101:
		g.depth--
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
				return
			}
		}
	case 2:
		group := decorGroups[g.group]
		if selection >= len(group.items) {
			return
		}
		g.item = group.items[selection]
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

func (g *decorate) lift(n game.NetState) {
	n.Speech(n.Mobile(), "Select the static you wish to copy")
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if tr.TargetObject == uo.SerialZero {
			if tr.Graphic != uo.GraphicNone {
				g.item.expression = strconv.FormatInt(int64(tr.Graphic), 10)
				g.item.name = "static #" + strconv.FormatInt(int64(tr.Graphic), 10)
			}
			return
		}
		s := game.Find[*game.StaticItem](tr.TargetObject)
		if s == nil {
			// Something wrong
			return
		}
		g.item.name = s.DisplayName()
		g.item.expression = strconv.FormatInt(int64(s.BaseGraphic()), 10)
		g.lift(n)
	})
}

func (g *decorate) eraseArea(n game.NetState) {
	n.Speech(n.Mobile(), "Select the start of the area")
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		start := tr.Location
		n.Speech(n.Mobile(), "Select the end of the area")
		n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
			end := tr.Location
			bounds := uo.BoundsOf(start, end)
			items := game.GetWorld().Map().ItemQuery("StaticItem", bounds)
			for _, item := range items {
				game.Remove(item)
			}
			g.eraseArea(n)
		})
	})
}

func (g *decorate) eraseSingle(n game.NetState) {
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if tr.TargetObject == uo.SerialZero {
			return
		}
		s := game.Find[*game.StaticItem](tr.TargetObject)
		if s == nil {
			// Something wrong
			return
		}
		game.Remove(s)
		g.eraseSingle(n)
	})
}

func (g *decorate) place(l uo.Location, exp string) bool {
	item := template.Create[*game.StaticItem]("StaticItem")
	if item == nil {
		// Something very wrong
		return false
	}
	item.SetBaseGraphic(uo.Graphic(util.RangeExpression(g.item.expression, game.GetWorld().Random())))
	if g.useFixedZ {
		l.Z = g.fixedZ
	} else {
		f, c := game.GetWorld().Map().GetFloorAndCeiling(l, true, false)
		if f != nil {
			l.Z = f.Highest()
		}
		if c != nil {
			if int(c.Z())-int(f.Highest()) < int(item.Height()) {
				// Not enough room to fit the static within the other statics in
				// that location, refuse to place the object.
				return false
			}
		}
	}
	item.SetLocation(l)
	game.GetWorld().Map().ForceAddObject(item)
	return true
}

func (g *decorate) placeSingle(n game.NetState) {
	n.Speech(n.Mobile(), "Select destination")
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		g.place(tr.Location, g.item.expression)
		g.placeSingle(n)
	})
}

func (g *decorate) areaFill(n game.NetState) {
	n.Speech(n.Mobile(), "Select starting location")
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		start := tr.Location
		n.Speech(n.Mobile(), "Select ending location")
		n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
			end := tr.Location
			bounds := uo.BoundsOf(start, end)
			l := uo.Location{Z: bounds.Z}
			for l.Y = bounds.Y; l.Y < bounds.Y+bounds.H; l.Y++ {
				for l.X = bounds.X; l.X < bounds.X+bounds.W; l.X++ {
					g.place(l, g.item.expression)
				}
			}
			g.areaFill(n)
		})
	})
}
