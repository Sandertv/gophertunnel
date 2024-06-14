package minecraft_test

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func ExampleListen() {
	// Create a minecraft.Listener with a specific name to be displayed as MOTD in the server list.
	name := "MOTD of this server"
	cfg := minecraft.ListenConfig{
		StatusProvider: minecraft.NewStatusProvider(name, "Gophertunnel"),
	}

	// Listen on the address with port 19132.
	address := ":19132"
	listener, err := cfg.Listen("raknet", address)
	if err != nil {
		panic(err)
	}

	for {
		// Accept connections in a for loop. Accept will only return an error if the minecraft.Listener is
		// closed. (So never unexpectedly.)
		c, err := listener.Accept()
		if err != nil {
			return
		}
		conn := c.(*minecraft.Conn)

		go func() {
			// Process the connection on another goroutine as you would with TCP connections.
			defer conn.Close()

			// Make the client spawn in the world using conn.StartGame. An error is returned if the client
			// times out during the connection.
			worldData := minecraft.GameData{ /* World data here */ }
			if err := conn.StartGame(worldData); err != nil {
				return
			}

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
				p := &packet.ChunkRadiusUpdated{ChunkRadius: 32}
				if err := conn.WritePacket(p); err != nil {
					break
				}
			}
		}()
	}
}
