package uo

// RandomSource is the interface sources of psuedo-random numbers that the
// Random* functions expect.
type RandomSource interface {
	// RandomBool returns a random boolean value
	RandomBool() bool
	// Random returns a random int value between min and max inclusive
	Random(int, int) int
}
