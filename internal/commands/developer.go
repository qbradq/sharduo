package commands

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

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
	regcmd(&cmdesc{"loaddoors", nil, commandLoadDoors, game.RoleDeveloper, "loaddoors", "Clears all doors then loads data/misc/doors.csv"})
	regcmd(&cmdesc{"loadregions", nil, commandLoadRegions, game.RoleDeveloper, "loadregions", "Clears all regions then loads data/misc/regions.csv"})
	regcmd(&cmdesc{"loadsigns", nil, commandLoadSigns, game.RoleDeveloper, "loadsigns", "Clears all signs then loads data/misc/signs.csv"})
	regcmd(&cmdesc{"loadstatics", nil, commandLoadStatics, game.RoleDeveloper, "loadstatics", "Clears all statics then loads data/misc/statics.csv"})
	regcmd(&cmdesc{"regions", nil, commandRegions, game.RoleDeveloper, "regions", "Calls up the regions GUMP"})
	regcmd(&cmdesc{"respawn", nil, commandRespawn, game.RoleDeveloper, "respawn", "Executes a full respawn on all spawning regions"})
	regcmd(&cmdesc{"savedoors", nil, commandSaveDoors, game.RoleDeveloper, "savedoors", "Generates data/misc/doors.csv"})
	regcmd(&cmdesc{"saveregions", nil, commandSaveRegions, game.RoleDeveloper, "saveregions", "Generates data/misc/regions.csv"})
	regcmd(&cmdesc{"savesigns", nil, commandSaveSigns, game.RoleDeveloper, "savesigns", "Generates data/misc/signs.csv"})
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
			s.Graphic(),
			s.Hue(),
		))
	}
	broadcast("Save Statics complete")
}

func commandDecorate(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("decorate"), nil, nil)
}

