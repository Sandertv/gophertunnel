package main

import (
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
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
		var g sync.WaitGroup
		g.Add(2)
		go func() {
			if err := conn.StartGame(serverConn.GameData()); err != nil {
				panic(err)
			}
			g.Done()
		}()
		go func() {
			if err := serverConn.DoSpawn(); err != nil {
				panic(err)
			}
			g.Done()
		}()
		g.Wait()

		go func() {
			defer listener.Disconnect(conn, "connection lost")
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
			defer listener.Disconnect(conn, "connection lost")
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
