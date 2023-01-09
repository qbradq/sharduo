package game

import (
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	for i := 0; i < mediumSpeedTimerPoolsCount; i++ {
		mediumSpeedTimerPools[i] = make(map[int]*Timer)
	}
	for i := 0; i < lowSpeedTimerPoolsCount; i++ {
		lowSpeedTimerPools[i] = make(map[int]*Timer)
	}
}

// Serial manager for timers
var timerSerialManager *uo.SerialManager

// Number of pools in the low-speed timer pool
const lowSpeedTimerPoolsCount = int(uo.DurationMinute)

// Number of pools in the medium speed timer pool
const mediumSpeedTimerPoolsCount = int(uo.DurationSecond)

// Timers in this pool are checked every tick
var highSpeedTimerPool map[uo.Serial]*Timer

// Timers in this pool are checked every real-world second
var mediumSpeedTimerPools map[int]map[uo.Serial]*Timer

// Timers in this pool are checked every real-world minute
var lowSpeedTimerPools map[int]map[uo.Serial]*Timer

// Timer dispatches an event after a set interval, optionally repeating. If
// either the receiver or source objects have been deleted prior to the trigger
// the event will not fire. Nil may be passed for either or both the receiver
// and source objects.
type Timer struct {
	// ID of the timer
	serial uo.Serial
	// When the timer should trigger next
	deadline uo.Time
	// Name of the event
	event string
	// Serial of the receiver of the event
	receiver uo.Serial
	// Serial of the source of the event
	source uo.Serial
}

// NewTimer creates a new timer with the given options, then adds the timer to
// the update pool most suitable for it.
func NewTimer(delay uo.Time, oneShot bool, event string, receiver, source Object) *Timer {
	if timerSerialManager == nil {
		timerSerialManager = uo.NewSerialManager(world.Random())
	}
	t := &Timer{
		serial:   timerSerialManager.New(uo.SerialTypeUnbound),
		deadline: world.Time() + delay,
		event:    event,
		receiver: receiver.Serial(),
		source:   source.Serial(),
	}
	if delay < uo.DurationSecond {
		highSpeedTimerPool[t.serial] = t
	} else if delay < uo.DurationMinute {
		mediumSpeedTimerPools[int(t.serial)%mediumSpeedTimerPoolsCount][t.serial] = t
	} else {
		lowSpeedTimerPools[int(t.serial)&lowSpeedTimerPoolsCount][t.serial] = t
	}
	return t
}

// UpdateTimers updates every timer within the update pools suitable for time.
func UpdateTimers(now uo.Time) {
	var toRemove []uo.Serial
	for s, t := range highSpeedTimerPool {
		if now >= t.deadline {
			t.Execute()
			toRemove = append(toRemove, s)
		}
	}
	for _, s := range toRemove {
		delete(highSpeedTimerPool, s)
	}
}

// Execute executes the event on the timer
func (t *Timer) Execute() {

}