func commandSaveDoors(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Doors: getting doors")
	doors := game.GetWorld().Map().ItemBaseQuery("BaseDoor", uo.BoundsZero)
	broadcast("Save Doors: sorting doors")
	// Correct door locations for doors that happen to be open
	for _, d := range doors {
		if d.Flipped() {
			l := d.Location()
			ofs := uo.DoorOffsets[d.Facing()]
			l.X -= ofs.X
			l.Y -= ofs.Y
			game.GetWorld().Map().ForceRemoveObject(d)
			d.SetLocation(l)
			d.Flip()
			d.SetDefForGraphic(d.Graphic())
			game.GetWorld().Map().ForceAddObject(d)
		}
	}
	// Sort based on corrected locations
	sort.Slice(doors, func(i, j int) bool {
		a := doors[i].Location()
		b := doors[j].Location()
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
	broadcast("Save Doors: writing doors")
	f, err := os.Create(path.Join("data", "misc", "doors.csv"))
	if err != nil {
		broadcast(fmt.Sprintf("Error generating doors.csv: %s", err.Error()))
		return
	}
	defer f.Close()
	f.WriteString(";X,Y,Z,TemplateName,Facing\n")
	for _, d := range doors {
		l := d.Location()
		f.WriteString(fmt.Sprintf("%d,%d,%d,%s,%d\n",
			l.X, l.Y, l.Z,
			d.TemplateName(),
			d.Facing(),
		))
	}
	broadcast("Save Doors complete")
}

func commandLoadDoors(n game.NetState, args CommandArgs, cl string) {
	var fn = func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(v)
	}
	broadcast("Load Doors: clearing all doors")
	for _, s := range game.GetWorld().Map().ItemBaseQuery("BaseDoor", uo.BoundsZero) {
		game.Remove(s)
	}
	broadcast("Load Doors: loading doors.csv")
	f, err := data.FS.Open(path.Join("misc", "doors.csv"))
	if err != nil {
		broadcast("Load Doors: error loading doors.csv: %s", err.Error())
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comment = ';'
	r.FieldsPerRecord = 5
	r.ReuseRecord = true
	broadcast("Load Doors: generating new doors")
	for {
		fields, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			broadcast("Load Doors: error reading doors.csv: %s", err.Error())
			return
		}
		d := template.Create[game.Item](fields[3])
		d.SetLocation(uo.Location{
			X: int16(fn(fields[0])),
			Y: int16(fn(fields[1])),
			Z: int8(fn(fields[2])),
		})
		d.SetFacing(uo.Direction(fn(fields[4])))
		d.SetBaseGraphic(d.BaseGraphic() + uo.Graphic(d.Facing()*2))
		d.SetFlippedGraphic(d.FlippedGraphic() + uo.Graphic(d.Facing()*2))
		game.GetWorld().Map().ForceAddObject(d)
	}
	broadcast("Load Doors: complete")
}

func commandSaveSigns(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Signs: getting signs")
	signs := game.GetWorld().Map().ItemQuery("BaseSign", uo.BoundsZero)
	broadcast("Save Signs: sorting signs")
	// Sort based on location
	sort.Slice(signs, func(i, j int) bool {
		a := signs[i].Location()
		b := signs[j].Location()
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
	broadcast("Save Signs: writing signs")
	f, err := os.Create(path.Join("data", "misc", "signs.csv"))
	if err != nil {
		broadcast(fmt.Sprintf("Error generating signs.csv: %s", err.Error()))
		return
	}
	defer f.Close()
	f.WriteString(";X,Y,Z,Graphic,\"Text\"\n")
	for _, s := range signs {
		l := s.Location()
		f.WriteString(fmt.Sprintf("%d,%d,%d,%d,\"%s\"\n",
			l.X, l.Y, l.Z,
			s.BaseGraphic(),
			s.Name(),
		))
	}
	broadcast("Save Signs complete")
}

func commandLoadSigns(n game.NetState, args CommandArgs, cl string) {
	var fn = func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(v)
	}
	broadcast("Load Signs: clearing all signs")
	for _, s := range game.GetWorld().Map().ItemBaseQuery("BaseSign", uo.BoundsZero) {
		game.Remove(s)
	}
	broadcast("Load Signs: loading signs.csv")
	f, err := data.FS.Open(path.Join("misc", "signs.csv"))
	if err != nil {
		broadcast("Load Signs: error loading signs.csv: %s", err.Error())
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comment = ';'
	r.FieldsPerRecord = 5
	r.ReuseRecord = true
	broadcast("Load Signs: generating new signs")
	for {
		fields, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			broadcast("Load Signs: error reading signs.csv: %s", err.Error())
			return
		}
		s := template.Create[game.Item]("BaseSign")
		s.SetLocation(uo.Location{
			X: int16(fn(fields[0])),
			Y: int16(fn(fields[1])),
			Z: int8(fn(fields[2])),
		})
		s.SetBaseGraphic(uo.Graphic(fn(fields[3])))
		s.SetName(fields[4])
		game.GetWorld().Map().ForceAddObject(s)
	}
	broadcast("Load Doors: complete")
}

func commandRegions(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("regions"), n.Mobile(), nil)
}

func commandSaveRegions(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Regions: getting and sorting regions")
	regions := game.GetWorld().Map().RegionsWithin(uo.BoundsFullMap)
	sort.Slice(regions, func(i, j int) bool {
		a := regions[i]
		b := regions[j]
		if a.Bounds.X < b.Bounds.X {
			return true
		}
		if a.Bounds.X > b.Bounds.X {
			return false
		}
		if a.Bounds.Y < b.Bounds.Y {
			return true
		}
		if a.Bounds.Y > b.Bounds.Y {
			return false
		}
		if a.Bounds.W < b.Bounds.W {
			return true
		}
		if a.Bounds.W > b.Bounds.W {
			return false
		}
		if a.Bounds.H < b.Bounds.H {
			return true
		}
		return a.Bounds.H > b.Bounds.H
	})
	broadcast("Save Regions: generating regions.ini")
	f, err := os.Create(path.Join("data", "misc", "regions.ini"))
	if err != nil {
		broadcast(fmt.Sprintf("Save Regions: Error creating regions.ini %s", err.Error()))
		return
	}
	defer f.Close()
	f.WriteString("; This file generated by saveregions command\n")
	for _, r := range regions {
		f.WriteString("\n")
		f.WriteString("[Region]\n")
		f.WriteString(fmt.Sprintf("Name=%s\n", r.Name))
		f.WriteString(fmt.Sprintf("Music=%s\n", r.Music))
		f.WriteString(fmt.Sprintf("Features=0x%04X\n", r.Features))
		f.WriteString(fmt.Sprintf("SpawnMinZ=%d\n", r.SpawnMinZ))
		f.WriteString(fmt.Sprintf("SpawnMaxZ=%d\n", r.SpawnMaxZ))
		for _, rect := range r.Rects {
			f.WriteString(fmt.Sprintf("Rect=%d,%d,%d,%d\n", rect.X, rect.Y, rect.W, rect.H))
		}
		for _, e := range r.Entries {
			f.WriteString(fmt.Sprintf("Spawn=%d,%d,%s\n", e.Amount, e.Delay, e.Template))
		}
	}
	broadcast("Save Regions: complete")
}

func commandLoadRegions(n game.NetState, args CommandArgs, cl string) {
	fn := func(s string) int {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(v)
	}
	broadcast("Load Regions: clearing all regions")
	regions := game.GetWorld().Map().RegionsWithin(uo.BoundsFullMap)
	for _, r := range regions {
		game.GetWorld().Map().RemoveRegion(r)
	}
	broadcast("Load Regions: loading regions.ini")
	f, err := data.FS.Open(path.Join("misc", "regions.ini"))
	if err != nil {
		broadcast("Load Regions: error loading regions.ini: %s", err.Error())
		return
	}
	defer f.Close()
	var r util.ListFileReader
	r.StartReading(f)
	for seg := r.ReadNextSegment(); seg != nil; seg = r.ReadNextSegment() {
		region := &game.Region{}
		for _, s := range seg.Contents {
			parts := strings.SplitN(s, "=", 2)
			if len(parts) != 2 {
				broadcast("Load Regions: malformed regions.ini")
				return
			}
			name := parts[0]
			value := parts[1]
			switch name {
			case "Name":
				region.Name = value
			case "Music":
				region.Music = value
			case "Features":
				region.Features = game.RegionFeature(fn(value))
			case "SpawnMinZ":
				region.SpawnMinZ = int8(fn(value))
			case "SpawnMaxZ":
				region.SpawnMaxZ = int8(fn(value))
			case "Rect":
				parts = strings.Split(value, ",")
				if len(parts) != 4 {
					broadcast("Load Regions: malformed bounds in regions.ini")
					return
				}
				region.Rects = append(region.Rects, uo.Bounds{
					X: int16(fn(parts[0])),
					Y: int16(fn(parts[1])),
					Z: uo.MapMinZ,
					W: int16(fn(parts[2])),
					H: int16(fn(parts[3])),
					D: int16(uo.MapMaxZ) - int16(uo.MapMinZ),
				})
			case "Spawn":
				parts = strings.SplitN(value, ",", 3)
				if len(parts) != 3 {
					broadcast("Load Regions: malformed spawn definition in regions.ini")
					return
				}
				region.Entries = append(region.Entries, &game.SpawnerEntry{
					Amount:   fn(parts[0]),
					Delay:    uo.Time(fn(parts[1])),
					Template: parts[2],
				})
			}
		}
		region.ForceRecalculateBounds()
		game.GetWorld().Map().AddRegion(region)
	}
	broadcast("Load Regions: complete")
}

func commandRespawn(n game.NetState, args CommandArgs, cl string) {
	for _, r := range game.GetWorld().Map().RegionsWithin(uo.BoundsFullMap) {
		r.FullRespawn()
	}
}
