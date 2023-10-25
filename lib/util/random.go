package util

import (
	"math/rand"
	"time"
)

// RandomSource is the interface sources of psuedo-random numbers that the
// Random* functions expect.
type RandomSource interface {
	// Random returns a random int value between min and max inclusive
	Random(int, int) int
}

// RNG is a statefull, NON-thread-safe random number generator
type RNG struct {
	r *rand.Rand
}

// NewRNG returns a new RNG object seeded with the current millisecond time.
func NewRNG() *RNG {
	return &RNG{
		r: rand.New(rand.NewSource(time.Now().UnixMilli())),
	}
}

// RandomBool returns a random bool value
func (r *RNG) RandomBool() bool {
	return r.r.Int31()%2 == 0
}

// Random returns a random int value between min and max inclusive
func (r *RNG) Random(min, max int) int {
	return r.r.Intn((max-min)+1) + min
}
