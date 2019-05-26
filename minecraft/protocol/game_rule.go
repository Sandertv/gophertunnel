package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// GameRules reads a map of game rules from Buffer src. It sets one of the types 'bool', 'float32' or 'uint32'
// to the map x, with the key being the name of the game rule.
func GameRules(src *bytes.Buffer, x *map[string]interface{}) error {
	var length uint32
	// The amount of game rules is in a varuint32 before the game rules.
	if err := Varuint32(src, &length); err != nil {
		return err
	}
	for i := uint32(0); i < length; i++ {
		// Each of the game rules holds a name and a value type, with the actual value depending on the type
		// that it is.
		var name string
		if err := String(src, &name); err != nil {
			return err
		}
		var valueType uint32
		if err := Varuint32(src, &valueType); err != nil {
			return err
		}
		switch valueType {
		case 1:
			var v bool
			if err := binary.Read(src, binary.LittleEndian, &v); err != nil {
				return err
			}
			(*x)[name] = v
		case 2:
			var v uint32
			if err := Varuint32(src, &v); err != nil {
				return err
			}
			(*x)[name] = v
		case 3:
			var v float32
			if err := Float32(src, &v); err != nil {
				return err
			}
			(*x)[name] = v
		default:
			// We got a game rule type which doesn't exist, so we return an error immediately.
			return fmt.Errorf("unknown game rule type %v: expected one of 1, 2, or 3", valueType)
		}
	}
	return nil
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
		return err
	}
	for name, value := range x {
		// We first write the name of the game rule.
		if err := WriteString(dst, name); err != nil {
			return err
		}
		switch v := value.(type) {
		case bool:
			// Game rule type 1 is for booleans.
			if err := WriteVaruint32(dst, 1); err != nil {
				return err
			}
			if err := binary.Write(dst, binary.LittleEndian, v); err != nil {
				return err
			}
		case uint32:
			// Game rule type 2 is for varuint32s.
			if err := WriteVaruint32(dst, 2); err != nil {
				return err
			}
			if err := WriteVaruint32(dst, v); err != nil {
				return err
			}
		case float32:
			// Game rule type 3 is for float32s.
			if err := WriteVaruint32(dst, 3); err != nil {
				return err
			}
			if err := WriteFloat32(dst, v); err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("invalid game rule type %T", v))
		}
	}
	return nil
}
