package script

import "github.com/yuin/gopher-lua"

// Module represents a module in a script. It may either be set by the Go program directly, or pre-loaded so
// that the script itself can 'require()' it.
type Module struct {
	Name      string
	Fields    map[string]lua.LValue
	Functions map[string]lua.LGFunction
}

// NewModule returns a new initialised module using the name passed. It has no fields or functions set, but
// may set to the script.
func NewModule(name string) *Module {
	return &Module{Name: name, Fields: make(map[string]lua.LValue), Functions: make(map[string]lua.LGFunction)}
}

// Func adds a function with a name passed to the module.
func (mod *Module) Func(name string, function lua.LGFunction) *Module {
	mod.Functions[name] = function
	return mod
}

// Field adds a field value with a name passed to the module.
func (mod *Module) Field(name string, value lua.LValue) *Module {
	mod.Fields[name] = value
	return mod
}
