package util

import "github.com/qbradq/sharduo/lib/uo"

type SliceElement interface {
	Serial() uo.Serial
}

// Slice is a generic wrapper around slices of objects
type Slice[T SliceElement] []T

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

// Insert inserts a value at the given index
func (s Slice[T]) Insert(idx int, v T) Slice[T] {
	if idx == len(s) {
		return append(s, v)
	}
	s = append(s[:idx+1], s[idx:]...)
	s[idx] = v
	return s
}
