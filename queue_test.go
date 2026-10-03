package queue

import (
	"slices"
	"testing"
)

func TestQueue(t *testing.T) {
	var q Queue[int]
	if got := q.Len(); got != 0 {
		t.Fatalf("initial Len() = %d, want 0", got)
	}

	for i := range 20 {
		q.Enqueue(i)
	}
	if got := q.Len(); got != 20 {
		t.Fatalf("Len() after Enqueue = %d, want 20", got)
	}
	if got := q.Peek(); got != 0 {
		t.Fatalf("Peek() = %d, want 0", got)
	}
	if got := q.Len(); got != 20 {
		t.Fatalf("Len() after Peek = %d, want 20", got)
	}

	var want []int
	for i := range 20 {
		want = append(want, i)
	}
	if got := slices.Collect(q.All()); !slices.Equal(got, want) {
		t.Fatalf("All() = %v, want %v", got, want)
	}

	// Drain the first segment only, to verify that it is unlinked while
	// the second segment remains at the front with start reset to zero.
	for i := range len(q.front.items) {
		if got := q.Dequeue(); got != i {
			t.Fatalf("Dequeue() = %d, want %d", got, i)
		}
	}
	if q.front == nil || q.start != 0 {
		t.Fatalf("front segment was not unlinked: front=%v, start=%d", q.front, q.start)
	}

	var firstTwo []int
	for x := range q.All() {
		firstTwo = append(firstTwo, x)
		if len(firstTwo) == 2 {
			break
		}
	}
	if want := []int{16, 17}; !slices.Equal(firstTwo, want) {
		t.Fatalf("partial All() = %v, want %v", firstTwo, want)
	}

	for i := 16; i < 20; i++ {
		if got := q.Dequeue(); got != i {
			t.Fatalf("Dequeue() = %d, want %d", got, i)
		}
	}
	if q.Len() != 0 || q.front != nil || q.back != nil || q.start != 0 {
		t.Fatalf("empty queue was not reset: len=%d front=%p back=%p start=%d", q.Len(), q.front, q.back, q.start)
	}
}

func TestQueueClear(t *testing.T) {
	var q Queue[*int]
	for i := range 20 {
		q.Enqueue(&i)
	}

	q.Clear()
	if q.Len() != 0 || q.front != nil || q.back != nil || q.start != 0 {
		t.Fatalf("Clear() did not reset queue: len=%d front=%p back=%p start=%d", q.Len(), q.front, q.back, q.start)
	}
	if got := slices.Collect(q.All()); len(got) != 0 {
		t.Fatalf("All() after Clear() = %v, want empty", got)
	}

	x := 42
	q.Enqueue(&x)
	if got := q.Peek(); got != &x {
		t.Fatalf("Peek() after reusing a pooled segment = %p, want %p", got, &x)
	}
}

func TestQueueDequeueEmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Dequeue() on empty queue did not panic")
		}
	}()

	var q Queue[int]
	q.Dequeue()
}

func TestQueuePeekEmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Peek() on empty queue did not panic")
		}
	}()

	var q Queue[int]
	q.Peek()
}
