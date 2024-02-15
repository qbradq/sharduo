package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/uo"
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
	broadcast("Save Regions: generating regions.json")
	fd, err := json.MarshalIndent(regions, "", "    ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(path.Join("data", "misc", "regions.json"), fd, 0666)
	if err != nil {
		broadcast(fmt.Sprintf("Save Regions: Error creating regions.json %s", err.Error()))
		return
	}
	broadcast("Save Regions: complete")
}

func commandLoadRegions(n game.NetState, args CommandArgs, cl string) {
	broadcast("Load Regions: clearing all regions")
	regions := game.World.Map().RegionsWithin(uo.BoundsFullMap)
	for _, r := range regions {
		game.World.Map().RemoveRegion(r)
	}
	broadcast("Load Regions: loading regions.json")
	fd, err := data.FS.ReadFile(path.Join("misc", "regions.json"))
	if err != nil {
		broadcast("Load Regions: error loading regions.json: %s", err.Error())
		return
	}
	regions = []*game.Region{}
	if err := json.Unmarshal(fd, &regions); err != nil {
		panic(err)
	}
	for _, region := range regions {
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
