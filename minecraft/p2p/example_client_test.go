package p2p

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-playfab/v2"
	"github.com/df-mc/go-xsapi/v2"
	"github.com/df-mc/go-xsapi/v2/xal/sisu"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling"
	"github.com/sandertv/gophertunnel/minecraft/service/signaling/messaging"
)

// ExampleClient lists the worlds and joins the first world in the result.
func ExampleClient() {
	token, err := auth.RequestLiveToken()
	if err != nil {
		panic(err)
	}
	msa := auth.AndroidConfig.TokenSource(context.TODO(), token)

	xbl, err := xsapi.ClientConfig{
		RTAMode: xsapi.RTALazy,
	}.New(context.TODO(), auth.AndroidConfig.New(msa, nil))
	if err != nil {
		var acct *sisu.AccountCreationRequiredError
		if errors.As(err, &acct) {
			fmt.Printf("create an Xbox Live account at %s", acct.SignupURL)
			return
		}
		panic(fmt.Sprintf("error logging in to Xbox Live network services: %s", err))
	}
	defer xbl.Close()

	discovery, err := service.Default(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("error retrieving default discovery data: %s", err))
	}
	env := new(service.AuthorizationEnvironment)
	if err := discovery.Environment(env); err != nil {
		panic(fmt.Sprintf("error resolving environment for %q: %s", env.ServiceName(), err))
	}

	pf, err := playfab.LoginWithXbox(context.TODO(), env.PlayFabTitleID, xbl, playfab.ClientConfig{
		CreateAccount: true,
	})
	if err != nil {
		panic(fmt.Sprintf("error logging in to playfab account: %s", err))
	}
	defer pf.Close()

	src := env.TokenSource(pf, service.TokenConfig{})

	client := NewClient(xbl)
	worlds, err := client.Worlds(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("error listing worlds: %s", err))
	}
	if len(worlds) == 0 {
		panic("no open worlds")
	}

	world := worlds[0]
	session, err := world.Join(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("error joining world: %s", err))
	}
	defer session.Close()

	connection := session.Connection()
	var s nethernet.Signaling
	switch connection.Type {
	case ConnectionTypeSignalingOverJSONRPC:
		var d messaging.Dialer
		conn, err := d.DialContext(context.TODO(), src)
		if err != nil {
			panic(fmt.Sprintf("error dialing messaging conn: %s", err))
		}
		defer conn.Close()
		s = conn
	case ConnectionTypeSignalingOverWebSocket:
		var d signaling.Dialer
		conn, err := d.DialContext(context.TODO(), src)
		if err != nil {
			panic(fmt.Sprintf("error dialing signaling conn: %s", err))
		}
		defer conn.Close()
		s = conn
	default:
		panic(fmt.Sprintf("invalid connection type: %d", connection.Type))
	}

	address := connection.Address()

	minecraft.RegisterNetwork("nethernet", func(l *slog.Logger) minecraft.Network {
		return minecraft.NetherNet{
			Signaling: s,
			Log:       l,
		}
	})

	conn, err := minecraft.Dialer{
		XBLClient:     xbl,
		PlayFabClient: pf,
		ClientData: login.ClientData{
			Nonce: session.Nonce(),
		},
	}.Dial("nethernet", address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.DoSpawn(); err != nil {
		panic(fmt.Sprintf("spawn: %s", err))
	}
}
