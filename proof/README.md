# Lean proof

This directory contains a Lean refinement model of the Go queue.

The model represents the active segment chain and the retained spare chain.
Its abstraction function, `State.contents`, flattens the active segments into a
mathematical list.

Proved properties:

- `enqueue_contents`: enqueue appends one value.
- `peek_spec`: a successful peek returns the first value.
- `dequeue_contents`: dequeue returns the first value and leaves the tail.
- `clear_contents`: clear produces an empty abstract queue.
- `empty_countCorrect`, `enqueue_countCorrect`, `dequeue_countCorrect`, and
  `clear_countCorrect`: operations preserve agreement between `Len` and the
  number of abstract elements.

Run the proof with:

```sh
cd proof
lake build
```

## Current boundary

This is an algorithmic refinement proof, not yet a proof of the Go source. The
model encodes `start` by removing consumed prefixes and encodes `end` by storing
only initialized segment prefixes. It does not yet model:

- Go pointers and heap aliasing,
- reachability of `front`, `back`, and `back.next`,
- the `sync.Pool` allocator,
- zeroing references for garbage collection,
- iterator execution and early termination.

The next stage is a heap model with segment identifiers and a refinement map
from concrete Go-style states to this model.
