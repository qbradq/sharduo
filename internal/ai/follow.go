package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Follow", func() AIModel { return &follow{} })
}

// follow implements a Thinker that follows its goal.
type follow struct {
}

// Act implements the AIModel interface.
func (a *follow) Act(m *game.Mobile, t uo.Time) {
	if m.Location.XYDistance(m.AIGoal.Location) < 3 {
		// We don't need to be all up in our target's business
		return
	}
	// Step toward the target
	if !m.CanTakeStep() {
		return
	}
	d := m.Location.DirectionTo(m.AIGoal.Location)
	if m.Step(d) {
		return
	}
	// Something is blocking us, try to step around
	step := 1
	if !d.IsDiagonal() {
		step++
	}
	if m.Step(d.RotateCounterclockwise(step)) {
		return
	}
	m.Step(d.RotateClockwise(step))
	// If we couldn't get around the blocking object we're stuck
	// TODO Path finding?
}

// Target implements the AIModel interface.
func (a *follow) Target(m *game.Mobile, t uo.Time) {
	// No target selection
}
