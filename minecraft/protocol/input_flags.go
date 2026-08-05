package protocol

import "slices"

// InputFlags is the set of input flags in a PlayerAuthInput packet. It is sent as a list of the IDs of the
// flags that are set, and is stored the same way, with the IDs kept unique.
//
// An absent InputFlags is not the same as an empty one: it is not sent at all. The zero value is absent and
// has no size, so use NewInputFlags before setting a flag on it. Decoded values are always sized.
type InputFlags struct {
	set  bool
	size int
	ids  []int32
}

// NewInputFlags creates an InputFlags holding flags in the range [0, size), with none of them set.
func NewInputFlags(size int) InputFlags {
	return InputFlags{set: true, size: size}
}

// NewInputFlagsFromIDs creates an InputFlags holding the IDs passed. Duplicate IDs are ignored.
func NewInputFlagsFromIDs(size int, ids []int32) InputFlags {
	flags := NewInputFlags(size)
	for _, id := range ids {
		flags.Set(int(id))
	}
	return flags
}

// Present returns whether the InputFlags was sent. An absent one reports every flag as unset.
func (f InputFlags) Present() bool {
	return f.set
}

// Load returns whether the flag at index i is set. Indices beyond the size panic.
func (f InputFlags) Load(i int) bool {
	if !f.set {
		return false
	}
	f.check(i)
	return slices.Contains(f.ids, int32(i))
}

// Set sets the flag at index i and marks the InputFlags as present. Setting a flag that is already set does
// nothing, as the flags sent must be unique. Indices beyond the size panic.
func (f *InputFlags) Set(i int) {
	f.check(i)
	f.set = true
	if !slices.Contains(f.ids, int32(i)) {
		f.ids = append(f.ids, int32(i))
	}
}

// Unset unsets the flag at index i, doing nothing if the InputFlags is absent. Indices beyond the size panic.
func (f *InputFlags) Unset(i int) {
	if !f.set {
		return
	}
	f.check(i)
	if at := slices.Index(f.ids, int32(i)); at != -1 {
		f.ids = slices.Delete(f.ids, at, at+1)
	}
}

// Len returns the amount of flags the InputFlags can hold.
func (f InputFlags) Len() int {
	return f.size
}

func (f InputFlags) check(i int) {
	if i < 0 || i >= f.size {
		panic("index out of bounds")
	}
}

// InputFlagList reads/writes an InputFlags holding flags in the range [0, size) as a list of the IDs of the
// flags that are set. Flag IDs must be unique and within range.
func InputFlagList(r IO, x *InputFlags, size int) {
	present := x.set
	r.Bool(&present)
	if !present {
		*x = InputFlags{size: size}
		return
	}

	// Check the count before reading, so a peer cannot claim far more flags than can legally be sent.
	count := uint32(len(x.ids))
	r.Varuint32(&count)
	if count > uint32(size) {
		r.InvalidValue(count, "player auth input data", "too many flags")
		return
	}
	ids := x.ids
	FuncSliceOfLen(r, count, &ids, r.Varint32)

	for n, id := range ids {
		if id < 0 || int(id) >= size {
			r.UnknownEnumOption(id, "player auth input data")
		} else if slices.Contains(ids[:n], id) {
			r.InvalidValue(id, "player auth input data", "flags must be unique")
		}
	}
	*x = InputFlags{set: true, size: size, ids: ids}
}
