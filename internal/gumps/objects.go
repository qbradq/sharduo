package gumps

import (
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	fn := func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(v)
	}
	reg("objects", func() GUMP {
		return &objects{}
	})
	var lfr util.ListFileReader
	r, err := data.FS.Open(path.Join("misc", "objects.ini"))
	if err != nil {
		panic("error: reading file misc/objects.ini")
	}
	defer r.Close()
	for _, s := range lfr.ReadSegments(r) {
		d := &objectDefinition{
			Name: s.Name,
		}
		for _, line := range s.Contents {
			parts := strings.Split(line, ",")
			if len(parts) != 5 {
				panic("error: malformed misc/objects.ini")
			}
			d.Tiles = append(d.Tiles, objectTile{
				Offset: uo.Location{
					X: int16(fn(parts[0])),
					Y: int16(fn(parts[1])),
					Z: int8(fn(parts[2])),
				},
				Graphic: uo.Graphic(fn(parts[3])),
				Hue:     uo.Hue(fn(parts[4])),
			})
		}
		objectDefinitions = append(objectDefinitions, d)
	}
	sort.Slice(objectDefinitions, func(i, j int) bool {
		a := objectDefinitions[i]
		b := objectDefinitions[j]
		return strings.Compare(a.Name, b.Name) < 0
	})
}

// objectTile defines a single tile of a multi-tile object.
type objectTile struct {
	Offset  uo.Location // Offset of the tile from the object origin
	Graphic uo.Graphic  // Tile graphic to use
	Hue     uo.Hue      // Tile hue
}

// objectDefinition defines a multi-tile object.
type objectDefinition struct {
	Name  string       // Name of the object from objects.ini
	Tiles []objectTile // Tiles of the object
}

var objectDefinitions []*objectDefinition

// Objects implements a text-only menu that allows placement of multi-tile
// statics.
type objects struct {
	StandardGUMP
	useFixedZ      bool               // If true we force the Z value to fixedZ
	fixedZ         int8               // Forced Z value if useFixedZ is true
	idx            int                // Currently selected object index
	lastPlacements []*game.StaticItem // All of the items created in the previous operation
}

// Layout implements the game.GUMP interface.
func (g *objects) Layout(target, param game.Object) {
	g.Window(10, 22, "Multi-Tile Object Placement", 0)
	g.Page(0)
	g.ReplyButton(0, 0, 3, 1, uo.HueDefault, "Undo", 1)
	g.CheckSwitch(3, 0, 3, 1, uo.HueDefault, "Fixed Z", 2, false)
	g.TextEntry(6, 0, 4, uo.HueDefault, strconv.FormatInt(int64(g.fixedZ), 10), 4, 3)
	g.HorizontalBar(0, 1, 10)
	page := uint32(0)
	for i, d := range objectDefinitions {
		if i%20 == 0 {
			page++
			g.Page(page)
		}
		ty := i % 20
		g.ReplyButton(0, ty+2, 10, 1, uo.HueDefault, d.Name, uint32(1001+i))
	}
}

// HandleReply implements the GUMP interface.
func (g *objects) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	// Data
	g.useFixedZ = p.Switch(2)
	z, err := strconv.ParseInt(p.Text(3), 0, 32)
	if err == nil {
		g.fixedZ = int8(z)
	}
	// Tool buttons
	switch p.Button {
	case 1:
		// Undo
		for _, s := range g.lastPlacements {
			game.Remove(s)
		}
		g.lastPlacements = nil
		return
	}
	if p.Button >= 1001 {
		idx := int(p.Button - 1001)
		if idx >= len(objectDefinitions) {
			return
		}
		g.idx = idx
		g.placementTarget(n)
	}
}

func (g *objects) placementTarget(n game.NetState) {
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		g.place(tr.Location)
		g.placementTarget(n)
	})
}

func (g *objects) place(l uo.Location) {
	if g.useFixedZ {
		l.Z = g.fixedZ
	} else {
		f, _ := game.GetWorld().Map().GetFloorAndCeiling(l, false, false)
		if f != nil {
			l.Z = f.Highest()
		}
	}
	g.lastPlacements = nil
	def := objectDefinitions[g.idx]
	for _, t := range def.Tiles {
		s := template.Create[*game.StaticItem]("StaticItem")
		nl := l
		nl.X += t.Offset.X
		nl.Y += t.Offset.Y
		nl.Z += t.Offset.Z
		s.SetBaseGraphic(t.Graphic)
		s.SetLocation(nl)
		game.GetWorld().Map().ForceAddObject(s)
		g.lastPlacements = append(g.lastPlacements, s)
	}
}
