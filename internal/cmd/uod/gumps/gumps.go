package gumps

import (
	"fmt"
	"hash/crc32"
	"log"

	"github.com/qbradq/sharduo/internal/game"
)

// gumpDefinition ties together a GUMP's type code and constructor
type gumpDefinition struct {
	typeCode uint32
	ctor     func() game.GUMP
}

// Registry of all GUMPs
var gumpDefs = map[string]gumpDefinition{}

// reg registers a GUMP constructor and generates its type code
func reg(name string, fn func() game.GUMP) {
	if _, duplicate := gumpDefs[name]; duplicate {
		panic(fmt.Sprintf("duplicate GUMP definition %s", name))
	}
	d := gumpDefinition{
		typeCode: crc32.ChecksumIEEE([]byte(name)),
		ctor:     fn,
	}
	for k, v := range gumpDefs {
		if v.typeCode == d.typeCode {
			panic(fmt.Sprintf("hash collision between GUMP names %s and %s", k, name))
		}
	}
	gumpDefs[name] = d
}

// New creates a new GUMP by name
func New(name string) game.GUMP {
	d, ok := gumpDefs[name]
	if !ok {
		log.Printf("error: GUMP %s not found", name)
	}
	g := d.ctor()
	if g == nil {
		return nil
	}
	g.SetTypeCode(d.typeCode)
	return g
}
