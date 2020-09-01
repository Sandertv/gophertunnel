package protocol

import (
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

// WriteGameRules writes a map of game rules x, indexed by their names to Writer w. The types of the map
// values must be either 'bool', 'float32' or 'uint32'. If one of the values has a different type, the
// function will panic.
func WriteGameRules(w *Writer, x *map[string]interface{}) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	for name, value := range *x {
		w.String(&name)
		switch v := value.(type) {
		case bool:
			id := uint32(1)
			w.Varuint32(&id)
			w.Bool(&v)
		case uint32:
			id := uint32(2)
			w.Varuint32(&id)
			w.Varuint32(&v)
		case float32:
			id := uint32(3)
			w.Varuint32(&id)
			w.Float32(&v)
		default:
			w.UnknownEnumOption(fmt.Sprintf("%T", value), "game rule type")
		}
	}
}
