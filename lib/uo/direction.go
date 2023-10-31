package uo

// A Dir is a 3-bit value indicating the direction a mobile is facing
type Direction byte

// Direction value meanings
const (
	DirectionNorth       Direction = 0
	DirectionNorthEast   Direction = 1
	DirectionEast        Direction = 2
	DirectionSouthEast   Direction = 3
	DirectionSouth       Direction = 4
	DirectionSouthWest   Direction = 5
	DirectionWest        Direction = 6
	DirectionNorthWest   Direction = 7
	DirectionRunningFlag Direction = 0x80
)

// Internal slice of direction offsets for use with GetForwardOffset
var dirOfs = [][]int16{
	{0, -1},
	{1, -1},
	{1, 0},
	{1, 1},
	{0, 1},
	{-1, 1},
	{-1, 0},
	{-1, -1},
}

// Bound returns the direction code bounded to valid values while preserving
// the running flag.
func (d Direction) Bound() Direction {
	isRunning := d.IsRunning()
	dirpart := d & 0x07
	if isRunning {
		return dirpart.SetRunningFlag()
	}
	return dirpart.StripRunningFlag()
}

// IsRunning returns true if the running flag is set
func (d Direction) IsRunning() bool {
	return d&DirectionRunningFlag == DirectionRunningFlag
}

// StripRunningFlag strips the running flag off of a Direction if present
func (d Direction) StripRunningFlag() Direction {
	return d & ^DirectionRunningFlag
}

// SetRunningFlag sets the running flag of a Direction
func (d Direction) SetRunningFlag() Direction {
	return d | DirectionRunningFlag
}

// IsDiagonal returns true if the direction refers to a non-cardinal direction.
func (d Direction) IsDiagonal() bool { return d&0x01 == 0x01 }

// Left returns the next direction counter-clockwise on the compass rose.
func (d Direction) Left() Direction { return (d - 1) & 0x07 }

// Right returns the next direction clockwise on the compass rose.
func (d Direction) Right() Direction { return (d + 1) & 0x07 }
