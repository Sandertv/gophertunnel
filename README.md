# gophertunnel
A Minecraft library containing packages to create clients, servers, proxies and other tools, and a proxy implementation using them.

[Module Documentation](https://pkg.go.dev/mod/github.com/sandertv/gophertunnel)

![telescope gopher](https://github.com/Sandertv/gophertunnel/blob/master/gophertunnel_telescope_coloured.png)

## Overview
gophertunnel is composed of several packages that may be of use for creating Minecraft related tools. A brief
overview of all packages may be found [here](https://pkg.go.dev/mod/github.com/sandertv/gophertunnel?tab=packages).

## Examples
Creating a Minecraft client that authenticates using an XBOX Live account and connects to a server:
```go
package main

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func main() {
	// Connect to the server.
	conn, err := minecraft.Dialer{
		TokenSource: auth.TokenSource,
	}.Dial("raknet", "mco.mineplex.com:19132")
	if err != nil {
		panic(err)
	}
	// Make the client spawn in the world.
	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}
	defer conn.Close()
	for {
		// Example: Read a packet from the connection.
		pk, err := conn.ReadPacket()
		if err != nil {
			break
		}

		// Example: Send a packet to the server in response to the previous packet.
		if err := conn.WritePacket(&packet.RequestChunkRadius{ChunkRadius: 32}); err != nil {
			break
		}
	}
}
```

Creating a Minecraft listener that can accept incoming clients and adapts the MOTD from another server:
```go
package main

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func main() {
	listener, err := minecraft.Listen("raknet", "0.0.0.0:19132")
	if err != nil {
		panic(err)
	}

	for {
		c, err := listener.Accept()
		if err != nil {
			return
		}
		conn := c.(*minecraft.Conn)

		go func() {
			// Process the connection on another goroutine as you would with TCP connections.
			defer conn.Close()

			// Make the client connecting spawn.
			if err := conn.StartGame(minecraft.GameData{ /* World data here */ }); err != nil {
				panic(err)
			}

			for {
				// Example: Read a packet from the client.
				if _, err := conn.ReadPacket(); err != nil {
					return
				}

				// Example: Send a packet to the client in response to the previous packet.
				if err := conn.WritePacket(&packet.ChunkRadiusUpdated{ChunkRadius: 32}); err != nil {
					break
				}
			}
		}()
	}
}
```

## Versions
Gophertunnel supports only one version at a time. Generally, a new minor version is tagged when gophertunnel
supports a new Minecraft version that was not previously supported. A list of the recommended gophertunnel
versions for past Minecraft versions is listed below.

| Version | Tag      |
|---------|----------|
| 1.16.20 | Latest   |
| 1.16.0  | v1.7.11  |
| 1.14.60 | v1.6.5   |
| 1.14.0  | v1.3.20  |
| 1.13.0  | v1.3.5   |
| 1.12.0  | v1.2.11  |

## Proxy
A MITM proxy program is implemented in the main.go file. It uses the gophertunnel libraries to create a proxy
that provides user authentication and proxying a connection to another server.

## Contact
[![Chat on Discord](https://img.shields.io/badge/Chat-On%20Discord-738BD7.svg?style=for-the-badge)](https://discord.gg/evzQR4R)