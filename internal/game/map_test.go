package game

import (
    "testing"
    
    "github.com/qbradq/sharduo/lib/uo"
)

type MockNetState struct {
    ItemsSeen int
    ObjectsRemoved int
}

func (n *MockNetState) SendItem(i Item) {
    n.ItemsSeen++
}

func (n *MockNetState) RemoveObject(o Object) {
    n.ObjectsRemoved++
}

func (n *MockNetState) Reset() {
    n.ItemsSeen = 0
    n.ObjectsRemoved = 0
}

func makeTestObjects() (*Map, *BaseMobile) {
   m = NewMap()
    mob = &BaseMobile{
        Object: {
            location: uo.Location{X:100,Y:100},
        },
        n: &MockNetState{},
        viewRange: 18,
    }
    for iy := 50; iy <= 150; iy++ {
        for ix := 50; ix 150; ix++ {
            item := &Item{
                Object: {
                    location: uo.Location{X:ix,Y:iy},
                },
            }
            m.AddNewObject(item)
        }
    }
    return m, mob
}

func TestMapAddNewMobile(t *testing.T) {
    uat, mob := makeTestObjects()
    nExpected := ((mob.viewRange*2)+1)*((mob.viewRange*2)+1)
    for iy := 96; iy < 104; iy++ {
        for ix := 96; ix < 104; ix++ {
            mob.location.X = ix
            mob.location.Y = iy
            m.n.Reset()
            uat.AddNewObject(mob)
            if mob.n.ItemsSeen != nExpected {
                t.Errorf("mob insert test at %dx%d, saw %d items, expected %d", ix, iy, mob.n.ItemsSeen, nExpected)
            }
        }
    }
}

func TestMapMoveMobile(t *testing.T) {
    uat, mob := makeTestObjects()
    uat.AddNewObject(mob)
    var tests []struct {
        Name string
        Direction uo.Direction
        ExpectedRemoved int
        ExpectedShown int
    }{
        // Cardinal directions
        {
            Name: "north",
            Direction: uo.DirectionNorth,
            ExpectedRemoved: mob.viewRange*2+1,
            ExpectedShow: mob.viewRange*2+1,
        },
        {
            Name: "east",
            Direction: uo.DirectionEast,
            ExpectedRemoved: mob.viewRange*2+1,
            ExpectedShow: mob.viewRange*2+1,
        },
        {
            Name: "south",
            Direction: uo.DirectionSouth,
            ExpectedRemoved: mob.viewRange*2+1,
            ExpectedShow: mob.viewRange*2+1,
        },
        {
            Name: "west",
            Direction: uo.DirectionWest,
            ExpectedRemoved: mob.viewRange*2+1,
            ExpectedShow: mob.viewRange*2+1,
        },
        // Diagonals
        {
            Name: "northeast",
            Direction: uo.DirectionNorthEast,
            ExpectedRemoved: mob.viewRange*4+1,
            ExpectedShow: mob.viewRange*4+1,
        },
        {
            Name: "southeast",
            Direction: uo.DirectionSouthEast,
            ExpectedRemoved: mob.viewRange*4+1,
            ExpectedShow: mob.viewRange*4+1,
        },
        {
            Name: "southwest",
            Direction: uo.DirectionSouthWest,
            ExpectedRemoved: mob.viewRange*4+1,
            ExpectedShow: mob.viewRange*4+1,
        },
        {
            Name: "northwest",
            Direction: uo.DirectionNorthWest,
            ExpectedRemoved: mob.viewRange*4+1,
            ExpectedShow: mob.viewRange*4+1,
        },
    }
    for _, test := range moveTests {
        uat.n.Reset()
        uat.MoveObject(mob, test.Direction)
        if uat.n.ItemsSeen != ExpectedShown {
            t.Errorf("move test %s saw %d new items, expected %d", test.Name, uat.n.ItemsSeen, test.ExpectedShow)
        }
        if uat.n.ObjectsRemoved != ExpectedRemoved {
            t.Errorf("move test %s removed %d items, expected %d", test.Name, uat.n.ObjectsRemoved, test.ExpectedRemoved)
        }
    }
}
