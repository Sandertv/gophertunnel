# Getting Started
Gophertunnel is written in Golang to work with Gophertunnel you must have Golang compiler installed. Information on the golang [compiler is avilible here](https://golang.org/)

In addition there are several learning tools you can go through to learn go. 

[Go by Example]( https://gobyexample.com/) - recommended by Skippy

[Getting Started with go.dev]( https://learn.go.dev/) - recommended by Doad

[A Tour of Go](https://tour.golang.org/welcome/1) - recommended by Skippy

[Effective Go]( https://golang.org/doc/effective_go.html) - recommended by Skippy

[Gopher Reading List](https://github.com/enocom/gopher-reading-list) - recommended by Strum355

### Gophercises

[coding exercises for budding gophers](https://gophercises.com/) - recommended by Skippy

[TutorialEdge Golang Courses](https://tutorialedge.net/course/golang/) - recommended by Skippy

These may not be required for everyone, however they can help you understand the language if you are new.
## Example 1: MITM proxy
If you want an example of gophertunnel being used a simple proxy server is included in the root of the git repository in the main.go file. To get this to work clone the repository, or download the zip. Then execute the following steps:
* Edit the config.toml file
  * LocalAddress is the port and address your client will connect to in game
  * RemoteAddress is the server you want to play on
* enable loop back (assuming the proxy server is running on the same machine as your proxy)
  * run as admin: `CheckNetIsolation LoopbackExempt -a -n="Microsoft.MinecraftUWP_8wekyb3d8bbwe"`
* start the server
  * `go run .`
* authenticate by copying the provided link into a web browser then copying the oauth key and logging in.

You proxy server is up and running. you are now able to edit the code as you wish

### Orientation to the example
The Main function handles authentication then enters an infinite loop waiting for a connection. upon recieiving a connection it launches the connection handler function.

The connection handler function sets up the server dialer to connect to the remote server, and the listener connection. two concurent loops handle each server.

This loops hands packets coming from the client going to the server
```golang
go func() {
	defer listener.Disconnect(conn, "connection lost")
	defer serverConn.Close()
	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			return
		}
		if err := serverConn.WritePacket(pk); err != nil {
			if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
				_ = listener.Disconnect(conn, disconnect.Error())
			}
			return
		}
	}
}()
```
This loops hands packets coming from the server going to the client
```golang
go func() {
	defer serverConn.Close()
	defer listener.Disconnect(conn, "connection lost")
	for {
		pk, err := serverConn.ReadPacket()
		if err != nil {
			if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
				_ = listener.Disconnect(conn, disconnect.Error())
			}
			return
		}
		if err := conn.WritePacket(pk); err != nil {
			return
		}
	}
}()
```
To add in packet sniffing the supported paket types can befoud in the [minecraft/protocol/packet](https://github.com/Sandertv/gophertunnel/tree/master/minecraft/protocol/packet) folder. An example on how to use these type can be found in [example_listener_test.go] (https://github.com/Sandertv/gophertunnel/blob/master/minecraft/example_listener_test.go) but simply build an expression such as the below switch case in the respective loop you want to sniff.
```golang
switch p := pk.(type) {
	case *packet.Emote:
		fmt.Printf("Emote packet received: %v\n", p.EmoteID)
	case *packet.MovePlayer:
		fmt.Printf("Player %v moved to %v\n", p.EntityRuntimeID, p.Position)
}
```
