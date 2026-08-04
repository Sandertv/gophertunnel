package protocol

import (
	"bytes"
	"testing"
)

const (
	testFlagSize = 66

	// testFlagAscend and testFlagJumping mirror packet.InputFlagAscend and packet.InputFlagJumping, which
	// cannot be referenced here as the packet package imports this one.
	testFlagAscend  = 0
	testFlagJumping = 6
)

// writeFlagList encodes ids the way the wire format specifies it, without going through InputFlags, so that
// the InputFlags codec can be compared against an independent encoding of the same data.
func writeFlagList(t *testing.T, present bool, ids []int32) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	w := NewWriter(buf, 0)
	w.Bool(&present)
	if present {
		count := uint32(len(ids))
		w.Varuint32(&count)
		for i := range ids {
			w.Varint32(&ids[i])
		}
	}
	return buf.Bytes()
}

func encodeFlags(t *testing.T, f InputFlags) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	w := NewWriter(buf, 0)
	InputFlagList(w, &f, testFlagSize)
	return buf.Bytes()
}

func decodeFlags(t *testing.T, b []byte) (f InputFlags, err error) {
	t.Helper()

	r := NewReader(bytes.NewReader(b), 0, true)
	defer func() {
		if recovered := recover(); recovered != nil {
			err, _ = recovered.(error)
			if err == nil {
				t.Fatalf("panic with non-error value: %v", recovered)
			}
		}
	}()
	InputFlagList(r, &f, testFlagSize)
	return f, nil
}

// TestInputFlagListWireFormat asserts that encoding an InputFlags produces exactly the bytes of the optional
// flag ID list that 1.26.40 specifies, so that backing the field with a Bitset does not alter the wire format.
func TestInputFlagListWireFormat(t *testing.T) {
	for _, test := range []struct {
		name string
		ids  []int32
	}{
		{name: "empty", ids: []int32{}},
		{name: "single", ids: []int32{testFlagJumping}},
		{name: "low and high", ids: []int32{testFlagAscend, testFlagSize - 1}},
		{name: "spanning the varint boundary", ids: []int32{63, 64, 65}},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := NewInputFlags(testFlagSize)
			for _, id := range test.ids {
				f.Set(int(id))
			}

			got, want := encodeFlags(t, f), writeFlagList(t, true, test.ids)
			if !bytes.Equal(got, want) {
				t.Fatalf("encoded %x, want %x", got, want)
			}

			decoded, err := decodeFlags(t, got)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !decoded.Present() {
				t.Fatal("decoded flags are absent")
			}
			for i := 0; i < testFlagSize; i++ {
				want := false
				for _, id := range test.ids {
					if int(id) == i {
						want = true
					}
				}
				if decoded.Load(i) != want {
					t.Fatalf("flag %v: got %v, want %v", i, decoded.Load(i), want)
				}
			}
		})
	}
}

// TestInputFlagListAbsent asserts that an absent set stays absent across a round trip rather than being
// silently promoted to a present, empty one.
func TestInputFlagListAbsent(t *testing.T) {
	var f InputFlags
	got, want := encodeFlags(t, f), writeFlagList(t, false, nil)
	if !bytes.Equal(got, want) {
		t.Fatalf("encoded %x, want %x", got, want)
	}

	decoded, err := decodeFlags(t, got)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Present() {
		t.Fatal("absent flags decoded as present")
	}
	if decoded.Load(testFlagJumping) {
		t.Fatal("absent flags reported a set flag")
	}
	// An absent set decoded from the wire must still be sized, so that a consumer setting a flag on it does
	// not panic on a peer's choice not to send the field.
	decoded.Set(testFlagJumping)
	if !decoded.Present() || !decoded.Load(testFlagJumping) {
		t.Fatal("setting a flag on absent flags did not take effect")
	}
}

func TestInputFlagListRejectsInvalid(t *testing.T) {
	for _, test := range []struct {
		name string
		b    []byte
	}{
		{name: "out of range", b: writeFlagList(t, true, []int32{testFlagSize})},
		{name: "negative", b: writeFlagList(t, true, []int32{-1})},
		{name: "duplicate", b: writeFlagList(t, true, []int32{testFlagJumping, testFlagJumping})},
		{name: "count beyond size", b: append([]byte{1, 0xff, 0x01}, bytes.Repeat([]byte{0}, 255)...)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodeFlags(t, test.b); err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

// TestInputFlagsZeroValue asserts the documented contract of the zero value: absent, reporting nothing set.
func TestInputFlagsZeroValue(t *testing.T) {
	var f InputFlags
	if f.Present() {
		t.Fatal("zero value is present")
	}
	if f.Load(testFlagJumping) {
		t.Fatal("zero value reported a set flag")
	}
	f.Unset(testFlagJumping)
	if f.Present() {
		t.Fatal("unsetting a flag made the zero value present")
	}
}
