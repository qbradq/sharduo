package commands

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Developer commands go here, generally these should not be used in production

func init() {
	regcmd(&cmdesc{"decorate", []string{"deco"}, commandDecorate, game.RoleDeveloper, "decorate", "Calls up the decoration GUMP"})
	regcmd(&cmdesc{"loadspawners", nil, commandLoadSpawners, game.RoleDeveloper, "loadspawners", "Clears all spawners in the world, then loads data/misc/spawners.ini and fully respawns all spawners"})
	regcmd(&cmdesc{"loadstatics", nil, commandLoadStatics, game.RoleDeveloper, "loadstatics", "Clears all statics in the world, then loads data/misc/statics.csv"})
	regcmd(&cmdesc{"savespawners", nil, commandSaveSpawners, game.RoleDeveloper, "savespawners", "Generates data/misc/spawners.ini"})
	regcmd(&cmdesc{"savestatics", nil, commandSaveStatics, game.RoleDeveloper, "savestatics", "Generates data/misc/statics.csv"})
}

func commandLoadSpawners(n game.NetState, args CommandArgs, cl string) {
	broadcast("Load Spawners: clearing all spawners")
	for _, s := range game.GetWorld().Map().ItemQuery("Spawner", uo.BoundsZero) {
		game.Remove(s)
	}
	broadcast("Load Spawners: loading spawners.ini")
	f, err := data.FS.Open(path.Join("misc", "spawners.ini"))
	if err != nil {
		broadcast("Load Spawners: error loading spawners.ini: %s", err.Error())
		return
	}
	defer f.Close()
	lfr := util.ListFileReader{}
	for _, lfs := range lfr.ReadSegments(f) {
		s := template.Create[*game.Spawner]("Spawner")
		if s == nil {
			// Something very wrong
			return
		}
		s.Read(lfs)
		game.GetWorld().Map().ForceAddObject(s)
		s.FullRespawn()
	}
}

func commandSaveSpawners(n game.NetState, args CommandArgs, cl string) {
	f, err := os.Create(path.Join("data", "misc", "spawners.ini"))
	if err != nil {
		broadcast("Save Spawners: error opening spawners.ini: %s", err.Error())
		return
	}
	defer f.Close()
	broadcast("Save Spawners: getting spawners")
	spawners := game.GetWorld().Map().ItemQuery("Spawner", uo.BoundsZero)
	broadcast("Save Spawners: sorting spawners")
	sort.Slice(spawners, func(i, j int) bool {
		a := spawners[i].Location()
		b := spawners[j].Location()
		if a.Y < b.Y {
			return true
		} else if a.Y == b.Y {
			if a.X < b.X {
				return true
			} else if a.X == b.X {
				return a.Z < b.Z
			}
			return false
		}
		return false
	})
	broadcast("Save Spawners: writing spawners")
	for _, item := range spawners {
		if s, ok := item.(*game.Spawner); ok {
			s.Write(f)
		}
	}
	broadcast("Save Spawners: complete")
}

func commandLoadStatics(n game.NetState, args CommandArgs, cl string) {
	var fn = func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(v)
	}
	broadcast("Load Statics: clearing all statics")
	for _, s := range game.GetWorld().Map().ItemQuery("StaticItem", uo.BoundsZero) {
		game.Remove(s)
	}
	broadcast("Load Statics: loading statics.csv")
	f, err := data.FS.Open(path.Join("misc", "statics.csv"))
	if err != nil {
		broadcast("Load Statics: error loading statics.csv: %s", err.Error())
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comment = ';'
	r.FieldsPerRecord = 5
	r.ReuseRecord = true
	broadcast("Load Statics: generating new statics")
	for {
		fields, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			broadcast("Load Statics: error reading statics.csv: %s", err.Error())
			return
		}
		s := template.Create[*game.StaticItem]("StaticItem")
		s.SetLocation(uo.Location{
			X: int16(fn(fields[0])),
			Y: int16(fn(fields[1])),
			Z: int8(fn(fields[2])),
		})
		s.SetBaseGraphic(uo.Graphic(fn(fields[3])))
		s.SetHue(uo.Hue(fn(fields[4])))
		game.GetWorld().Map().ForceAddObject(s)
	}
	broadcast("Load Statics: complete")
}

func commandSaveStatics(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Statics: getting statics")
	statics := game.GetWorld().Map().ItemQuery("StaticItem", uo.BoundsZero)
	broadcast("Save Statics: sorting statics")
	sort.Slice(statics, func(i, j int) bool {
		a := statics[i].Location()
		b := statics[j].Location()
		if a.Y < b.Y {
			return true
		} else if a.Y == b.Y {
			if a.X < b.X {
				return true
			} else if a.X == b.X {
				return a.Z < b.Z
			}
			return false
		}
		return false
	})
	broadcast("Save Statics: writing statics")
	f, err := os.Create(path.Join("data", "misc", "statics.csv"))
	if err != nil {
		broadcast(fmt.Sprintf("Error generating statics.csv: %s", err.Error()))
		return
	}
	defer f.Close()
	f.WriteString(";X,Y,Z,Graphic,Hue\n")
	for _, s := range statics {
		l := s.Location()
		f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n",
			l.X, l.Y, l.Z,
			s.BaseGraphic(),
			s.Hue(),
		))
	}
	broadcast("Save Statics complete")
}

func commandDecorate(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("decorate"), nil, nil)
}
