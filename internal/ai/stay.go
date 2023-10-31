package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Stay", func() AIModel { return &stay{} })
}

// stay implements a Thinker that stands in place and reacts to nothing.
type stay struct {
}

// Act implements the AIModel interface.
func (a *stay) Act(m game.Mobile, t uo.Time) {
	// Do nothing, ever
}

// Target implements the AIModel interface.
func (a *stay) Target(m game.Mobile, t uo.Time) {
	// No target selection
}
