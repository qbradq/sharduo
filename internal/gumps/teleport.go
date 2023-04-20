package gumps

import (
	"fmt"
	"log"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("teleport", func() GUMP {
		return &teleport{
			currentGroup: -1,
		}
	})
}

type teleportDestination struct {
	Name        string
	Destination uo.Location
}

type teleportGroup struct {
	Name         string
	Destinations []*teleportDestination
}

var teleportGroups []*teleportGroup

// teleport implements a simple menu that allows teleportation of game masters.
type teleport struct {
	StandardGUMP
	currentGroup int
}

// Layout implements the GUMP interface.
func (g *teleport) Layout(target, param game.Object) {
	if teleportGroups == nil {
		g.LoadTeleportData()
	}
	// Group and destination buttons start at 2000
	if g.currentGroup < 0 {
		g.Window(11, 11, "Global Teleport Menu", 0)
		page := uint32(0)
		iy := 0
		for i, grp := range teleportGroups {
			if i%10 == 0 {
				page++
				iy = 0
				g.Page(page)
			}
			g.ReplyButton(0, iy+1, 11, 1, uo.HueDefault, grp.Name, 2000+uint32(i))
			iy++
		}
	} else if g.currentGroup < len(teleportGroups) {
		grp := teleportGroups[g.currentGroup]
		g.Window(11, 11, fmt.Sprintf("Global Teleport Menu - %s", grp.Name), 0)
		// Back to groups button is ID 1000
		g.ReplyButton(0, 0, 11, 1, uo.HueDefault, "Back to Groups", 1000)
		page := uint32(0)
		iy := 0
		for i, dest := range grp.Destinations {
			if i%10 == 0 {
				page++
				iy = 0
				g.Page(page)
			}
			g.ReplyButton(0, iy+1, 11, 1, uo.HueDefault, dest.Name, 2000+uint32(i))
			iy++
		}
	} else {
		log.Printf("bad teleport group %d", g.currentGroup)
	}
}

// HandleReply implements the GUMP interface.
func (g *teleport) HandleReply(n game.NetState, p *clientpacket.GUMPReply) {
	// Standard behavior handling
	if g.StandardReplyHandler(p) {
		return
	}
	if p.Button == 1000 {
		g.currentGroup = -1
	}
	if p.Button >= 2000 {
		didx := int(p.Button - 2000)
		if g.currentGroup < 0 {
			g.currentGroup = didx
		} else if g.currentGroup < len(teleportGroups) {
			grp := teleportGroups[g.currentGroup]
			if didx < 0 || didx >= len(grp.Destinations) {
				log.Printf("bad teleport destination %d in group %d", didx, g.currentGroup)
				return
			}
			dest := grp.Destinations[didx]
			if n == nil || n.Mobile() == nil {
				return
			}
			game.GetWorld().Map().TeleportMobile(n.Mobile(), dest.Destination)
		} else {
			log.Printf("bad teleport group %d", g.currentGroup)
		}
	}
}

// LoadTeleportData loads all of the entries in data/misc/teleport-locations.ini
func (g *teleport) LoadTeleportData() {
	r := util.ListFileReader{}
	f, err := data.FS.Open("misc/teleport-locations.ini")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	r.StartReading(f)
	for {
		s := r.ReadNextSegment()
		if s == nil {
			if r.HasErrors() {
				for _, err := range r.Errors() {
					log.Println(err)
				}
			}
			break
		}
		g := &teleportGroup{
			Name: s.Name,
		}
		teleportGroups = append(teleportGroups, g)
		for _, line := range s.Contents {
			name, lstr, err := util.ParseTagLine(line)
			if err != nil {
				log.Println(err)
				continue
			}
			l, err := util.ParseLocation(lstr)
			if err != nil {
				log.Println(err)
				continue
			}
			g.Destinations = append(g.Destinations, &teleportDestination{
				Name:        name,
				Destination: l,
			})
		}
	}
}
