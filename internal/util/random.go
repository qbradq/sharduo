package util

import "math/rand"

// RandomBool returns a random bool value
func RandomBool() bool {
	return rand.Int31()%2 == 0
}
