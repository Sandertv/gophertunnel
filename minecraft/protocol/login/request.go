package login

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// chain holds a chain with claims, each with their own headers, payloads and signatures. Each claim holds
// a public key used to verify other claims.
type chain []string

type certificate struct {
	Chain chain `json:"chain"`
}

// request is the outer encapsulation of the request. It holds a chain and a ClientData object.
type request struct {
	// Certificate holds the client certificate chain. The chain holds several claims that the server may verify in order to
	// make sure that the client is logged into XBOX Live.
	Certificate certificate `json:"Certificate"`
	// AuthenticationType is the authentication type of the request.
	AuthenticationType uint8 `json:"AuthenticationType"`
	// Token is an empty string, it's unclear what's used for.
	Token string `json:"Token"`
	// RawToken holds the raw token that follows the JWT chain, holding the ClientData.
	RawToken string `json:"-"`
	// Legacy specifies whether to use the legacy format of the request or not.
	Legacy bool `json:"-"`
}

func (r *request) MarshalJSON() ([]byte, error) {
	if r.Legacy {
		return json.Marshal(r.Certificate)
	}

	cert, err := json.Marshal(r.Certificate)
	if err != nil {
		return nil, err
	}

	type Alias request
	return json.Marshal(&struct {
		Certificate string `json:"Certificate"`
		Alias
	}{
		Certificate: string(cert),
		Alias:       (Alias)(*r),
	})
}

func init() {
	//noinspection SpellCheckingInspection
	const mojangPublicKey = `MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAECRXueJeTDqNRRgJi/vlRufByu/2G0i2Ebt6YMar5QX/R0DIIyrJMcUpruK4QveTfJSTp3Shlq4Gk34cD/4GUWwkv0DVuzeuB+tXija7HBxii03NHDbPAD0AKnLr2wdAp`

	data, _ := base64.StdEncoding.DecodeString(mojangPublicKey)
	publicKey, _ := x509.ParsePKIXPublicKey(data)
	mojangKey = publicKey.(*ecdsa.PublicKey)
}

// mojangKey holds the parsed Mojang ecdsa.PublicKey.
var mojangKey = new(ecdsa.PublicKey)

// AuthResult is returned by a call to Parse. It holds the ecdsa.PublicKey of the client and a bool that
// indicates if the player was logged in with XBOX Live.
type AuthResult struct {
	PublicKey             *ecdsa.PublicKey
	XBOXLiveAuthenticated bool
}

