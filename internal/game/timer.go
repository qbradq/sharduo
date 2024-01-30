package game

import (
	"io"
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
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

// Collection of serials currently in use
var timerSerials = map[uo.Serial]*Timer{}

// Timer dispatches an event after a set interval, optionally repeating. If
// either the receiver or source objects have been deleted prior to the trigger
// the event will not fire. Nil may be passed for either or both the receiver
// and source objects.
type Timer struct {
	// When the timer should trigger next
	deadline uo.Time
	// What pool the timer lives in
	pool int
	// Name of the event
	event string
	// Serial of the receiver of the event
	receiver uo.Serial
	// Serial of the source of the event
	source uo.Serial
	// True if this is an non-saved timer
	noRent bool
	// Parameter of the event, only used if this is an ephemeral timer
	parameter any
}

// NewTimer creates a new timer with the given options, then adds the timer to
// the update pool most suitable for it. A serial number uniquely identifying
// the timer is returned. See CancelTimer().
func NewTimer(delay uo.Time, event string, receiver, source any, noRent bool, parameter any) uo.Serial {
	rs := uo.SerialZero
	ss := uo.SerialZero
	switch o := receiver.(type) {
	case *Item:
		rs = o.Serial
	case *Mobile:
		rs = o.Serial
	}
	switch o := source.(type) {
	case *Item:
		ss = o.Serial
	case *Mobile:
		ss = o.Serial
	}
	t := &Timer{
		deadline:  Time() + delay,
		event:     event,
		receiver:  rs,
		source:    ss,
		noRent:    noRent,
		parameter: parameter,
	}
	for {
		serial := uo.Serial(util.Random(int(uo.SerialFirstMobile), int(uo.SerialLastMobile)))
		if timerSerials[serial] != nil {
			// Duplicate serial
			continue
		}
		if delay > uo.DurationSecond && delay < uo.DurationMinute {
			t.pool = 1 + (int(serial) % mediumSpeedTimerPoolsCount)
		} else if delay >= uo.DurationMinute {
			t.pool = 1 + mediumSpeedTimerPoolsCount + (int(serial) % lowSpeedTimerPoolsCount)
		} // Else delay <= uo.DurationSecond, pool stays 0
		timerPools[t.pool][serial] = t
		timerSerials[serial] = t
		return serial
	}
}

// CancelTimer cancels the timer identified by serial if it exists.
func CancelTimer(s uo.Serial) {
	t, found := timerSerials[s]
	if !found {
		return
	}
	delete(timerSerials, s)
	delete(timerPools[t.pool], s)
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
			delete(timerSerials, s)
		}
	}
	fn(timerPools[0])
	fn(timerPools[1+(int(now)%mediumSpeedTimerPoolsCount)])
	fn(timerPools[1+mediumSpeedTimerPoolsCount+(int(now)%lowSpeedTimerPoolsCount)])
}

// WriteTimers writes all timer data to the writer.
func WriteTimers(w io.Writer) {
	for pool, timers := range timerPools {
		for serial, t := range timers {
			if t.noRent {
				// Ephemeral timers do not get saved
				continue
			}
			util.PutBool(w, true)
			util.PutUInt32(w, uint32(serial))
			util.PutUInt16(w, uint16(pool))
			util.PutUInt64(w, uint64(t.deadline))
			util.PutUInt32(w, uint32(t.receiver))
			util.PutUInt32(w, uint32(t.source))
			util.PutString(w, t.event)
		}
	}
	util.PutBool(w, false)
}

// ReadTimers reads all timers from the segment.
func ReadTimers(r io.Reader) {
	for util.GetBool(r) {
		serial := uo.Serial(util.GetUInt32(r))
		pool := int(util.GetUInt16(r))
		deadline := uo.Time(util.GetUInt64(r))
		receiver := uo.Serial(util.GetUInt32(r))
		source := uo.Serial(util.GetUInt32(r))
		event := util.GetString(r)
		if pool < 0 || pool >= len(timerPools) {
			log.Printf("error: timer %s pool %d out of range", serial.String(), pool)
			continue
		}
		if _, duplicate := timerSerials[serial]; duplicate {
			log.Printf("error: timer %s is a duplicate in pool %d", serial.String(), pool)
			continue
		}
		t := &Timer{
			deadline: deadline,
			event:    event,
			receiver: receiver,
			source:   source,
		}
		timerPools[pool][serial] = t
		timerSerials[serial] = t
	}
}

// Execute executes the event on the timer
func (t *Timer) Execute() {
	var receiver any
	var source any
	if t.receiver != uo.SerialZero {
		receiver = Find(t.receiver)
		if receiver == nil {
			return
		}
		switch o := receiver.(type) {
		case *Item:
			if o.Removed {
				return
			}
		case *Mobile:
			if o.Removed {
				return
			}
		}
	}
	if t.source != uo.SerialZero {
		source = Find(t.source)
		if source == nil {
			return
		}
		switch o := source.(type) {
		case *Item:
			if o.Removed {
				return
			}
		case *Mobile:
			if o.Removed {
				return
			}
		}
	}
	ExecuteEventHandler(t.event, receiver, source, t.parameter)
}
