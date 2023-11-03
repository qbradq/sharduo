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
func (a *follow) Act(m game.Mobile, t uo.Time) {
	o := m.AIGoal()
	if o == nil {
		return
	}
	ft, ok := o.(game.Mobile)
	if !ok {
		return
	}
	if m.Location().XYDistance(ft.Location()) < 3 {
		// We don't need to be all up in our target's business
		return
	}
	// Step toward the target
	if !m.CanTakeStep() {
		return
	}
	d := m.Location().DirectionTo(ft.Location())
	if m.Step(d) {
		return
	}
	// Something is blocking us, try to step around
	d = (d - 1).Bound()
	if m.Step(d) {
		return
	}
	d = (d + 2).Bound()
	m.Step(d)
	// If we couldn't get around the blocking object we're stuck
	// TODO Path finding?
}

// Target implements the AIModel interface.
func (a *follow) Target(m game.Mobile, t uo.Time) {
	// No target selection
}
