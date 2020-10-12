package minecraft_test

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func ExampleDial() {
	// Create a minecraft.Dialer with an auth.TokenSource to authenticate to the server.
	dialer := minecraft.Dialer{
		TokenSource: auth.TokenSource,
	}
	// Dial a new connection to the target server.
	address := "mco.mineplex.com:19132"
	conn, err := dialer.Dial("raknet", address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Make the client spawn in the world: This is a blocking operation that will return an error if the
	// client times out while spawning.
	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}

	// You will then want to start a for loop that reads packets from the connection until it is closed.
	for {
		// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
		// a read timeout is set. You will generally want to return or break if this happens.
		pk, err := conn.ReadPacket()
		if err != nil {
			break
		}

		// The pk variable is of type packet.Packet, which may be type asserted to gain access to the data
		// they hold:
		switch p := pk.(type) {
		case *packet.Emote:
			fmt.Printf("Emote packet received: %v\n", p.EmoteID)
		case *packet.MovePlayer:
			fmt.Printf("Player %v moved to %v\n", p.EntityRuntimeID, p.Position)
		}

		// Write a packet to the connection: Similarly to ReadPacket, WritePacket will (only) return an error
		// if the connection is closed.
		p := &packet.RequestChunkRadius{ChunkRadius: 32}
		if err := conn.WritePacket(p); err != nil {
			break
		}
	}
}
