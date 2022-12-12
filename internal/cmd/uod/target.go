package uod

import (
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// TargetCallback is the function signature for target callbacks
type TargetCallback func(*clientpacket.TargetResponse, interface{})

// Target represents an outstanding targeting request with a player
type Target struct {
	util.BaseSerialer
	// NetState the targeting request is bound for
	NetState *NetState
	// Callback function
	Callback TargetCallback
	// Context for the callback
	Context interface{}
	// Time to live for the request
	TTL uo.Time
}

// TargetManager manages all targeting requests for the world
type TargetManager struct {
	targets *util.SerialPool[*Target]
}

// NewTargetManager returns a new TargetManager object ready for use
func NewTargetManager(rng uo.RandomSource) *TargetManager {
	return &TargetManager{
		targets: util.NewSerialPool[*Target]("targets", rng),
	}
}

// New adds a new Target object to the manager, sets the Serial to a unique
// value, and returns the object
func (m *TargetManager) New(t *Target) *Target {
	m.targets.Add(t, uo.SerialTypeUnbound)
	return t
}

// Execute attempts to execute the callback for the given target. It returns
// true if the target still existed and the callback was executed.
func (m *TargetManager) Execute(r *clientpacket.TargetResponse) bool {
	t := m.targets.Get(r.TargetSerial)
	if t != nil {
		t.Callback(r, t.Context)
		return true
	}
	return false
}
