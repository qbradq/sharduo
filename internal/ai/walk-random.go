package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("WalkRandom", func() AIModel { return &walkRandom{} })
}

// walkRandom implements a Thinker that stands in place and reacts to nothing.
type walkRandom struct {
	nextWalkDeadline uo.Time      // Next time we should start walking
	walkDirection    uo.Direction // Direction we are walking
	stepsLeft        int          // Steps left in the current walk
	walking          bool         // If true we are currently walking, not waiting
}

// Act implements the AIModel interface.
func (a *walkRandom) Act(m game.Mobile, t uo.Time) {
	if t >= a.nextWalkDeadline {
		a.walking = true
	}
	if a.walking && a.stepsLeft < 1 {
		a.walking = false
		a.stepsLeft = game.GetWorld().Random().Random(1, 7)
		a.nextWalkDeadline = t + uo.Time(game.GetWorld().Random().Random(
			int(uo.DurationSecond)*5,
			int(uo.DurationSecond)*15))
		a.walkDirection = uo.Direction(game.GetWorld().Random().Random(0, 7))
	}
	if a.walking && m.CanTakeStep() {
		outOfBounds := false
		o := m.Owner()
		if o != nil {
			if s, ok := o.(*game.Spawner); ok {
				outOfBounds = m.Location().Forward(a.walkDirection).XYDistance(s.Location()) > int16(s.Radius)
			}
		}
		if outOfBounds {
			a.walkDirection = uo.Direction(game.GetWorld().Random().Random(0, 7))
		} else if !m.Step(a.walkDirection) {
			a.walking = false
			a.stepsLeft = 0
			a.nextWalkDeadline = t + uo.Time(game.GetWorld().Random().Random(
				int(uo.DurationSecond)*1,
				int(uo.DurationSecond)*5))
			a.walkDirection = uo.Direction(game.GetWorld().Random().Random(0, 7))
		} else {
			a.stepsLeft--
		}
	}
}

// Target implements the AIModel interface.
func (a *walkRandom) Target(m game.Mobile, t uo.Time) {
	// No target selection
}
