package uod

import (
	"encoding/csv"
	"log"
	"strings"
)

// Command is the interface all command objects implement
type Command interface {
	// Compile takes all appropriate steps that can be done in advance of
	// command execution.
	Compile() error
	// Execute executes the command and should only be called after a call
	// to Compile. Execute may be ran multiple times per object.
	Execute() error
}

// ParseCommand returns a Command object parsed from a command line
func ParseCommand(line string) Command {
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ' ' // Space
	c, err := r.Read()
	if err != nil {
		log.Println(err)
	}
	return ret
}