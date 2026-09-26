package gatherings_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/df-mc/go-playfab/v2"
	"github.com/df-mc/go-xsapi/v2"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/gatherings"
)

func ExampleClient() {
	discovery, err := service.Default(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("error discovering service endpoints: %s", err))
	}
	env := new(service.AuthorizationEnvironment)
	if err := discovery.Environment(env); err != nil {
		panic(fmt.Sprintf("error resolving environment for %q: %s", env.ServiceName(), err))
	}

	xbl, err := xsapi.NewClient(auth.AndroidConfig.New(auth.AndroidConfig.WriterTokenSource(os.Stdout), nil))
	if err != nil {
		panic(fmt.Sprintf("error logging in to xbox live: %s", err))
	}
	defer xbl.Close()

	pf, err := playfab.LoginWithXbox(context.TODO(), env.PlayFabTitleID, xbl, playfab.ClientConfig{
		CreateAccount: true,
	})
	if err != nil {
		panic(fmt.Sprintf("error logging in to playfab account: %s", err))
	}
	defer pf.Close()

	src := env.TokenSource(pf, service.TokenConfig{})

	client := gatherings.NewClient(src)
	experiences, err := client.Experiences(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("error searching for experiences: %s", err))
	}
	if len(experiences) == 0 {
		panic("no experiences found")
	}
	experience := experiences[0]

	address, err := experience.Join(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("error joining experience: %s", err))
	}

	conn, err := minecraft.Dialer{
		XBLClient:     xbl,
		PlayFabClient: pf,
	}.DialTimeout("raknet", address.String(), time.Minute*5)
	if err != nil {
		panic(fmt.Sprintf("error connecting to experience: %s", err))
	}
	defer conn.Close()

	if err := conn.DoSpawn(); err != nil {
		panic(fmt.Sprintf("error spawning to experience: %s", err))
	}
}
