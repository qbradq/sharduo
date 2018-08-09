package world

import "github.com/qbradq/sharduo/internal/common"

// An Object is an abstraction of all items and mobiles and their physical
// relationships.
type Object struct {
	Name       string
	Hue        common.Hue
	Body, X, Y uint16
	Z          int8
}
