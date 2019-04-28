package minecraft

import (
	"github.com/sandertv/go-raknet"
	"log"
	"testing"
	"time"
)

func TestListen(t *testing.T) {
	listener, err := Listen("raknet", "0.0.0.0:19132")
	if err != nil {
		panic(err)
	}
	listener.ErrorLog = log.New()
	if err := listener.listener.(*raknet.Listener).HijackPong("mco.mineplex.com:19132"); err != nil {
		panic(err)
	}
	_ = listener
	time.Sleep(time.Hour)
}
