package uod

import (
	"io"
	"log"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

// World encapsulates all of the data for the world and the goroutine that
// manipulates it.
type World struct {
	// The world map
	m *game.Map
	// The object data store for the entire world
	om *game.ObjectManager
	// The random number generator for the world
	rng *util.RNG
	// Inbound requests
	requestQueue chan WorldRequest
}

// NewWorld creates a new, empty world
func NewWorld() *World {
	rng := util.NewRNG()
	return &World{
		m:            game.NewMap(),
		om:           game.NewObjectManager(rng),
		rng:          rng,
		requestQueue: make(chan WorldRequest, 1024),
	}
}

// Load loads all of the data stores that the world is responsible for
func (w *World) Load(r io.Reader) []error {
	var ret []error
	errs := w.om.Load(r)
	if errs != nil {
		ret = append(ret, errs...)
	}
	return ret
}

// Save saves all of the data stores that the world is responsible for
func (w *World) Save(o io.Writer) []error {
	var ret []error
	errs := w.om.Save(o)
	if errs != nil {
		ret = append(ret, errs...)
	}
	return ret
}

// SendRequest sends a WorldRequest to the world's goroutine. Returns true if
// the command was successfully queued. This never blocks.
func (w *World) SendRequest(cmd WorldRequest) bool {
	select {
	case w.requestQueue <- cmd:
		return true
	default:
		return false
	}
}

// Random returns the *util.RNG the world is using for sync operations
func (w *World) Random() *util.RNG {
	return w.rng
}

// NewItem adds a new item to the world. It is assigned a unique serial. The
// item is returned.
func (w *World) NewItem(item game.Item) game.Item {
	return w.om.NewItem(item)
}

// NewMobile adds a new mobile to the world. It is assigned a unique serial. The
// mobile is returned.
func (w *World) NewMobile(mob game.Mobile) game.Mobile {
	return w.om.NewMobile(mob)
}

// Process is the goroutine that services the command queue and is the only
// goroutine allowed to interact with the contents of the world.
func (w *World) Process() {
	for r := range w.requestQueue {
		if err := r.Execute(); err != nil {
			log.Println(err)
		}
	}
}
