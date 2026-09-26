package minecraft

import (
	"testing"
	"time"
)

func TestSetReadDeadlineClearsZeroWithLocation(t *testing.T) {
	conn := &Conn{}
	if err := conn.SetReadDeadline(time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	zero := time.Time{}.In(time.FixedZone("test", 3600))
	if err := conn.SetReadDeadline(zero); err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.readDeadline:
		t.Fatal("cleared deadline is ready")
	default:
	}
}
