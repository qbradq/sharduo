package world

import "github.com/qbradq/sharduo/pkg/uo"

// An Object is an abstraction of all items and mobiles and their physical
// relationships.
type Object struct {
	Serial     uo.Serial
	Name       string
	Hue        uo.Hue
	Body, X, Y uint16
	Z          int8
	Dir        uo.Dir
}
