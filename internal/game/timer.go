package game

import (
	"errors"
	"fmt"
	"io"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	timerPools = make([]map[uo.Serial]*Timer, 1+mediumSpeedTimerPoolsCount+lowSpeedTimerPoolsCount)
	for i := 0; i < len(timerPools); i++ {
		timerPools[i] = make(map[uo.Serial]*Timer)
	}
}

// Serial manager for timers
var timerSerialManager *uo.SerialManager

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
func NewTimer(delay uo.Time, oneShot bool, event string, receiver, source Object) *Timer {
	if timerSerialManager == nil {
		timerSerialManager = uo.NewSerialManager(world.Random())
	}
	receiverSerial := uo.SerialZero
	sourceSerial := uo.SerialZero
	if receiver != nil {
		receiverSerial = receiver.Serial()
	}
	if source != nil {
		sourceSerial = source.Serial()
	}
	serial := timerSerialManager.New(uo.SerialTypeUnbound)
	t := &Timer{
		deadline: world.Time() + delay,
		event:    event,
		receiver: receiverSerial,
		source:   sourceSerial,
	}
	pool := 0
	if delay > uo.DurationSecond && delay < uo.DurationMinute {
		pool = 1 + (int(serial) % mediumSpeedTimerPoolsCount)
	} else {
		pool = 1 + mediumSpeedTimerPoolsCount + (int(serial) % lowSpeedTimerPoolsCount)
	}
	timerPools[pool][serial] = t
	return t
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

// WriteTimers writes all timers to an io.WriteCloser in list file format. This
// function always returns a nil slice. This is for compatibility with other
// similar functions.
func WriteTimers(w io.WriteCloser) []error {
	lfw := util.NewListFileWriter(w)
	defer lfw.Close()

	lfw.WriteComment("compact timer data")
	lfw.WriteBlankLine()
	lfw.WriteSegmentHeader("Timers")
	for pool, timers := range timerPools {
		for s, t := range timers {
			lfw.WriteLine(fmt.Sprintf("%d,%d,%d,%s,%d,%d", s, pool, t.deadline, t.event, t.receiver, t.source))
		}
	}
	lfw.WriteBlankLine()
	lfw.WriteComment("END OF FILE")
	return nil
}

// ReadTimers reads all timers from an io.Reader in list file format
func ReadTimers(r io.Reader) []error {
	var ret []error
	lfr := &util.ListFileReader{}
	lfr.StartReading(r)
	header := lfr.StreamNextSegmentHeader()
	if header == "" {
		return append(lfr.Errors(), errors.New("expected segment header while reading timers"))
	}
	for {
		line := lfr.StreamNextEntry()
		if line == "" {
			break
		}
		var serial, pool, deadline, receiver, source int
		var event string
		if n, err := fmt.Sscanf(line, "%d,%d,%d,%s,%d,%d",
			&serial, &pool, &deadline, &event, &receiver, &source); n != 6 || err != nil {
			if err != nil {
				ret = append(ret, err)
			}
			continue
		}
		s := uo.Serial(serial)
		if pool < 0 || pool >= len(timerPools) {
			ret = append(ret, fmt.Errorf("timer %s pool %d out of range", s.String(), pool))
			continue
		}
		if _, duplicate := timerPools[pool][s]; duplicate {
			ret = append(ret, fmt.Errorf("timer %s is a duplicate in pool %d", s.String(), pool))
			continue
		}
		timerPools[pool][s] = &Timer{
			deadline: uo.Time(deadline),
			event:    event,
			receiver: uo.Serial(receiver),
			source:   uo.Serial(source),
		}
	}
	return append(ret, lfr.Errors()...)
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
	ExecuteEventHandler(t.event, receiver, source)
}