// Parse parses and verifies the login request passed. The AuthResult returned holds the ecdsa.PublicKey that
// was parsed (which is used for encryption) and a bool specifying if the request was authenticated by XBOX
// Live.
// Parse returns IdentityData and ClientData, of which IdentityData cannot under any circumstance be edited by
// the client. Rather, it is obtained from an authentication endpoint. The ClientData can, however, be edited
// freely by the client.
func Parse(request []byte) (IdentityData, ClientData, AuthResult, error) {
	var (
		iData IdentityData
		cData ClientData
		res   AuthResult
		key   = &ecdsa.PublicKey{}
	)
	req, err := parseLoginRequest(request)
	if err != nil {
		return iData, cData, res, fmt.Errorf("parse login request: %w", err)
	}
	tok, err := jwt.ParseSigned(req.Certificate.Chain[0], []jose.SignatureAlgorithm{jose.ES384})
	if err != nil {
		return iData, cData, res, fmt.Errorf("parse token 0: %w", err)
	}

	// The first token holds the client's public key in the x5u (it's self signed).
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	raw, _ := tok.Headers[0].ExtraHeaders["x5u"]
	if err := parseAsKey(raw, key); err != nil {
		return iData, cData, res, fmt.Errorf("parse x5u: %w", err)
	}

	var identityClaims identityClaims
	var authenticated bool
	t, iss := time.Now(), "Mojang"

	switch len(req.Certificate.Chain) {
	case 1:
		// Player was not authenticated with XBOX Live, meaning the one token in here is self-signed.
		if err := parseFullClaim(req.Certificate.Chain[0], key, &identityClaims); err != nil {
			return iData, cData, res, err
		}
		if err := identityClaims.Validate(jwt.Expected{Time: t}); err != nil {
			return iData, cData, res, fmt.Errorf("validate token 0: %w", err)
		}
	case 3:
		// Player was (or should be) authenticated with XBOX Live, meaning the chain is exactly 3 tokens
		// long.
		var c jwt.Claims
		if err := parseFullClaim(req.Certificate.Chain[0], key, &c); err != nil {
			return iData, cData, res, fmt.Errorf("parse token 0: %w", err)
		}
		if err := c.Validate(jwt.Expected{Time: t}); err != nil {
			return iData, cData, res, fmt.Errorf("validate token 0: %w", err)
		}
		authenticated = bytes.Equal(key.X.Bytes(), mojangKey.X.Bytes()) && bytes.Equal(key.Y.Bytes(), mojangKey.Y.Bytes())

		if err := parseFullClaim(req.Certificate.Chain[1], key, &c); err != nil {
			return iData, cData, res, fmt.Errorf("parse token 1: %w", err)
		}
		if err := c.Validate(jwt.Expected{Time: t, Issuer: iss}); err != nil {
			return iData, cData, res, fmt.Errorf("validate token 1: %w", err)
		}
		if err := parseFullClaim(req.Certificate.Chain[2], key, &identityClaims); err != nil {
			return iData, cData, res, fmt.Errorf("parse token 2: %w", err)
		}
		if err := identityClaims.Validate(jwt.Expected{Time: t, Issuer: iss}); err != nil {
			return iData, cData, res, fmt.Errorf("validate token 2: %w", err)
		}
		if authenticated != (identityClaims.ExtraData.XUID != "") {
			return iData, cData, res, fmt.Errorf("identity data must have an XUID when logged into XBOX Live only")
		}
	default:
		return iData, cData, res, fmt.Errorf("unexpected login chain length %v", len(req.Certificate.Chain))
	}
	if err := parseFullClaim(req.RawToken, key, &cData); err != nil {
		return iData, cData, res, fmt.Errorf("parse client data: %w", err)
	}
	if strings.Count(cData.ServerAddress, ":") > 1 && cData.ServerAddress[0] != '[' {
		// IPv6: We can't net.ResolveUDPAddr this directly, because Mojang does
		// not always put [] around the IP if it isn't added by the player in
		// the External Server adding screen. We'll have to do this manually:
		ind := strings.LastIndex(cData.ServerAddress, ":")
		cData.ServerAddress = "[" + cData.ServerAddress[:ind] + "]" + cData.ServerAddress[ind:]
	}
	if err := cData.Validate(); err != nil {
		return iData, cData, res, fmt.Errorf("validate client data: %w", err)
	}
	return identityClaims.ExtraData, cData, AuthResult{PublicKey: key, XBOXLiveAuthenticated: authenticated}, nil
}

// parseLoginRequest parses the structure of a login request from the data passed and returns it.
func parseLoginRequest(requestData []byte) (*request, error) {
	buf := bytes.NewBuffer(requestData)
	chain, err := decodeChain(buf)
	if err != nil {
		return nil, err
	}
	if len(chain) < 1 {
		return nil, fmt.Errorf("JWT chain must be at least 1 token long")
	}
	var rawLength int32
	if err := binary.Read(buf, binary.LittleEndian, &rawLength); err != nil {
		return nil, fmt.Errorf("read raw token length: %w", err)
	}
	return &request{
		Certificate: certificate{Chain: chain},
		RawToken:    string(buf.Next(int(rawLength))),
	}, nil
}

