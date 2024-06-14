package query

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// version is the version of the query protocol. It represents Gamespy Query Protocol version 4.
var version = [2]byte{0xfe, 0xfd}

// padding is the padding used for the queryTypeInformation request.
var padding = [4]byte{0xff, 0xff, 0xff, 0x01}

// splitNum is the split number set in a response before the Information is written. It is conventionally
// written as a string 'splitnum' terminated by a null byte, so we do this too.
var splitNum = [9]byte{'S', 'P', 'L', 'I', 'T', 'N', 'U', 'M', 0x00}

// playerKey is the key under which players are typically stored.
var playerKey = [...]byte{0x00, 0x01, 'p', 'l', 'a', 'y', 'e', 'r', '_', 0x00, 0x00}

const (
	queryTypeHandshake   = 0x09
	queryTypeInformation = 0x00
)

// request is a packet sent by the client to the server. It is first used to request the handshake, and after
// that to request the information.
type request struct {
	// RequestType is the type of the request. It is either queryTypeHandshake or queryTypeInformation,
	// with queryTypeHandshake being sent first and queryTypeInformation being sent in response to the
	// queryTypeHandshake.
	RequestType byte
	// SequenceNumber is a sequence number identifying the request. Typically, this is a timestamp, but it is
	// merely used to match request with response, so the actual value it holds isn't relevant.
	SequenceNumber int32
	// ResponseNumber is the number sent in the response following a handshake request. It only requires being
	// set if RequestType is queryTypeInformation.
	ResponseNumber int32
}

// response is a packet sent by the server to the client. It is sent in response to a request, with either a
// response indicating the handshake was successful or the actual information of the query.
type response struct {
	// ResponseType is the RequestType of the request that the packet is a response to. It is either
	// queryTypeHandshake, which holds simply a number for the next request, or queryTypeInformation, which
	// holds the information of the server.
	ResponseType byte
	// SequenceNumber is the SequenceNumber sent in the request packet. Typically, this is a timestamp, but it
	// is merely used to match request with response, so the actual value it holds isn't relevant.
	SequenceNumber int32
	// ResponseNumber is a number sent only if ResponseType is queryTypeHandshake. The request packet holds
	// this number in the next request.
	ResponseNumber int32
	// Information is a list of all information of the server. It is sent only if ResponseType is
	// queryTypeInformation.
	Information map[string]string
}

// Marshal ...
func (pk *request) Marshal(w io.Writer) {
	_, _ = w.Write(version[:])
	_ = binary.Write(w, binary.BigEndian, pk.RequestType)
	_ = binary.Write(w, binary.BigEndian, pk.SequenceNumber)
	if pk.RequestType == queryTypeInformation {
		_ = binary.Write(w, binary.BigEndian, pk.ResponseNumber)
		_, _ = w.Write(padding[:])
	}
}

// Unmarshal ...
func (pk *request) Unmarshal(r io.Reader) error {
	v := make([]byte, 2)
	if _, err := r.Read(v); err != nil {
		return err
	}
	if !bytes.Equal(v, version[:]) {
		return fmt.Errorf("invalid query request version: expected %X, got %X", version, v)
	}
	if err := binary.Read(r, binary.BigEndian, &pk.RequestType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &pk.RequestType); err != nil {
		return err
	}
	if pk.RequestType == queryTypeInformation {
		if err := binary.Read(r, binary.BigEndian, &pk.ResponseNumber); err != nil {
			return err
		}
		p := make([]byte, 4)
		_, err := r.Read(p)
		return err
	} else if pk.RequestType != queryTypeHandshake {
		return fmt.Errorf("unknown request type %X", pk.RequestType)
	}
	return nil
}

// Marshal ...
func (pk *response) Marshal(w io.Writer) {
	_ = binary.Write(w, binary.BigEndian, pk.ResponseType)
	_ = binary.Write(w, binary.BigEndian, pk.SequenceNumber)
	if pk.ResponseType == queryTypeHandshake {
		v := []byte(fmt.Sprint(pk.ResponseNumber))
		if len(v) != 12 {
			// Pad the response number to 12 bytes.
			v = append(v, make([]byte, 12-len(v))...)
		}
		_, _ = w.Write(v)
	} else {
		_, _ = w.Write(splitNum[:])
		_ = binary.Write(w, binary.BigEndian, byte(0x80)) // Number of packets, but in our case always 0x80.
		_ = binary.Write(w, binary.BigEndian, byte(0))    // Unused.
		values := make([][]byte, 0, len(pk.Information)*2)
		for key, value := range pk.Information {
			values = append(values, []byte(key))
			values = append(values, []byte(value))
		}
		// Join all keys and values together using a null byte.
		_, _ = w.Write(bytes.Join(values, []byte{0x00}))
	}
}

// Unmarshal ...
func (pk *response) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &pk.ResponseType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &pk.SequenceNumber); err != nil {
		return err
	}
	switch pk.ResponseType {
	case queryTypeHandshake:
		numBytes := make([]byte, 12)
		if _, err := r.Read(numBytes); err != nil {
			return err
		}
		index := bytes.Index(numBytes, []byte{0x00})
		if index != -1 {
			numBytes = numBytes[:index]
		}
		num, err := strconv.ParseInt(string(numBytes), 10, 32)
		if err != nil {
			return fmt.Errorf("invalid response number in handshake query response: %w", err)
		}
		pk.ResponseNumber = int32(num)
	case queryTypeInformation:
		// We don't care about these first 11 bytes, so we skip them all and move on to the actual information.
		v := make([]byte, 11)
		if _, err := r.Read(v); err != nil {
			return err
		}
		information, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		playerIndex := bytes.Index(information, playerKey[:])
		data := information
		if playerIndex != -1 {
			information = information[:playerIndex]
		}
		values := bytes.Split(information, []byte{0x00})
		pk.Information = make(map[string]string, len(values)/2)
		if len(values)%2 != 0 {
			// Sometimes, the information is null terminated, whereas with others it is not. We remove the
			// null byte if it's there.
			values = values[:len(values)-1]
		}
		for i := 0; i < len(values); i += 2 {
			pk.Information[string(values[i])] = string(values[i+1])
		}
		if playerIndex != -1 {
			playerData := data[playerIndex+len(playerKey):]
			values = bytes.Split(playerData, []byte{0x00})
			players := make([]string, 0, len(values))
			for i := 0; i < len(values); i++ {
				if len(values[i]) == 0 {
					// Empty string means we've reached the end of the data. Break immediately so that the
					// name isn't added to the players slice.
					break
				}
				players = append(players, string(values[i]))
			}
			pk.Information["players"] = strings.Join(players, ", ")
		}

	default:
		return fmt.Errorf("unknown response type %X", pk.ResponseType)
	}
	return nil
}
