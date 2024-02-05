// Package game implements all of the inner glue code for the game. Actual
// game play elements reside in packages [internal/ai], [internal/events] and
// [internal/gumps].
package game

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	// Load all object lists
	errors := false
	for _, err := range data.Walk("lists", func(s string, b []byte) []error {
		// Ignore legacy files
		if filepath.Ext(s) != ".json" {
			return nil
		}
		// Load lists
		ps := map[string][]string{}
		if err := json.Unmarshal(b, &ps); err != nil {
			return []error{err}
		}
		// Merge lists
		for k, p := range ps {
			// Check for duplicates
			if _, duplicate := mobilePrototypes[k]; duplicate {
				return []error{fmt.Errorf("duplicate list %s", k)}
			}
			// Initialize non-zero default values
			TemplateLists[k] = p
		}
		return nil
	}) {
		errors = true
		log.Printf("error: during list load: %v", err)
	}
	if errors {
		panic("errors during list load")
	}
}

// TemplateLists is a mapping of all template lists by name.
var TemplateLists = map[string][]string{}

// ListMember returns a random member from the named list.
func ListMember(which string) string {
	if l, found := TemplateLists[which]; found {
		return l[util.Random(0, len(l)-1)]
	}
	return ""
}

// ExecuteEventHandler must execute the named event handler.
var ExecuteEventHandler func(string, any, any, any) bool

// EventIndex must return a unique index number for the given event name.
var EventIndex func(string) uint16

// World is the current world we are simulating.
var World WorldInterface

// MapLocation returns the map location of the item if it directly on the map,
// or the location of the top-most container if the top-most container is on
// the map, or the location of the mobile who is wearing the top-most container.
// This function is useful during range checks.
func MapLocation(i *Item) uo.Point {
	if i.Wearer != nil {
		return i.Wearer.Location
	}
	for {
		if i.Container == nil {
			return i.Location
		}
		i = i.Container
	}
}

// Owner returns the mobile that is either wearing this item or contained within
// its backpack or bank box. May return nil.
func Owner(i *Item) *Mobile {
	for {
		if i.Wearer != nil {
			return i.Wearer
		}
		if i.Container == nil {
			return nil
		}
		i = i.Container
	}
}

// UOError represents a game rules violation and contains information on how
// to alert the player.
type UOError struct {
	Cliloc    uo.Cliloc // Cliloc of the error message or zero if the error message is a string
	Arguments []string  // Arguments to Cliloc, only valid if Cliloc is non zero
	Message   string    // String message, only valid if Cliloc is zero
}

// Error implements the Error interface.
func (e *UOError) Error() string {
	if e.Cliloc != 0 {
		return fmt.Sprintf("cliloc error %d %v", e.Cliloc, e.Arguments)
	}
	return e.Message
}

// Packet returns the server packet to send for this error.
func (e *UOError) Packet() serverpacket.Packet {
	if e.Cliloc != 0 {
		return &serverpacket.ClilocMessage{
			Hue:       1153,
			Cliloc:    e.Cliloc,
			Arguments: e.Arguments,
		}
	}
	return &serverpacket.Speech{
		Hue:  1153,
		Type: uo.SpeechTypeNormal,
		Text: e.Message,
	}
}
