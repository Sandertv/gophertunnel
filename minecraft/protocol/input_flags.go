package protocol

// InputFlags is the set of input flags in a PlayerAuthInput packet. It is sent as a list of the flag IDs that
// are set, but is backed by a Bitset so that flags can be read and changed in constant time.
//
// An absent InputFlags is not the same as an empty one: it is not sent at all. The zero value is absent and
// has no size, so use NewInputFlags before setting a flag on it. Decoded values are always sized.
type InputFlags struct {
	set  bool
	bits Bitset
}

// NewInputFlags creates an InputFlags holding flags in the range [0, size), with none of them set.
func NewInputFlags(size int) InputFlags {
	return InputFlags{set: true, bits: NewBitset(size)}
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
	return f.bits.Load(i)
}

// Set sets the flag at index i and marks the InputFlags as present. Indices beyond the size panic.
func (f *InputFlags) Set(i int) {
	f.bits.Set(i)
	f.set = true
}

// Unset unsets the flag at index i, doing nothing if the InputFlags is absent. Indices beyond the size panic.
func (f *InputFlags) Unset(i int) {
	if !f.set {
		return
	}
	f.bits.Unset(i)
}

// Len returns the amount of flags the InputFlags can hold.
func (f InputFlags) Len() int {
	return f.bits.Len()
}

// ids returns the indices of the flags that are set, in ascending order.
func (f InputFlags) ids() []int32 {
	if !f.set || f.bits.int == nil {
		return nil
	}
	ids := make([]int32, 0, f.bits.int.BitLen())
	for i := 0; i < f.bits.size; i++ {
		if f.bits.Load(i) {
			ids = append(ids, int32(i))
		}
	}
	return ids
}

// InputFlagList reads/writes an InputFlags holding flags in the range [0, size) as a list of the flag IDs that
// are set. Flag IDs must be unique and within range.
func InputFlagList(r IO, x *InputFlags, size int) {
	ids := x.ids()
	present := x.set
	r.Bool(&present)
	if !present {
		*x = InputFlags{bits: NewBitset(size)}
		return
	}

	// Check the count before reading, so a peer cannot claim far more flags than can legally be sent.
	count := uint32(len(ids))
	r.Varuint32(&count)
	if count > uint32(size) {
		r.InvalidValue(count, "player auth input data", "too many flags")
		return
	}
	FuncSliceOfLen(r, count, &ids, r.Varint32)

	out := NewInputFlags(size)
	for _, id := range ids {
		switch {
		case id < 0 || int(id) >= size:
			r.UnknownEnumOption(id, "player auth input data")
		case out.bits.Load(int(id)):
			r.InvalidValue(id, "player auth input data", "flags must be unique")
		default:
			out.bits.Set(int(id))
		}
	}
	*x = out
}
