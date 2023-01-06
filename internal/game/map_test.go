package game

import (
	"testing"

	"github.com/qbradq/sharduo/lib/uo"
)

type MockNetState struct {
	ObjectsSeen    int
	ObjectsRemoved int
}

func (n *MockNetState) SendObject(o Object) {
	n.ObjectsSeen++
}

func (n *MockNetState) RemoveObject(o Object) {
	n.ObjectsRemoved++
}

func (n *MockNetState) Reset() {
	n.ObjectsSeen = 0
	n.ObjectsRemoved = 0
}

// Required for interface compliance
func (n *MockNetState) SystemMessage(fmtstr string, args ...interface{})       {}
func (n *MockNetState) Speech(from Object, fmtstr string, args ...interface{}) {}
func (n *MockNetState) DrawPlayer()                                            {}
func (n *MockNetState) MoveMobile(mob Mobile)                                  {}
func (n *MockNetState) UpdateObject(Object)                                    {}
func (n *MockNetState) WornItem(wearable Wearable, mob Mobile)                 {}
func (n *MockNetState) DropReject(reason uo.MoveItemRejectReason)              {}
func (n *MockNetState) ContainerOpen(c Container)                              {}
func (n *MockNetState) ContainerClose(c Container)                             {}
func (n *MockNetState) ContainerItemAdded(c Container, item Item)              {}
func (n *MockNetState) ContainerItemRemoved(c Container, item Item)            {}
func (n *MockNetState) ContainerRangeCheck()                                   {}
func (n *MockNetState) ContainerIsObserving(o Object) bool                     { return false }
func (n *MockNetState) OpenPaperDoll(m Mobile)                                 {}
func (n *MockNetState) CloseGump(gump uo.Serial)                               {}
func (n *MockNetState) DragItem(item Item, srcMob Mobile, srcLoc uo.Location, destMob Mobile, destLoc uo.Location) {
}

func makeTestObjects() (*Map, *BaseMobile) {
	m := NewMap()
	mob := &BaseMobile{
		BaseObject: BaseObject{
			location: uo.Location{X: 100, Y: 100},
		},
		n:         &MockNetState{},
		viewRange: 18,
	}
	for iy := 50; iy <= 150; iy++ {
		for ix := 50; ix <= 150; ix++ {
			item := &BaseItem{
				BaseObject: BaseObject{
					location: uo.Location{X: ix, Y: iy},
				},
			}
			m.AddObject(item)
		}
	}
	return m, mob
}

func TestMapGetChunksInBounds(t *testing.T) {
	uat, _ := makeTestObjects()
	tests := []struct {
		X               int
		Y               int
		W               int
		H               int
		ExpectedNChunks int
	}{
		{
			X:               100,
			Y:               100,
			W:               uo.ChunkWidth,
			H:               uo.ChunkHeight,
			ExpectedNChunks: 4,
		},
		{
			X:               96,
			Y:               96,
			W:               uo.ChunkWidth,
			H:               uo.ChunkHeight,
			ExpectedNChunks: 1,
		},
		{
			X:               100,
			Y:               100,
			W:               uo.ChunkWidth*4 - 1,
			H:               uo.ChunkHeight*3 - 2,
			ExpectedNChunks: 20,
		},
	}
	for _, test := range tests {
		chunks := uat.getChunksInBounds(uo.Bounds{
			X: test.X,
			Y: test.Y,
			W: test.W,
			H: test.H,
		})
		if len(chunks) != test.ExpectedNChunks {
			t.Errorf("getChunksInBounds returned %d chunks, expected %d at X=%d Y=%d W=%d H=%d",
				len(chunks), test.ExpectedNChunks, test.X, test.Y, test.W, test.H)
		}
	}
}

func TestMapGetChunksInRange(t *testing.T) {
	uat, _ := makeTestObjects()
	tests := []struct {
		X               int
		Y               int
		R               int
		ExpectedNChunks int
	}{
		{
			X:               100,
			Y:               100,
			R:               2,
			ExpectedNChunks: 1,
		},
		{
			X:               100,
			Y:               100,
			R:               uo.ChunkWidth,
			ExpectedNChunks: 9,
		},
	}
	for _, test := range tests {
		chunks := uat.getChunksInRange(uo.Location{
			X: test.X,
			Y: test.Y,
		}, test.R)
		if len(chunks) != test.ExpectedNChunks {
			t.Errorf("getChunksInRange got %d chunks back, expected %d at X=%d Y=%d R=%d",
				len(chunks), test.ExpectedNChunks, test.X, test.Y, test.R)
		}
	}
}

