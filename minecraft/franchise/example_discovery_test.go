package franchise_test

import (
	"log"

	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func ExampleDiscovery() {
	// Discover obtains a Discovery for the specific version, which includes environments for various franchise services.
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		log.Fatalf("Error obtaining discovery: %s", err)
	}

	// Look up and decode an environment for authorization.
	auth := new(franchise.AuthorizationEnvironment)
	if err := discovery.Environment(auth, franchise.EnvironmentTypeProduction); err != nil {
		log.Fatalf("Error reading environment for %q: %s", auth.EnvironmentName(), err)
	}

	// Use discovery and auth for further use.
	_ = discovery
	_ = auth
}
