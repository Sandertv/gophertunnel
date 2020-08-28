package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// GameRules reads a map of game rules from Reader r. It sets one of the types 'bool', 'float32' or 'uint32'
// to the map x, with the key being the name of the game rule.
func GameRules(r *Reader, x *map[string]interface{}) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, mediumLimit)

	for i := uint32(0); i < count; i++ {
		// Each of the game rules holds a name and a value type, with the actual value depending on the type
		// that it is.
		var name string
		var valueType uint32

		r.String(&name)
		r.Varuint32(&valueType)
		switch valueType {
		case 1:
			var v bool
			r.Bool(&v)
			(*x)[name] = v
		case 2:
			var v uint32
			r.Varuint32(&v)
			(*x)[name] = v
		case 3:
			var v float32
			r.Float32(&v)
			(*x)[name] = v
		default:
			r.UnknownEnumOption(valueType, "game rule type")
		}
	}
}

// WriteGameRules writes a map of game rules x, indexed by their names to Buffer dst. The types of the map
// values must be either 'bool', 'float32' or 'uint32'. If one of the values has a different type, the
// function will panic.
func WriteGameRules(dst *bytes.Buffer, x map[string]interface{}) error {
	if x == nil {
		return WriteVaruint32(dst, 0)
	}
	// The game rules are always prefixed with a varuint32 indicating the amount.
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return wrap(err)
	}
	for name, value := range x {
		// We first write the name of the game rule.
		if err := WriteString(dst, name); err != nil {
			return wrap(err)
		}
		switch v := value.(type) {
		case bool:
			// Game rule type 1 is for booleans.
			if err := WriteVaruint32(dst, 1); err != nil {
				return wrap(err)
			}
			if err := binary.Write(dst, binary.LittleEndian, v); err != nil {
				return wrap(err)
			}
		case uint32:
			// Game rule type 2 is for varuint32s.
			if err := WriteVaruint32(dst, 2); err != nil {
				return wrap(err)
			}
			if err := WriteVaruint32(dst, v); err != nil {
				return wrap(err)
			}
		case float32:
			// Game rule type 3 is for float32s.
			if err := WriteVaruint32(dst, 3); err != nil {
				return wrap(err)
			}
			if err := WriteFloat32(dst, v); err != nil {
				return wrap(err)
			}
		default:
			panic(fmt.Sprintf("invalid game rule type %T", v))
		}
	}
	return nil
}
