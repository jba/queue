import Std

/-!
A first refinement model of the Go queue.

`active` models the linked segments from `front` through `back`. `spare`
models the retained chain after `back`. A segment is represented by the list
of its logically initialized items; the Go `start` and `end` indexes are
encoded by removing consumed prefixes and tracking only initialized prefixes.

The abstraction function `contents` flattens the active segments. The main
theorems show that the modeled operations refine ordinary list operations.
This does not yet prove the Go heap manipulation or `sync.Pool` behavior.
-/

namespace QueueProof

abbrev capacity : Nat := 8

structure State (α : Type) where
  active : List (List α)
  spare : List (List α)
  count : Nat
  deriving Repr

namespace State

/-- The mathematical sequence represented by the queue. -/
def contents (q : State α) : List α := q.active.flatten

/-- The stored count agrees with the mathematical sequence. -/
def CountCorrect (q : State α) : Prop := q.count = q.contents.length

def empty : State α := { active := [], spare := [], count := 0 }

/-- Append to the last active segment, reusing a spare segment when possible. -/
def enqueueSegments : List (List α) → List (List α) → α → List (List α) × List (List α)
  | [], spare, x => ([[x]], spare)
  | [segment], spare, x =>
      if segment.length < capacity then
        ([segment ++ [x]], spare)
      else
        match spare with
        | [] => ([segment, [x]], [])
        | _ :: remaining => ([segment, [x]], remaining)
  | segment :: next :: remaining, spare, x =>
      let result := enqueueSegments (next :: remaining) spare x
      (segment :: result.1, result.2)

theorem enqueueSegments_contents (segments spare : List (List α)) (x : α) :
    (enqueueSegments segments spare x).1.flatten = segments.flatten ++ [x] := by
  induction segments with
  | nil => simp [enqueueSegments]
  | cons segment segments ih =>
      cases segments with
      | nil =>
          simp only [enqueueSegments]
          split
          · simp
          · cases spare <;> simp
      | cons next remaining =>
          simp [enqueueSegments, ih]

def enqueue (q : State α) (x : α) : State α :=
  let result := enqueueSegments q.active q.spare x
  { active := result.1, spare := result.2, count := q.count + 1 }

/-- Enqueue refines appending one element to a list. -/
theorem enqueue_contents (q : State α) (x : α) :
    (enqueue q x).contents = q.contents ++ [x] := by
  simp [enqueue, contents, enqueueSegments_contents]

def peekSegments (segments : List (List α)) : Option α :=
  segments.head?.bind List.head?

theorem peekSegments_spec (segments : List (List α)) (x : α)
    (h : peekSegments segments = some x) :
    segments.flatten.head? = some x := by
  cases segments with
  | nil => simp [peekSegments] at h
  | cons segment rest =>
      cases segment with
      | nil => simp [peekSegments] at h
      | cons first remaining =>
          simp [peekSegments] at h
          simp [h]

def peek (q : State α) : Option α := peekSegments q.active

/-- A successful peek returns the head of the abstract sequence. -/
theorem peek_spec (q : State α) (x : α) (h : q.peek = some x) :
    q.contents.head? = some x :=
  peekSegments_spec q.active x h

def dequeueSegments : List (List α) → Option (α × List (List α))
  | [] => none
  | [] :: _ => none
  | (x :: remaining) :: rest =>
      match remaining, rest with
      | [], [] => some (x, [[]])
      | [], _ :: _ => some (x, rest)
      | _ :: _, _ => some (x, remaining :: rest)

theorem dequeueSegments_spec (segments : List (List α)) (x : α) (remaining : List (List α))
    (h : dequeueSegments segments = some (x, remaining)) :
    segments.flatten = x :: remaining.flatten := by
  cases segments with
  | nil => simp [dequeueSegments] at h
  | cons segment rest =>
      cases segment with
      | nil => simp [dequeueSegments] at h
      | cons first tail =>
          cases tail with
          | nil =>
              cases rest with
              | nil =>
                  simp [dequeueSegments] at h
                  obtain ⟨rfl, rfl⟩ := h
                  rfl
              | cons next more =>
                  simp [dequeueSegments] at h
                  obtain ⟨rfl, rfl⟩ := h
                  rfl
          | cons next more =>
              simp [dequeueSegments] at h
              obtain ⟨rfl, rfl⟩ := h
              rfl

def dequeue (q : State α) : Option (α × State α) :=
  match dequeueSegments q.active with
  | none => none
  | some (x, active) => some (x, { q with active, count := q.count - 1 })

/-- A successful dequeue returns the abstract head and leaves the abstract tail. -/
theorem dequeue_contents (q : State α) (x : α) (q' : State α)
    (h : q.dequeue = some (x, q')) :
    q.contents = x :: q'.contents := by
  unfold dequeue at h
  cases hs : dequeueSegments q.active with
  | none => simp [hs] at h
  | some result =>
      rcases result with ⟨first, active⟩
      simp [hs] at h
      obtain ⟨rfl, rfl⟩ := h
      exact dequeueSegments_spec q.active first active hs

theorem dequeue_count (q : State α) (x : α) (q' : State α)
    (h : q.dequeue = some (x, q')) :
    q'.count = q.count - 1 := by
  unfold dequeue at h
  cases hs : dequeueSegments q.active with
  | none => simp [hs] at h
  | some result =>
      rcases result with ⟨first, active⟩
      simp [hs] at h
      obtain ⟨rfl, rfl⟩ := h
      rfl

/-- Dequeue preserves agreement between the stored count and abstract length. -/
theorem dequeue_countCorrect (q : State α) (x : α) (q' : State α)
    (hc : q.CountCorrect) (hd : q.dequeue = some (x, q')) :
    q'.CountCorrect := by
  unfold CountCorrect at hc ⊢
  rw [dequeue_count q x q' hd]
  have hcontents := dequeue_contents q x q' hd
  rw [hcontents] at hc
  simp at hc
  omega

/-- Clear retains the first segment and moves later active segments to the spare chain. -/
def clear (q : State α) : State α :=
  match q.active with
  | [] => { q with count := 0 }
  | _ :: remaining => { active := [[]], spare := remaining ++ q.spare, count := 0 }

/-- Clear makes the abstract queue empty. -/
theorem clear_contents (q : State α) : q.clear.contents = [] := by
  cases q with
  | mk active spare count =>
      cases active <;> simp [clear, contents]

theorem empty_countCorrect : CountCorrect (empty : State α) := by
  simp [CountCorrect, empty, contents]

theorem enqueue_countCorrect (q : State α) (x : α) (h : q.CountCorrect) :
    (q.enqueue x).CountCorrect := by
  unfold CountCorrect at h ⊢
  rw [enqueue_contents]
  simp [enqueue, h]

theorem clear_countCorrect (q : State α) : q.clear.CountCorrect := by
  unfold CountCorrect
  rw [clear_contents]
  cases q with
  | mk active spare count =>
      cases active <;> simp [clear]

end State
end QueueProof
