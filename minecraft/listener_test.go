package minecraft

import (
	"fmt"
	"github.com/sandertv/go-raknet"
	"testing"
)

func TestListen(t *testing.T) {
	listener, err := Listen("raknet", "0.0.0.0:19132")
	if err != nil {
		panic(err)
	}
	if err := listener.listener.(*raknet.Listener).HijackPong("mco.mineplex.com:19132"); err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		fmt.Println(conn.RemoteAddr(), "connected.")

		go func() {
			defer func() {
				_ = conn.Close()
				fmt.Println(conn.RemoteAddr(), "disconnected.")
			}()
			for {
				if _, err := conn.(*Conn).ReadPacket(); err != nil {
					return
				}
			}
		}()
	}
}
