package game

import (
	"io"

	"github.com/qbradq/sharduo/lib/util"
)

// MaxStabledPets is the maximum number of pets that can be stabled at one time.
const MaxStabledPets int = 10

// PlayerData encapsulates all of the data that is player-specific. This reduces
// memory consumption for non-player mobiles.
type PlayerData struct {
	StabledPets []*Mobile // Currently stabled pets
}

// NewPlayerData returns a new PlayerData struct ready for use.
func NewPlayerData() *PlayerData {
	return &PlayerData{}
}

// Write writes all player data to the writer.
func (d *PlayerData) Write(w io.Writer) {
	util.PutUInt32(w, uint32(len(d.StabledPets))) // Stabled pets
	for _, p := range d.StabledPets {
		p.Write(w)
	}
}

// NewPlayerDataFromReader reads the player data from r.
func NewPlayerDataFromReader(r io.Reader) *PlayerData {
	ret := &PlayerData{}
	ret.StabledPets = make([]*Mobile, util.GetUInt32(r)) // Stabled pets
	for i := range ret.StabledPets {
		ret.StabledPets[i] = NewMobileFromReader(r)
	}
	return ret
}

// Claim claims a stabled pet and gives it to the mobile.
func (d *PlayerData) Claim(m, owner *Mobile) {
	idx := -1
	for i, p := range d.StabledPets {
		if p == m {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	copy(d.StabledPets[idx:], d.StabledPets[idx+1:])
	d.StabledPets[len(d.StabledPets)-1] = nil
	d.StabledPets = d.StabledPets[:len(d.StabledPets)-1]
	m.Location = owner.Location
	m.AI = "Follow"
	m.AIGoal = owner
	World.Map().AddMobile(m, true)
}
