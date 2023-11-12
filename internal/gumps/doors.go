package gumps

import (
	"path"
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
	reg("doors", func() GUMP {
		return &doors{}
	})
	var lfr util.ListFileReader
	r, err := data.FS.Open(path.Join("misc", "doors.ini"))
	if err != nil {
		panic("error: reading file misc/doors.ini")
	}
	defer r.Close()
	segs := lfr.ReadSegments(r)
	if len(segs) != 1 || segs[0].Name != "Doors" {
		panic("error: malformed misc/doors.ini")
	}
	for _, s := range segs[0].Contents {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			panic("error: malformed misc/doors.ini")
		}
		v, err := strconv.ParseInt(parts[1], 0, 32)
		if err != nil {
			panic(err)
		}
		doorTypes = append(doorTypes, parts[0])
		doorIcons[parts[0]] = uo.Graphic(v)
	}
}

// List of door template names
var doorTypes []string

// Door template names to icon item mapping
var doorIcons = map[string]uo.Graphic{}

// English names of door orientations
var doorNames = []string{
	"West CW",
	"East CCW",
	"West CCW",
	"East CW",
	"South CW",
	"North CCW",
	"South CCW",
	"North CW",
}

// GUMP graphics for door swing indicators
var doorIndicators = []uo.GUMP{
	0x5787,
	0x5785,
	0x5786,
	0x5784,
	0x5782,
	0x5780,
	0x5783,
	0x5781,
}

// doors implements a door placement menu
type doors struct {
	StandardGUMP
	doorType string // Currently selected door type, or an empty string if on the type selection menu.
	facing   int    // Currently selected door facing
}

// Layout implements the game.GUMP interface.
func (g *doors) Layout(target, param game.Object) {
	g.Window(10, 25, "Door Placement", 0)
	g.Page(0)
	if g.doorType == "" {
		var page uint32
		for i, s := range doorTypes {
			if i%8 == 0 {
				page++
				g.Page(page)
			}
			tx := i % 2
			ty := (i / 2) % 4
			g.ReplyButton(tx*5, ty*6+1+0, 5, 1, uo.HueDefault, s, uint32(1001+i))
			g.Item(tx*5+1, ty*6+1+1, 0, 0, uo.HueDefault, doorIcons[s])
		}
	} else {
		g.ReplyButton(7, 0, 3, 1, uo.HueDefault, "Back", 1)
		baseGraphic := doorIcons[g.doorType]
		for i := 0; i < 8; i++ {
			tx := i % 2
			ty := i / 2
			g.ReplyButton(tx*5+0, ty*6+1+0, 5, 1, uo.HueDefault, doorNames[i], uint32(2001+i))
			g.Item(tx*5+1, ty*6+1+1, 0, 0, uo.HueDefault, baseGraphic+uo.Graphic(i*2))
			g.Image(tx*5+1, ty*6+1+4, 0, 0, uo.HueDefault, doorIndicators[i])
		}
	}
}

// HandleReply implements the GUMP interface.
func (g *doors) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	if g.StandardReplyHandler(p) {
		return
	}
	// Control buttons
	switch p.Button {
	case 1:
		g.doorType = ""
		return
	}
	if p.Button >= 2001 {
		// Facing selection
		facing := int(p.Button - 2001)
		if facing > 7 {
			return
		}
		g.facing = facing
		g.singlePlacement(n)
		return
	} else if p.Button >= 1001 {
		// Door type selection
		dt := int(p.Button - 1001)
		if dt >= len(doorTypes) {
			return
		}
		g.doorType = doorTypes[dt]
		return
	}
}

func (g *doors) singlePlacement(n game.NetState) {
	n.TargetSendCursor(uo.TargetTypeLocation, func(tr *clientpacket.TargetResponse) {
		door := template.Create[game.Item](g.doorType)
		if door == nil {
			// Something wrong
			return
		}
		door.SetBaseGraphic(door.BaseGraphic() + uo.Graphic(g.facing*2))
		door.SetFlippedGraphic(door.FlippedGraphic() + uo.Graphic(g.facing*2))
		door.SetLocation(tr.Location)
		door.SetFacing(uo.Direction(g.facing))
		game.GetWorld().Map().ForceAddObject(door)
	})
}
