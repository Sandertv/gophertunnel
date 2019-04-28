package login

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login/jwt"
	"reflect"
)

// Chain holds a chain with claims, each with their own headers, payloads and signatures. Each claim holds
// a public key used to verify other claims.
type Chain []string

// request is the outer encapsulation of the request. It holds a Chain and a ClientData object.
type request struct {
	// Chain is the client certificate chain. It holds several claims that the server may verify in order to
	// make sure that the client is logged into XBOX Live.
	Chain Chain `json:"chain"`
}

func init() {
	// By default only allow the ES384 algorithm, as that is the only one that Minecraft will ever use.
	jwt.AllowAlg("ES384")
}

// Verify verifies the login request string passed. It ensures the claims found in the certificate chain are
// signed correctly and it looks for the Mojang public key to find out if the player was authenticated.
func Verify(requestString string) (publicKey *ecdsa.PublicKey, authenticated bool, err error) {
	chain, err := chain(bytes.NewBuffer([]byte(requestString)))
	if err != nil {
		return nil, false, err
	}
	pubKey := &ecdsa.PublicKey{}
	for _, claim := range chain {
		// Verify each of the claims found in the chain using the empty public key above, which will be set
		// after verifying the first public key.
		if hasKey, err := jwt.Verify(claim, pubKey); err != nil {
			return nil, false, fmt.Errorf("error verifying claim: %v", err)
		} else {
			if hasKey == true {
				// If the claim we just verified had the Mojang public key in it, we set the authenticated
				// bool to true.
				authenticated = true
			}
		}
	}
	return pubKey, authenticated, nil
}

// Decode decodes the login request string passed into an IdentityData struct, which contains trusted identity
// data such as the UUID of the player, and ClientData, which contains user specific data such as the skin of
// a player.
// Decode does not verify the request passed. For that reason, login.Verify() should be called on that same
// string before login.Decode().
func Decode(requestString string) (IdentityData, ClientData, error) {
	identityData, clientData := IdentityData{}, ClientData{}
	buf := bytes.NewBuffer([]byte(requestString))
	chain, err := chain(buf)
	if err != nil {
		return identityData, clientData, err
	}
	for _, claim := range chain {
		container := &identityDataContainer{}
		payload, err := jwt.Payload(claim)
		if err != nil {
			return identityData, clientData, fmt.Errorf("error parsing payload from claim: %v", err)
		}
		if err := json.Unmarshal(payload, &container); err != nil {
			return identityData, clientData, fmt.Errorf("error JSON decoding claim payload: %v", err)
		}
		// If the extra data decoded is not equal to the identity data (in other words, not empty), we set the
		// data and break out of the loop.
		if container.ExtraData != identityData {
			identityData = container.ExtraData
			break
		}
	}

	// Just like the certificate chain, the length of the raw token is also prefixed with an int, so we decode
	// that first.
	var rawLength int32
	if err := binary.Read(buf, binary.LittleEndian, &rawLength); err != nil {
		return identityData, clientData, fmt.Errorf("error reading raw token length: %v", err)
	}
	rawToken := buf.Next(int(rawLength))

	// We take the payload directly out of the raw token, as the header and signature aren't relevant here.
	payload, err := jwt.Payload(string(rawToken))
	if err != nil {
		return identityData, clientData, fmt.Errorf("error reading payload from raw token: %v", err)
	}
	// Finally we decode the data in the client data.
	if err := json.Unmarshal(payload, &clientData); err != nil {
		return identityData, clientData, fmt.Errorf("error decoding raw token payload JSON: %v", err)
	}

	// We JSON encode our ClientData struct again and check it against the original data to see if there is
	// any data we missed.
	if !equalJSON(payload, clientData) {
		data, _ := json.Marshal(clientData)
		return identityData, clientData, fmt.Errorf("original raw token payload is not equal to the parsed data: \n	payload: %v\n	decoded: %v", string(payload), string(data))
	}

	return identityData, clientData, nil
}

// identityDataContainer is used to decode identity data found in a JWT claim into an IdentityData struct.
type identityDataContainer struct {
	ExtraData IdentityData `json:"extraData"`
}

// chain reads a certificate chain from the buffer passed and returns each claim found in the chain.
func chain(buf *bytes.Buffer) (Chain, error) {
	var chainLength int32
	if err := binary.Read(buf, binary.LittleEndian, &chainLength); err != nil {
		return nil, fmt.Errorf("error reading chain length: %v", err)
	}
	chainData := buf.Next(int(chainLength))

	request := &request{}
	if err := json.Unmarshal(chainData, &request); err != nil {
		return nil, fmt.Errorf("error decoding request chain JSON: %v", err)
	}
	// First check if the chain actually has any elements in it.
	if len(request.Chain) == 0 {
		return nil, fmt.Errorf("connection request had no claims in the chain")
	}
	return request.Chain, nil
}

// equalJSON checks if the raw JSON passed and the JSON encoded representation of the decoded value passed are
// considered equal.
func equalJSON(original []byte, decoded interface{}) bool {
	originalData := map[string]interface{}{}
	_ = json.Unmarshal(original, &originalData)
	encoded, _ := json.Marshal(decoded)
	decodedData := map[string]interface{}{}
	_ = json.Unmarshal(encoded, &decodedData)
	return reflect.DeepEqual(originalData, decodedData)
}