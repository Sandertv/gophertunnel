package protocol

// InputFlags is a set of input flags sent in a PlayerAuthInput packet. On the wire it is an optional list of
// the IDs of the flags that are set, but in memory it is backed by a Bitset so that individual flags may be
// tested and modified in constant time.
//
// An InputFlags may be absent, which is distinct from being present but empty: the former is not sent over the
// wire at all. The zero value is absent and has no size, so it must be created using NewInputFlags before any
// flag may be set on it. Values decoded from the wire are always sized, absent or not.
type InputFlags struct {
	set  bool
	bits Bitset
}

// NewInputFlags creates a present InputFlags with no flags set, able to hold flags in the range [0, size).
// Setting a flag at an index beyond size will panic.
func NewInputFlags(size int) InputFlags {
	return InputFlags{set: true, bits: NewBitset(size)}
}

// Present returns whether the InputFlags was sent over the wire. An absent InputFlags reports every flag as
// unset.
func (f InputFlags) Present() bool {
	return f.set
}

// Load returns whether the flag at index i is set. It returns false if the InputFlags is absent. If i is
// beyond the size of a present InputFlags, a panic will occur.
func (f InputFlags) Load(i int) bool {
	if !f.set {
		return false
	}
	return f.bits.Load(i)
}

// Set sets the flag at index i, marking the InputFlags as present. If i is beyond the size of the InputFlags,
// a panic will occur.
func (f *InputFlags) Set(i int) {
	f.bits.Set(i)
	f.set = true
}

// Unset unsets the flag at index i. It is a no-op if the InputFlags is absent. If i is beyond the size of a
// present InputFlags, a panic will occur.
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

// InputFlagList reads/writes an InputFlags x holding flags in the range [0, size) as an optional list of the
// IDs of the flags that are set. Flag IDs must be unique and within range: a Reader rejects input that is not.
func InputFlagList(r IO, x *InputFlags, size int) {
	ids := x.ids()
	present := x.set
	r.Bool(&present)
	if !present {
		*x = InputFlags{bits: NewBitset(size)}
		return
	}

	// Bound the count by the amount of distinct flags that can legally be sent before any of them are read, so
	// that a peer cannot inflate the allocation up to maxSliceLength.
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