// parseFullClaim parses and verifies a full claim using the ecdsa.PublicKey passed. The key passed is updated
// if the claim holds an identityPublicKey field.
// The value v passed is decoded into when reading the claims.
func parseFullClaim(claim string, key *ecdsa.PublicKey, v any) error {
	tok, err := jwt.ParseSigned(claim, []jose.SignatureAlgorithm{jose.ES384})
	if err != nil {
		return fmt.Errorf("error parsing signed token: %w", err)
	}
	var m map[string]any
	if err := tok.Claims(key, v, &m); err != nil {
		return fmt.Errorf("error verifying claims of token: %w", err)
	}
	newKey, present := m["identityPublicKey"]
	if present {
		if err := parseAsKey(newKey, key); err != nil {
			return fmt.Errorf("error parsing identity public key: %w", err)
		}
	}
	return nil
}

// parseAsKey parses the base64 encoded ecdsa.PublicKey held in k as a public key and sets it to the variable
// pub passed.
func parseAsKey(k any, pub *ecdsa.PublicKey) error {
	kStr, _ := k.(string)
	if err := ParsePublicKey(kStr, pub); err != nil {
		return fmt.Errorf("error parsing public key: %w", err)
	}
	return nil
}

// Encode encodes a login request using the encoded login chain passed and the client data. The request's
// client data token is signed using the private key passed. It must be the same as the one used to get the
// login chain.
func Encode(loginChain string, data ClientData, key *ecdsa.PrivateKey, legacy bool) []byte {
	// We first decode the login chain we actually got in a new certificate.
	cert := &certificate{}
	_ = json.Unmarshal([]byte(loginChain), &cert)

	// We parse the header of the first claim it has in the chain, which will soon be the second claim.
	keyData := MarshalPublicKey(&key.PublicKey)
	tok, _ := jwt.ParseSigned(cert.Chain[0], []jose.SignatureAlgorithm{jose.ES384})

	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	x5uData, _ := tok.Headers[0].ExtraHeaders["x5u"]
	x5u, _ := x5uData.(string)
	claims := jwt.Claims{
		Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour * 6)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour * 6)),
	}

	signer, _ := jose.NewSigner(jose.SigningKey{Key: key, Algorithm: jose.ES384}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]any{"x5u": keyData},
	})
	firstJWT, _ := jwt.Signed(signer).Claims(identityPublicKeyClaims{
		Claims:               claims,
		IdentityPublicKey:    x5u,
		CertificateAuthority: true,
	}).Serialize()

	req := &request{
		Certificate: certificate{
			// We add our own claim at the start of the chain.
			Chain: append(chain{firstJWT}, cert.Chain...),
		},
		Legacy: legacy,
	}
	// We create another token this time, which is signed the same as the claim we just inserted in the chain,
	// just now it contains client data.
	req.RawToken, _ = jwt.Signed(signer).Claims(data).Serialize()

	return encodeRequest(req)
}

// encodeRequest encodes the request passed to a byte slice which is suitable for setting to the Connection
// Request field in a Login packet.
func encodeRequest(req *request) []byte {
	chainBytes, _ := json.Marshal(req)

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, int32(len(chainBytes)))
	_, _ = buf.WriteString(string(chainBytes))

	_ = binary.Write(buf, binary.LittleEndian, int32(len(req.RawToken)))
	_, _ = buf.WriteString(req.RawToken)
	return buf.Bytes()
}

