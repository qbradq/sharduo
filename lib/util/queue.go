package util

// Queue manages a queue of values with a maximum capacity.
type Queue[V any] struct {
	// Queue of values
	v []V
	// Capacity of the Queue
	cap int
}

// NewQueue creates a new queue with the given capacity.
func NewQueue[V any](cap int) *Queue[V] {
	return &Queue[V]{
		v:   make([]V, 0, cap),
		cap: cap,
	}
}

// Enqueue places a value at the end of the queue and returns true on success.
func (q *Queue[V]) Enqueue(v V) bool {
	if len(q.v) >= q.cap {
		return false
	}
	q.v[len(q.v)-1] = v
	return true
}

// Dequeue removes and returns the first element of the queue.
func (q *Queue[V]) Dequeue() V {
	var zero V
	if len(q.v) == 0 {
		return zero
	}
	v := q.v[0]
	copy(q.v[:len(q.v)-1], q.v[1:])
	q.v[len(q.v)-1] = zero
	q.v = q.v[:len(q.v)-1]
	return v
}

// Length returns the current number of values in the queue.
func (q *Queue[V]) Length() int { return len(q.v) }
