package main

import (
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/script"
	"github.com/yuin/gluamapper"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"log"
	"os"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
// It has scripting functionality, in the sense that Lua code may be written in the stdin to send packets,
// either to the server or to the client.
// The Lua code may be either:
// single.line(), or
// "
// multi()
// line()
// "
// Currently, two functions are supported, one for sending packets to the client, and one for sending packets
// to the server:
// client.send("PacketName", {PacketField = "Something"})
// server.send("PacketName", {PacketField = true})
// The packets that may be sent are all packets that may be found in the protocol/packet package, with the
// name being the exact name of the packet struct, for example 'Text'. The fields must be fields found in the
// packet.
func main() {
	config := readConfig()

	listener, err := minecraft.Listen("raknet", config.Connection.LocalAddress)
	if err != nil {
		panic(err)
	}
	if err := listener.HijackPong(config.Connection.RemoteAddress); err != nil {
		panic(err)
	}
	defer listener.Close()
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		conn := c.(*minecraft.Conn)

		data := conn.ClientData()
		data.ServerAddress = config.Connection.RemoteAddress
		serverConn, err := minecraft.Dialer{
			Email:      config.Credentials.Email,
			Password:   config.Credentials.Password,
			ClientData: data,
		}.Dial("raknet", config.Connection.RemoteAddress)
		if err != nil {
			panic(err)
		}
		s := script.New()
		s.SetModule(script.NewModule("client").
			Func("send", func(L *lua.LState) int {
				str := L.CheckString(1)
				table := L.CheckTable(2)
				pk := packet.PacketsByName[str]()
				if err := gluamapper.Map(table, pk); err != nil {
					panic(err)
				}
				if err := conn.WritePacket(pk); err != nil {
					panic(err)
				}
				return 0
			}),
		)
		s.SetModule(script.NewModule("server").
			Func("send", func(L *lua.LState) int {
				str := L.CheckString(1)
				table := L.CheckTable(2)
				pk := packet.PacketsByName[str]()
				if err := gluamapper.Map(table, pk); err != nil {
					panic(err)
				}
				if err := serverConn.WritePacket(pk); err != nil {
					panic(err)
				}
				return 0
			}),
		)
		s.RunStdin()

		go func() {
			defer s.Close()
			defer conn.Close()
			defer serverConn.Close()
			for {
				pk, err := conn.ReadPacket()
				if err != nil {
					return
				}
				if err := serverConn.WritePacket(pk); err != nil {
					return
				}
			}
		}()
		go func() {
			defer serverConn.Close()
			defer conn.Close()
			for {
				pk, err := serverConn.ReadPacket()
				if err != nil {
					return
				}
				if err := conn.WritePacket(pk); err != nil {
					return
				}
			}
		}()
	}
}

type config struct {
	Connection struct {
		LocalAddress  string
		RemoteAddress string
	}
	Credentials struct {
		Email    string
		Password string
	}
}

func readConfig() config {
	c := config{}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, err := os.Create("config.toml")
		if err != nil {
			log.Fatalf("error creating config: %v", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("error encoding default config: %v", err)
		}
		if _, err := f.Write(data); err != nil {
			log.Fatalf("error writing encoded default config: %v", err)
		}
		_ = f.Close()
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	if c.Connection.LocalAddress == "" {
		c.Connection.LocalAddress = "0.0.0.0:19132"
	}
	data, _ = toml.Marshal(c)
	_ = ioutil.WriteFile("config.toml", data, 0644)
	return c
}
