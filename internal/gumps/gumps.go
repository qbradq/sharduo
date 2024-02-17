package gumps

import (
	"fmt"
	"hash/crc32"
	"log"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

// Static GUMP IDs, these shouldn't be needed often.
const (
	GUMPIDDecorate uo.Serial = 1
	GUMPIDRegions  uo.Serial = 2
)

// TypeCodeByName returns the type code for the given name.
func TypeCodeByName(name string) uo.Serial {
	return uo.Serial(crc32.ChecksumIEEE([]byte(name)))
}

// gumpDefinition ties together a GUMP's type code and constructor
type gumpDefinition struct {
	typeCode uo.Serial
	ctor     func() GUMP
}

// Registry of all GUMPs
var gumpDefs = map[string]gumpDefinition{}

// Command execution function
var ExecuteCommand func(game.NetState, string)

// reg registers a GUMP constructor and generates its type code
func reg(name string, tc uo.Serial, fn func() GUMP) {
	if _, duplicate := gumpDefs[name]; duplicate {
		panic(fmt.Sprintf("duplicate GUMP definition %s", name))
	}
	d := gumpDefinition{
		typeCode: TypeCodeByName(name),
		ctor:     fn,
	}
	if tc != uo.SerialZero {
		d.typeCode = tc
	}
	for k, v := range gumpDefs {
		if v.typeCode == d.typeCode {
			panic(fmt.Sprintf("hash collision between GUMP names %s and %s", k, name))
		}
	}
	gumpDefs[name] = d
}

// New creates a new GUMP by name
func New(name string) GUMP {
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

// Edit opens the editing GUMP for the object if any
func Edit(m *game.Mobile, o any) {
	if m == nil || o == nil || m.NetState == nil {
		return
	}
	if item, ok := o.(game.Item); ok {
		if item.TemplateName == "BaseSign" {
			m.NetState.GetText(item.Name, "Please enter the text of the sign.", 30, func(s string) {
				item.Name = s
			})
		}
	}
}
