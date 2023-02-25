package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	timerPools = make([]map[uo.Serial]*Timer, 1+mediumSpeedTimerPoolsCount+lowSpeedTimerPoolsCount)
	for i := 0; i < len(timerPools); i++ {
		timerPools[i] = make(map[uo.Serial]*Timer)
	}
}

// There is always only 1 high-speed timer pool

// Number of pools in the low-speed timer pool
const lowSpeedTimerPoolsCount = int(uo.DurationMinute)

// Number of pools in the medium speed timer pool
const mediumSpeedTimerPoolsCount = int(uo.DurationSecond)

// Slice of all timer pools
var timerPools []map[uo.Serial]*Timer

// Timer dispatches an event after a set interval, optionally repeating. If
// either the receiver or source objects have been deleted prior to the trigger
// the event will not fire. Nil may be passed for either or both the receiver
// and source objects.
type Timer struct {
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
func NewTimer(delay uo.Time, event string, receiver, source Object) {
	receiverSerial := uo.SerialZero
	sourceSerial := uo.SerialZero
	if receiver != nil {
		receiverSerial = receiver.Serial()
	}
	if source != nil {
		sourceSerial = source.Serial()
	}
	t := &Timer{
		deadline: world.Time() + delay,
		event:    event,
		receiver: receiverSerial,
		source:   sourceSerial,
	}
	for {
		serial := uo.RandomMobileSerial(world.Random())
		pool := 0
		if delay > uo.DurationSecond && delay < uo.DurationMinute {
			pool = 1 + (int(serial) % mediumSpeedTimerPoolsCount)
		} else {
			pool = 1 + mediumSpeedTimerPoolsCount + (int(serial) % lowSpeedTimerPoolsCount)
		}
		if timerPools[pool][serial] == nil {
			timerPools[pool][serial] = t
			break
		}
		// Duplicate serial in pool, try another
	}
}

// UpdateTimers updates every timer within the update pools suitable for time.
func UpdateTimers(now uo.Time) {
	fn := func(timers map[uo.Serial]*Timer) {
		var toRemove []uo.Serial
		for s, t := range timers {
			if now >= t.deadline {
				t.Execute()
				toRemove = append(toRemove, s)
			}
		}
		for _, s := range toRemove {
			delete(timers, s)
		}
	}
	fn(timerPools[0])
	fn(timerPools[1+(int(now)%mediumSpeedTimerPoolsCount)])
	fn(timerPools[1+mediumSpeedTimerPoolsCount+(int(now)%lowSpeedTimerPoolsCount)])
}

// MarshalTimers writes all timers to the segment.
func MarshalTimers(s *marshal.TagFileSegment) {
	for pool, timers := range timerPools {
		for serial, t := range timers {
			s.PutInt(uint32(serial))
			s.PutShort(uint16(pool))
			s.PutLong(uint64(t.deadline))
			s.PutInt(uint32(t.receiver))
			s.PutInt(uint32(t.source))
			s.PutString(t.event)
			s.IncrementRecordCount()
		}
	}
}

// UnmarshalTimers reads all timers from the segment.
func UnmarshalTimers(s *marshal.TagFileSegment) {
	for i := uint32(0); i < s.RecordCount(); i++ {
		serial := uo.Serial(s.Int())
		pool := int(s.Short())
		deadline := uo.Time(s.Long())
		receiver := uo.Serial(s.Int())
		source := uo.Serial(s.Int())
		event := s.String()
		s := uo.Serial(serial)
		if pool < 0 || pool >= len(timerPools) {
			log.Printf("timer %s pool %d out of range", s.String(), pool)
			continue
		}
		if _, duplicate := timerPools[pool][s]; duplicate {
			log.Printf("timer %s is a duplicate in pool %d", s.String(), pool)
			continue
		}
		timerPools[pool][serial] = &Timer{
			deadline: deadline,
			event:    event,
			receiver: receiver,
			source:   source,
		}
	}
}

// Execute executes the event on the timer
func (t *Timer) Execute() {
	var receiver Object
	var source Object
	if t.receiver != uo.SerialZero {
		receiver = world.Find(t.receiver)
		if receiver == nil {
			return
		}
	}
	if t.source != uo.SerialZero {
		source = world.Find(t.source)
		if source == nil {
			return
		}
	}
	ExecuteEventHandler(t.event, receiver, source, nil)
}
