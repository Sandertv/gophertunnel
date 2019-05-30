package script

import (
	"github.com/yuin/gopher-lua"
)

// Value represents a value that may be fed into a script and taken out of a script. Scripts may operate on
// these values.
type Value struct {
	lua.LValue
}

// ValueOf converts a basic scalar value to a value that may be used to feed into the script. It panics if
// returned if the value's type is not one of bool, uint8, int8, uint16, int16, uint32, int32, uint64, int64,
// uint, int, float32, float64, string or nil.
func ValueOf(value interface{}) Value {
	switch v := value.(type) {
	case nil:
		return Value{lua.LNil}
	case bool:
		return Value{lua.LBool(v)}
	case uint8:
		return Value{lua.LNumber(v)}
	case uint16:
		return Value{lua.LNumber(v)}
	case uint32:
		return Value{lua.LNumber(v)}
	case uint64:
		return Value{lua.LNumber(v)}
	case uint:
		return Value{lua.LNumber(v)}
	case int8:
		return Value{lua.LNumber(v)}
	case int16:
		return Value{lua.LNumber(v)}
	case int32:
		return Value{lua.LNumber(v)}
	case int64:
		return Value{lua.LNumber(v)}
	case int:
		return Value{lua.LNumber(v)}
	case float32:
		return Value{lua.LNumber(v)}
	case float64:
		return Value{lua.LNumber(v)}
	case string:
		return Value{lua.LString(v)}
	}
	panic("ValueOf: unsupported value type")
}

// Interface returns the underlying Go value of a Value. Note that numeric types will always be returned as
// a float64.
func (value Value) Interface() interface{} {
	switch val := value.LValue.(type) {
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case lua.LString:
		return string(val)
	case *lua.LFunction:
		return val
	case *lua.LTable:
		return val
	case *lua.LUserData:
		return val.Value
	}
	return nil
}