func TestMapAddNewMobile(t *testing.T) {
	uat, mob := makeTestObjects()
	nExpected := ((mob.viewRange * 2) + 1) * ((mob.viewRange * 2) + 1)
	for iy := 96; iy < 104; iy++ {
		for ix := 96; ix < 104; ix++ {
			mob.location.X = ix
			mob.location.Y = iy
			mob.n.(*MockNetState).Reset()
			uat.AddObject(mob)
			if mob.n.(*MockNetState).ObjectsSeen != nExpected {
				t.Errorf("mob insert test at %dx%d, saw %d items, expected %d", ix, iy, mob.n.(*MockNetState).ObjectsSeen, nExpected)
			}
		}
	}
}

func TestMapMoveMobile(t *testing.T) {
	uat, mob := makeTestObjects()
	uat.AddObject(mob)
	tests := []struct {
		Name            string
		Direction       uo.Direction
		ExpectedRemoved int
		ExpectedShown   int
	}{
		// Cardinal directions
		{
			Name:            "north",
			Direction:       uo.DirectionNorth,
			ExpectedRemoved: mob.viewRange*2 + 1,
			ExpectedShown:   mob.viewRange*2 + 1,
		},
		{
			Name:            "east",
			Direction:       uo.DirectionEast,
			ExpectedRemoved: mob.viewRange*2 + 1,
			ExpectedShown:   mob.viewRange*2 + 1,
		},
		{
			Name:            "south",
			Direction:       uo.DirectionSouth,
			ExpectedRemoved: mob.viewRange*2 + 1,
			ExpectedShown:   mob.viewRange*2 + 1,
		},
		{
			Name:            "west",
			Direction:       uo.DirectionWest,
			ExpectedRemoved: mob.viewRange*2 + 1,
			ExpectedShown:   mob.viewRange*2 + 1,
		},
		// Diagonals
		{
			Name:            "northeast",
			Direction:       uo.DirectionNorthEast,
			ExpectedRemoved: mob.viewRange*4 + 1,
			ExpectedShown:   mob.viewRange*4 + 1,
		},
		{
			Name:            "southeast",
			Direction:       uo.DirectionSouthEast,
			ExpectedRemoved: mob.viewRange*4 + 1,
			ExpectedShown:   mob.viewRange*4 + 1,
		},
		{
			Name:            "southwest",
			Direction:       uo.DirectionSouthWest,
			ExpectedRemoved: mob.viewRange*4 + 1,
			ExpectedShown:   mob.viewRange*4 + 1,
		},
		{
			Name:            "northwest",
			Direction:       uo.DirectionNorthWest,
			ExpectedRemoved: mob.viewRange*4 + 1,
			ExpectedShown:   mob.viewRange*4 + 1,
		},
	}
	// Single-step tests
	for _, test := range tests {
		mob.n.(*MockNetState).Reset()
		mob.facing = test.Direction
		uat.MoveMobile(mob, test.Direction)
		if mob.n.(*MockNetState).ObjectsSeen != test.ExpectedShown {
			t.Errorf("single-step move test %s saw %d new items, expected %d", test.Name, mob.n.(*MockNetState).ObjectsSeen, test.ExpectedShown)
		}
		if mob.n.(*MockNetState).ObjectsRemoved != test.ExpectedRemoved {
			t.Errorf("single-step move test %s removed %d items, expected %d", test.Name, mob.n.(*MockNetState).ObjectsRemoved, test.ExpectedRemoved)
		}
	}
	// Multi-step tests
	for _, test := range tests {
		mob.n.(*MockNetState).Reset()
		mob.facing = test.Direction
		for i := 0; i < 30; i++ {
			mob.n.(*MockNetState).Reset()
			mob.location.X = 100
			mob.location.Y = 100
			uat.MoveMobile(mob, test.Direction)
			if mob.n.(*MockNetState).ObjectsSeen != test.ExpectedShown {
				t.Errorf("multi-step move test %s saw %d new items, expected %d on step %d", test.Name, mob.n.(*MockNetState).ObjectsSeen, test.ExpectedShown, i)
			}
			if mob.n.(*MockNetState).ObjectsRemoved != test.ExpectedRemoved {
				t.Errorf("multi-step move test %s removed %d items, expected %d on step %d", test.Name, mob.n.(*MockNetState).ObjectsRemoved, test.ExpectedRemoved, i)
			}
		}
	}
}
