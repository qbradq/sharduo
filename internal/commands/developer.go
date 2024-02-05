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
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Developer commands go here, generally these should not be used in production

func init() {
	reg(&cmDesc{"decorate", []string{"deco"}, commandDecorate, game.RoleDeveloper, "decorate", "Calls up the decoration GUMP"})
	reg(&cmDesc{"load_doors", nil, commandLoadDoors, game.RoleDeveloper, "load_doors", "Clears all doors then loads data/misc/doors.csv"})
	reg(&cmDesc{"load_regions", nil, commandLoadRegions, game.RoleDeveloper, "load_regions", "Clears all regions then loads data/misc/regions.csv"})
	reg(&cmDesc{"load_signs", nil, commandLoadSigns, game.RoleDeveloper, "load_signs", "Clears all signs then loads data/misc/signs.csv"})
	reg(&cmDesc{"load_statics", nil, commandLoadStatics, game.RoleDeveloper, "load_statics", "Clears all statics then loads data/misc/statics.csv"})
	reg(&cmDesc{"regions", nil, commandRegions, game.RoleDeveloper, "regions", "Calls up the regions GUMP"})
	reg(&cmDesc{"respawn", nil, commandRespawn, game.RoleDeveloper, "respawn", "Executes a full respawn on all spawning regions"})
	reg(&cmDesc{"save_doors", nil, commandSaveDoors, game.RoleDeveloper, "save_doors", "Generates data/misc/doors.csv"})
	reg(&cmDesc{"save_regions", nil, commandSaveRegions, game.RoleDeveloper, "save_regions", "Generates data/misc/regions.csv"})
	reg(&cmDesc{"save_signs", nil, commandSaveSigns, game.RoleDeveloper, "save_signs", "Generates data/misc/signs.csv"})
	reg(&cmDesc{"save_statics", nil, commandSaveStatics, game.RoleDeveloper, "save_statics", "Generates data/misc/statics.csv"})
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
	for _, s := range game.World.Map().ItemQuery("StaticItem", uo.Point{}, 0) {
		game.World.RemoveItem(s)
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
		s := game.NewItem("StaticItem")
		s.Location = uo.Point{
			X: fn(fields[0]),
			Y: fn(fields[1]),
			Z: fn(fields[2]),
		}
		s.Graphic = uo.Graphic(fn(fields[3]))
		s.Hue = uo.Hue(fn(fields[4]))
		game.World.Map().AddItem(s, true)
	}
	broadcast("Load Statics: complete")
}

func commandSaveStatics(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Statics: getting statics")
	statics := game.World.Map().ItemQuery("StaticItem", uo.Point{}, 0)
	broadcast("Save Statics: sorting statics")
	sort.Slice(statics, func(i, j int) bool {
		a := statics[i].Location
		b := statics[j].Location
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
		l := s.Location
		f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n",
			l.X, l.Y, l.Z,
			s.CurrentGraphic(),
			s.Hue,
		))
	}
	broadcast("Save Statics complete")
}

func commandDecorate(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("decorate"), 0, 0)
}

func commandSaveDoors(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Doors: getting doors")
	doors := game.World.Map().ItemBaseQuery("BaseDoor", uo.Point{}, 0)
	broadcast("Save Doors: sorting doors")
	// Correct door locations for doors that happen to be open
	for _, d := range doors {
		if d.Flipped {
			l := d.Location
			ofs := uo.DoorOffsets[d.Facing]
			l.X -= ofs.X
			l.Y -= ofs.Y
			game.World.Map().RemoveItem(d)
			d.Location = l
			d.Flipped = false
			d.Def = game.World.ItemDefinition(d.CurrentGraphic())
			game.World.Map().AddItem(d, true)
		}
	}
	// Sort based on corrected locations
	sort.Slice(doors, func(i, j int) bool {
		a := doors[i].Location
		b := doors[j].Location
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
		l := d.Location
		f.WriteString(fmt.Sprintf("%d,%d,%d,%s,%d\n",
			l.X, l.Y, l.Z,
			d.TemplateName,
			d.Facing,
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
	for _, s := range game.World.Map().ItemBaseQuery("BaseDoor", uo.Point{}, 0) {
		game.World.RemoveItem(s)
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
		d := game.NewItem(fields[3])
		d.Location = uo.Point{
			X: fn(fields[0]),
			Y: fn(fields[1]),
			Z: fn(fields[2]),
		}
		d.Facing = uo.Direction(fn(fields[4]))
		d.Graphic = d.BaseGraphic() + uo.Graphic(d.Facing*2)
		d.FlippedGraphic = d.FlippedGraphic + uo.Graphic(d.Facing*2)
		game.World.Map().AddItem(d, true)
	}
	broadcast("Load Doors: complete")
}

func commandSaveSigns(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Signs: getting signs")
	signs := game.World.Map().ItemQuery("BaseSign", uo.Point{}, 0)
	broadcast("Save Signs: sorting signs")
	// Sort based on location
	sort.Slice(signs, func(i, j int) bool {
		a := signs[i].Location
		b := signs[j].Location
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
		l := s.Location
		f.WriteString(fmt.Sprintf("%d,%d,%d,%d,\"%s\"\n",
			l.X, l.Y, l.Z,
			s.Graphic,
			s.Name,
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
	for _, s := range game.World.Map().ItemBaseQuery("BaseSign", uo.Point{}, 0) {
		game.World.RemoveItem(s)
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
		s := game.NewItem("BaseSign")
		s.Location = uo.Point{
			X: fn(fields[0]),
			Y: fn(fields[1]),
			Z: fn(fields[2]),
		}
		s.Graphic = uo.Graphic(fn(fields[3]))
		s.Name = fields[4]
		game.World.Map().AddItem(s, true)
	}
	broadcast("Load Doors: complete")
}

func commandRegions(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("regions"), n.Mobile().Serial, 0)
}

func commandSaveRegions(n game.NetState, args CommandArgs, cl string) {
	broadcast("Save Regions: getting and sorting regions")
	regions := game.World.Map().RegionsWithin(uo.BoundsFullMap)
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
	regions := game.World.Map().RegionsWithin(uo.BoundsFullMap)
	for _, r := range regions {
		game.World.Map().RemoveRegion(r)
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
				region.SpawnMinZ = fn(value)
			case "SpawnMaxZ":
				region.SpawnMaxZ = fn(value)
			case "Rect":
				parts = strings.Split(value, ",")
				if len(parts) != 4 {
					broadcast("Load Regions: malformed bounds in regions.ini")
					return
				}
				region.Rects = append(region.Rects, uo.Bounds{
					X: fn(parts[0]),
					Y: fn(parts[1]),
					Z: uo.MapMinZ,
					W: fn(parts[2]),
					H: fn(parts[3]),
					D: uo.MapMaxZ - uo.MapMinZ,
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
		game.World.Map().AddRegion(region)
	}
	broadcast("Load Regions: complete")
}

func commandRespawn(n game.NetState, args CommandArgs, cl string) {
	for _, r := range game.World.Map().RegionsWithin(uo.BoundsFullMap) {
		r.FullRespawn()
	}
}
