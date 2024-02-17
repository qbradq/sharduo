package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("WalkRandom", func() game.AIModel { return &walkRandom{} })
}

// walkRandom implements a Thinker that stands in place and reacts to nothing.
type walkRandom struct {
	nextWalkDeadline uo.Time      // Next time we should start walking
	walkDirection    uo.Direction // Direction we are walking
	stepsLeft        int          // Steps left in the current walk
	walking          bool         // If true we are currently walking, not waiting
}

// Act implements the AIModel interface.
func (a *walkRandom) Act(m *game.Mobile, t uo.Time) {
	if t >= a.nextWalkDeadline {
		a.walking = true
	}
	if a.walking && a.stepsLeft < 1 {
		a.walking = false
		a.stepsLeft = util.Random(1, 7)
		a.nextWalkDeadline = t + uo.Time(util.Random(
			int(uo.DurationSecond)*5,
			int(uo.DurationSecond)*15))
		a.walkDirection = uo.Direction(util.Random(0, 7))
	}
	if a.walking && m.CanTakeStep() {
		if m.Spawner != nil && !m.Spawner.Contains(m.Location.Forward(a.walkDirection)) {
			a.walkDirection = uo.Direction(util.Random(0, 7))
		}
		if !m.Step(a.walkDirection) {
			a.walking = false
			a.stepsLeft = 0
			a.nextWalkDeadline = t + uo.Time(util.Random(
				int(uo.DurationSecond)*1,
				int(uo.DurationSecond)*5))
			a.walkDirection = uo.Direction(util.Random(0, 7))
		} else {
			a.stepsLeft--
		}
	}
}

// Target implements the AIModel interface.
func (a *walkRandom) Target(m *game.Mobile, t uo.Time) {
	// No target selection
}
