package uo

// A uo.Time value represents the number of seconds since the beginning of
// the Sossarian universe. One second of Sossarian time equals about 1/12 second
// real-world time.
type Time uint64

// The zero value for Time
const TimeZero Time = 0

// The start of the Ultima Online game services. A brand-new server starts at
// this time.
const TimeEpoch Time = 289 /*years*/ * 876 /*days per year*/ * 24 /*hours in a day*/ * 60 /*minutes in an hour*/ * 60 /*seconds in a minute*/

// A value meaning "will never happen"
const TimeNever Time = Time(^uint64(0))

// Ticks per real-world second
const DurationSecond Time = 12

// Ticks per real-world minute
const DurationMinute Time = DurationSecond * 60

// Ticks per real-world hour
const DurationHour Time = DurationMinute * 60

// Ticks per real-world day
const DurationDay Time = DurationHour * 24
