package util

// Slice is a generic wrapper around slices of objects
type Slice[T Serialer] []T

// Append appends a value to the slice and returns the slice
func (s Slice[T]) Append(v T) Slice[T] {
	return append(s, v)
}

// IndexOf returns the index of the first of that value in the slice, or -1
func (s Slice[T]) IndexOf(v T) int {
	for idx, iv := range s {
		if iv.Serial() == v.Serial() {
			return idx
		}
	}
	return -1
}

// Remove returns the slice with the first value v removed
func (s Slice[T]) Remove(v T) Slice[T] {
	var zero T
	idx := s.IndexOf(v)
	if idx < 0 {
		return s
	}
	copy(s[idx:], s[idx+1:])
	s[len(s)-1] = zero
	return s[:len(s)-1]
}

// Contains returns true if the value v is found in the slice
func (s Slice[T]) Contains(v T) bool {
	return s.IndexOf(v) >= 0
}

// Copy returns a copy of the slice
func (s Slice[T]) Copy() Slice[T] {
	r := make([]T, len(s))
	copy(r, s)
	return r
}
