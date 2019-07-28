package script

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"log"
	"os"
)

// Script is a wrapper around a Lua state. It provides methods to easily add globals and modules to the
// script from the Go side.
type Script struct {
	// ErrorLog is a log.Logger that errors that occur during the interpreting of Lua code. By default, it is
	// set to the package level log variable.
	ErrorLog *log.Logger

	state  *lua.LState
	closed bool
}

// New returns a new initialised script. It does not run any code yet.
func New() *Script {
	return &Script{state: lua.NewState(), ErrorLog: log.New(os.Stderr, "", log.LstdFlags)}
}

// RunString runs a string of Lua source code passed. Multiple strings may be ran successively in order to
// create a continuously running script.
func (script *Script) RunString(lua string) error {
	if script.closed {
		return fmt.Errorf("script closed")
	}
	return script.state.DoString(lua)
}

// RunFile runs a file containing Lua source code at the path passed.
func (script *Script) RunFile(file string) error {
	if script.closed {
		return fmt.Errorf("script closed")
	}
	return script.state.DoFile(file)
}

// Close closes the script, making it stop processing any code.
func (script *Script) Close() error {
	script.state = nil
	script.closed = true
	return nil
}

// SetGlobal sets a global value to the script, that the script may use during execution.
func (script *Script) SetGlobal(name string, value lua.LValue) {
	script.state.SetGlobal(name, value)
}

// Global returns a global value from the script. The global may be either set from the Go side or the
// script's side.
func (script *Script) Global(name string) (v Value, ok bool) {
	value := script.state.GetGlobal(name)
	if value == lua.LNil {
		return Value{}, false
	}
	return Value{value}, true
}

// SetModule sets a script module so that the script can directly access it by the name of the module. No
// calls to 'require()' need to be made.
func (script *Script) SetModule(mod *Module) {
	modTable := script.state.RegisterModule(mod.Name, mod.Functions).(*lua.LTable)
	for fieldName, field := range mod.Fields {
		script.state.SetField(modTable, fieldName, field)
	}
}

// SetModule pre-loads a script module so that it may be imported in the script using 'require()'. It must
// be done by the script itself.
// For the module to be set directly so that the script doesn't have to use 'require()', see SetModule.
func (script *Script) PreloadModule(mod *Module) {
	script.state.PreloadModule(mod.Name, func(L *lua.LState) int {
		table := L.NewTable()
		L.SetFuncs(table, mod.Functions)
		for fieldName, field := range mod.Fields {
			L.SetField(table, fieldName, field)
		}
		L.Push(table)
		return 1
	})
}

// RunStdin starts interpreting the stdin as Lua code. It reports any errors during interpreting to the
// ErrorLog logger of the script.
// Multiline code may be encapsulated within double quotes ("") to await execution until the ending double
// quotes have been written.
func (script *Script) RunStdin() {
	if distributor == nil {
		distributor = Distribute(os.Stdin)
	}
	distributor.Subscribe(script)
}
