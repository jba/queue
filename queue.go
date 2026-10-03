// Package queue implements a simple queue.
package queue

// The queue retains cleared segments for reuse, making Clear constant time.
// References in retained segments may keep values alive until those segments
// are reused or the queue itself becomes unreachable.

import (
	"iter"
	"sync"
)

type segment[T any] struct {
	items [16]T
	end   int
	next  *segment[T]
}

// Queue is a first-in, first-out collection of values.
// A Queue must not be copied after first use.
type Queue[T any] struct {
	front *segment[T]
	back  *segment[T]
	start int
	len   int
	pool  sync.Pool
}

// Len returns the number of elements in q.
func (q *Queue[T]) Len() int {
	return q.len
}

// Enqueue adds x to the back of q.
func (q *Queue[T]) Enqueue(x T) {
	if q.back == nil {
		q.back = q.getSegment()
		q.front = q.back
	} else if q.back.end == len(q.back.items) {
		if q.back.next == nil {
			q.back.next = q.getSegment()
		}
		q.back = q.back.next
		q.resetSegment(q.back)
	}

	q.back.items[q.back.end] = x
	q.back.end++
	q.len++
}

// Peek returns the element at the front of q without removing it.
// It panics if q is empty.
func (q *Queue[T]) Peek() T {
	if q.len == 0 {
		panic("queue: peek from empty queue")
	}
	return q.front.items[q.start]
}

// Dequeue removes and returns the element at the front of q.
// It panics if q is empty.
func (q *Queue[T]) Dequeue() T {
	if q.len == 0 {
		panic("queue: dequeue from empty queue")
	}

	x := q.front.items[q.start]
	q.start++
	q.len--

	// Zero out for garbage collector.
	var zero T
	q.front.items[q.start-1] = zero

	if q.start == q.front.end {
		empty := q.front
		q.start = 0
		if empty == q.back {
			q.resetSegment(empty)
		} else {
			q.front = empty.next
			q.putSegment(empty)
		}
	}

	return x
}

// All returns an iterator over the elements of q from front to back.
func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		start := q.start
		for s := q.front; s != nil; s = s.next {
			for i := start; i < s.end; i++ {
				if !yield(s.items[i]) {
					return
				}
			}
			if s == q.back {
				return
			}
			start = 0
		}
	}
}

// Clear removes all elements from q in constant time.
// Clear may not release memory associated with the queue. To ensure that the
// memory is released, drop all pointers to the queue so that it can be garbage
// collected.
func (q *Queue[T]) Clear() {
	q.start = 0
	q.len = 0
	q.back = q.front
	if q.back != nil {
		q.resetSegment(q.back)
	}
}

func (q *Queue[T]) getSegment() *segment[T] {
	if s := q.pool.Get(); s != nil {
		return s.(*segment[T])
	}
	return new(segment[T])
}

func (q *Queue[T]) putSegment(s *segment[T]) {
	*s = segment[T]{}
	q.pool.Put(s)
}

func (q *Queue[T]) resetSegment(s *segment[T]) {
	next := s.next
	*s = segment[T]{next: next}
}
