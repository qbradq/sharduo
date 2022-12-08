package uod

import (
	"io"
	"log"
	"sync"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// World encapsulates all of the data for the world and the goroutine that
// manipulates it.
type World struct {
	// The world map
	m *game.Map
	// The data store of the user accounts
	ads *util.DataStore[*game.Account]
	// Index of usernames ot account serials
	aidx map[string]uo.Serial
	// The object data store for the entire world
	ods *util.DataStore[game.Object]
	// The random number generator for the world
	rng *util.RNG
	// Inbound requests
	requestQueue chan WorldRequest
	// Save/Load Mutex
	lock sync.Mutex
}

// NewWorld creates a new, empty world
func NewWorld() *World {
	rng := util.NewRNG()
	return &World{
		m:            game.NewMap(),
		ads:          util.NewDataStore[*game.Account]("accounts", rng, game.ObjectFactory),
		aidx:         make(map[string]uo.Serial),
		ods:          util.NewDataStore[game.Object]("objects", rng, game.ObjectFactory),
		rng:          rng,
		requestQueue: make(chan WorldRequest, 1024),
	}
}

// Read reads all of the data stores that the world is responsible for
func (w *World) Read(r io.Reader) []error {
	var ret []error
	ret = append(ret, w.ads.Read(r)...)
	ret = append(ret, w.ods.Read(r)...)
	return ret
}

// Write writes all of the data stores that the world is responsible for
func (w *World) Write(o io.Writer) []error {
	var ret []error
	ret = append(ret, w.ods.Write(o)...)
	ret = append(ret, w.ads.Write(o)...)
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

// New adds a new object to the world. It is assigned a unique serial. The
// object is returned.
func (w *World) New(o game.Object) game.Object {
	w.ods.Add(o, o.GetSerialType())
	return o
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