// EncodeOffline creates a login request using the identity data and client data passed. The private key
// passed will be used to self sign the JWTs.
// Unlike Encode, EncodeOffline does not have a token signed by the Mojang key. It consists of only one JWT
// which holds the identity data of the player.
func EncodeOffline(identityData IdentityData, data ClientData, key *ecdsa.PrivateKey, legacy bool) []byte {
	keyData := MarshalPublicKey(&key.PublicKey)
	claims := jwt.Claims{
		Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour * 6)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour * 6)),
	}

	signer, _ := jose.NewSigner(jose.SigningKey{Key: key, Algorithm: jose.ES384}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]any{"x5u": keyData},
	})
	firstJWT, _ := jwt.Signed(signer).Claims(identityClaims{
		Claims:            claims,
		ExtraData:         identityData,
		IdentityPublicKey: keyData,
	}).Serialize()

	req := &request{
		Certificate: certificate{
			Chain: chain{firstJWT},
		},
		AuthenticationType: 2,
		Legacy:             legacy,
	}
	// We create another token this time, which is signed the same as the claim we just inserted in the chain,
	// just now it contains client data.
	req.RawToken, _ = jwt.Signed(signer).Claims(data).Serialize()

	return encodeRequest(req)
}

// decodeChain reads a certificate chain from the buffer passed and returns each claim found in the chain.
func decodeChain(buf *bytes.Buffer) (chain, error) {
	var chainLength int32
	if err := binary.Read(buf, binary.LittleEndian, &chainLength); err != nil {
		return nil, fmt.Errorf("read chain length: %w", err)
	}
	if chainLength <= 0 {
		return nil, fmt.Errorf("invalid chain length: %d", chainLength)
	}
	chainData := buf.Next(int(chainLength))

	req := struct {
		AuthenticationType uint8  `json:"AuthenticationType"`
		Certificate        string `json:"Certificate"`
		Chain              chain  `json:"chain"`
	}{}
	if err := json.Unmarshal(chainData, &req); err != nil {
		return nil, fmt.Errorf("decode chain JSON: %w", err)
	}

	var ch chain
	if req.Certificate != "" {
		cert := &certificate{}
		_ = json.Unmarshal([]byte(req.Certificate), &cert)
		ch = cert.Chain
	} else {
		ch = req.Chain
	}

	// First check if the chain actually has any elements in it.
	if len(ch) == 0 {
		return nil, fmt.Errorf("decode chain: no elements")
	}

	// Then check if the authentication type is guest mode.
	if req.AuthenticationType == 1 {
		return nil, fmt.Errorf("guest authentication is not supported")
	}
	return ch, nil
}

// identityClaims holds the claims for the last token in the chain, which contains the IdentityData of the
// player.
type identityClaims struct {
	jwt.Claims

	// ExtraData holds the extra data of this claim, which is the IdentityData of the player.
	ExtraData IdentityData `json:"extraData"`

	IdentityPublicKey string `json:"identityPublicKey"`
}

// Validate validates the identity claims held by the struct and returns an error if any illegal data was
// encountered.
func (c identityClaims) Validate(e jwt.Expected) error {
	if err := c.Claims.Validate(e); err != nil {
		return err
	}
	return c.ExtraData.Validate()
}

// identityPublicKeyClaims holds the claims for a JWT that holds an identity public key.
type identityPublicKeyClaims struct {
	jwt.Claims

	// IdentityPublicKey holds a serialised ecdsa.PublicKey used in the next JWT in the chain.
	IdentityPublicKey    string `json:"identityPublicKey"`
	CertificateAuthority bool   `json:"certificateAuthority,omitempty"`
}

// ParsePublicKey parses an ecdsa.PublicKey from the base64 encoded public key data passed and sets it to a
// pointer. If parsing failed or if the public key was not of the type ECDSA, an error is returned.
func ParsePublicKey(b64Data string, key *ecdsa.PublicKey) error {
	data, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return fmt.Errorf("decode public key data: %w", err)
	}
	publicKey, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}
	ecdsaKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected ECDSA public key, got %v", key)
	}
	*key = *ecdsaKey
	return nil
}

// MarshalPublicKey marshals an ecdsa.PublicKey to a base64 encoded binary representation.
func MarshalPublicKey(key *ecdsa.PublicKey) string {
	data, _ := x509.MarshalPKIXPublicKey(key)
	return base64.StdEncoding.EncodeToString(data)
}
