package franchise

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"testing"
)

func TestDiscover(t *testing.T) {
	d, err := Discover(protocol.CurrentVersion)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", d)

	a := new(AuthorizationEnvironment)
	if err := d.Environment(a, EnvironmentTypeProduction); err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", a)
}
