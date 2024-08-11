package rta

import (
	"encoding/json"
	"strconv"
	"strings"
)

type handshake struct {
	sequence uint32
	status   int32
	payload  []json.RawMessage
}

const (
	typeSubscribe uint32 = iota + 1
	typeUnsubscribe
	typeEvent
	typeResync
)

const (
	operationSubscribe uint8 = iota
	operationUnsubscribe
	operationCapacity // The capacity of expected handshake uses.
)

func typeToOperation(typ uint32) uint8 {
	switch typ {
	case typeSubscribe:
		return operationSubscribe
	case typeUnsubscribe:
		return operationUnsubscribe
	default:
		panic("unreachable")
	}
}

func operationToType(op uint8) uint32 {
	switch op {
	case operationSubscribe:
		return typeSubscribe
	case operationUnsubscribe:
		return typeUnsubscribe
	default:
		panic("unreachable")
	}
}

func (c *Conn) shake(op uint8, sequence uint32, payload []any) (<-chan *handshake, error) {
	if err := c.write(operationToType(op), append([]any{sequence}, payload...)); err != nil {
		return nil, err
	}
	hand := make(chan *handshake)
	c.expectedMu.Lock()
	c.expected[op][sequence] = hand
	c.expectedMu.Unlock()
	return hand, nil
}

func (c *Conn) release(op uint8, sequence uint32) {
	c.expectedMu.Lock()
	delete(c.expected[op], sequence)
	c.expectedMu.Unlock()
}

type UnexpectedStatusError struct {
	Code    int32
	Message string
}

func (e *UnexpectedStatusError) Error() string {
	b := &strings.Builder{}
	b.WriteString("rta: code ")
	b.WriteString(strconv.FormatInt(int64(e.Code), 10))
	if e.Message != "" {
		b.WriteByte(':')
		b.WriteByte(' ')
		b.WriteString(e.Message)
	}
	return b.String()
}

const (
	StatusOK int32 = iota
	StatusUnknownResource
	StatusSubscriptionLimitReached
	StatusNoResourceData
	StatusThrottled          = 1001
	StatusServiceUnavailable = 1002
)
