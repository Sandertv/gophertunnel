package room

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Dialer struct {
	Log *slog.Logger
}

func (d Dialer) DialContext(ctx context.Context, a ConnAnnouncer, n net.Conn, ref Reference) (*Conn, error) {
	if err := a.Join(ctx, ref); err != nil {
		return nil, fmt.Errorf("join: %w", err)
	}

	return &Conn{
		d: d,

		announcer: a,
		conn:      n,

		closed: make(chan struct{}),
	}, nil
}

type Conn struct {
	d Dialer

	announcer ConnAnnouncer
	conn      net.Conn

	closed chan struct{}
	once   sync.Once
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		close(c.closed)

		errs := []error{c.conn.Close()}
		if err := c.announcer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close announcer: %w", err))
		}
		err = errors.Join(errs...)
	})
	return err
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
