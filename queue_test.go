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
	if q.Len() != 0 || q.front == nil || q.back != q.front || q.start != 0 || q.front.end != 0 {
		t.Fatalf("empty queue was not reset: len=%d front=%p back=%p start=%d end=%d", q.Len(), q.front, q.back, q.start, q.front.end)
	}

	// Filling the retained segment forces the next segment to come from the pool.
	for i := range 17 {
		q.Enqueue(i)
	}
	if got := slices.Collect(q.All()); !slices.Equal(got, want[:17]) {
		t.Fatalf("All() after segment reuse = %v, want %v", got, want[:17])
	}
}

func TestQueueClear(t *testing.T) {
	var q Queue[*int]
	for i := range 20 {
		q.Enqueue(&i)
	}
	front := q.front
	spare := front.next

	q.Clear()
	if q.Len() != 0 || q.front != front || q.back != front || q.start != 0 || q.front.end != 0 {
		t.Fatalf("Clear() did not reset queue: len=%d front=%p back=%p start=%d end=%d", q.Len(), q.front, q.back, q.start, q.front.end)
	}
	if q.front.next != spare {
		t.Fatal("Clear() did not retain the spare segment")
	}
	if got := slices.Collect(q.All()); len(got) != 0 {
		t.Fatalf("All() after Clear() = %v, want empty", got)
	}

	values := make([]int, 17)
	for i := range values {
		q.Enqueue(&values[i])
	}
	if q.back != spare {
		t.Fatal("Enqueue did not reuse the spare segment")
	}
	if got := q.Peek(); got != &values[0] {
		t.Fatalf("Peek() after segment reuse = %p, want %p", got, &values[0])
	}
	for i := 1; i < len(spare.items); i++ {
		if spare.items[i] != nil {
			t.Fatalf("reused segment item %d was not cleared", i)
		}
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
