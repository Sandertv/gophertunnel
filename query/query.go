package query

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"
)

// Do queries a server at the address passed using the UT3 query protocol. If the server responds, a map
// containing information is returned.
// Note that some servers do not support querying, in which case the query will time out. Do will take at
// most five seconds to try and get the query information.
func Do(address string) (information map[string]string, err error) {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return nil, fmt.Errorf("dial udp: %w", err)
	}
	// We set a deadline of five seconds: There is no point waiting even longer if the query isn't finished
	// by then.
	if err := conn.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
		return nil, err
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))

	b := new(bytes.Buffer)
	(&request{
		RequestType:    queryTypeHandshake,
		SequenceNumber: r.Int31(),
	}).Marshal(b)
	if _, err := conn.Write(b.Bytes()); err != nil {
		return nil, err
	}
	b.Reset()

	// This data buffer is way too big for the first response, but that is fine, as we re-use it for the next
	// response.
	data := make([]byte, math.MaxUint16)
	n, err := conn.Read(data)
	if err != nil {
		return nil, err
	}
	resp := &response{}
	if err := resp.Unmarshal(bytes.NewBuffer(data[:n])); err != nil {
		return nil, err
	}

	(&request{
		RequestType:    queryTypeInformation,
		SequenceNumber: r.Int31(),
		ResponseNumber: resp.ResponseNumber,
	}).Marshal(b)
	if _, err := conn.Write(b.Bytes()); err != nil {
		return nil, err
	}

	n, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	resp = &response{}
	return resp.Information, resp.Unmarshal(bytes.NewBuffer(data[:n]))
}
