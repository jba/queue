// Package queue implements a simple queue.
package queue

import "iter"

type segment[T any] struct {
	items [16]T
	end   int
	next  *segment[T]
}

// Queue is a first-in, first-out collection of values.
type Queue[T any] struct {
	front *segment[T]
	back  *segment[T]
	start int
	len   int
}

// Len returns the number of elements in q.
func (q *Queue[T]) Len() int {
	return q.len
}

// Enqueue adds x to the back of q.
func (q *Queue[T]) Enqueue(x T) {
	if q.back == nil {
		q.back = new(segment[T])
		q.front = q.back
	} else if q.back.end == len(q.back.items) {
		q.back.next = new(segment[T])
		q.back = q.back.next
	}

	q.back.items[q.back.end] = x
	q.back.end++
	q.len++
}

// Peek returns the element at the front of q without removing it.
// It panics if q is empty.
func (q *Queue[T]) Peek() T {
	if q.front == nil {
		panic("queue: peek from empty queue")
	}
	return q.front.items[q.start]
}

// Dequeue removes and returns the element at the front of q.
// It panics if q is empty.
func (q *Queue[T]) Dequeue() T {
	if q.front == nil {
		panic("queue: dequeue from empty queue")
	}

	x := q.front.items[q.start]
	q.start++
	q.len--

	// Zero out for garbage collector.
	var zero T
	q.front.items[q.start-1] = zero

	if q.start == q.front.end {
		q.front = q.front.next
		q.start = 0
		if q.front == nil {
			q.back = nil
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
			start = 0
		}
	}
}

// Clear removes all elements from q.
func (q *Queue[T]) Clear() {
	*q = Queue[T]{}
}
