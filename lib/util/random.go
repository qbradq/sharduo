package util

import (
	"math/rand"
)

// RandomBool returns a random bool value
func RandomBool() bool {
	return rand.Int()%2 == 0
}

// Random returns a random int value between min and max inclusive
func Random(min, max int) int {
	return rand.Intn((max-min)+1) + min
}
