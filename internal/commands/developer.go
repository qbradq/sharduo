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
)

// Developer commands go here, generally these should not be used in production

func init() {
	regcmd(&cmdesc{"decorate", []string{"deco"}, commandDecorate, game.RoleDeveloper, "decorate", "Calls up the decoration GUMP"})
	regcmd(&cmdesc{"loadstatics", nil, commandLoadStatics, game.RoleDeveloper, "loadstatics", "Clears all statics in the world, then loads data/misc/statics.csv"})
	regcmd(&cmdesc{"savestatics", nil, commandSaveStatics, game.RoleDeveloper, "savestatics", "Generates data/misc/statics.csv"})
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
	r.FieldsPerRecord = 6
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
			X: int16(fn(fields[1])),
			Y: int16(fn(fields[2])),
			Z: int8(fn(fields[3])),
		})
		s.SetBaseGraphic(uo.Graphic(fn(fields[4])))
		s.SetHue(uo.Hue(fn(fields[5])))
		game.GetWorld().Map().ForceAddObject(s)
	}
	broadcast("Load Statics: complete")
}

func commandSaveStatics(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Statics: querying statics...")
	statics := game.GetWorld().Map().ItemQuery("StaticItem", uo.BoundsZero)
	broadcast("Save Statics: sorting statics...")
	sort.Slice(statics, func(i, j int) bool {
		return statics[i].Serial() < statics[j].Serial()
	})
	broadcast("Save Statics: writing statics...")
	f, err := os.Create(path.Join("data", "misc", "statics.csv"))
	if err != nil {
		broadcast(fmt.Sprintf("Error generating statics.csv: %s", err.Error()))
		return
	}
	defer f.Close()
	f.WriteString(";Serial,X,Y,Z,Graphic,Hue\n")
	for _, s := range statics {
		l := s.Location()
		f.WriteString(fmt.Sprintf("%s,%d,%d,%d,%d,%d\n",
			s.Serial().String(),
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
