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
