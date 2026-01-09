package login

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
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
// The verifier will be used for parsing the OpenID token included in the first chain of the login request.
func Parse(request []byte, verifier *oidc.IDTokenVerifier) (IdentityData, ClientData, AuthResult, error) {
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

	var (
		authenticated bool
		t             = time.Now()
	)
	if verifier != nil && req.Token != "" {
		// The context here is used for making requests via remote key set, which does not normally
		// occur in this case since we use a custom-made OIDC verifier that has already static key set included.
		idt, err := verifier.Verify(context.Background(), req.Token)
		if err != nil {
			return iData, cData, res, fmt.Errorf("verify ID token: %w", err)
		}
		var claims tokenClaims
		if err := idt.Claims(&claims); err != nil {
			return iData, cData, res, fmt.Errorf("parse ID token: %w", err)
		}
		if err := claims.Validate(jwt.Expected{Time: t}); err != nil {
			return iData, cData, res, fmt.Errorf("validate ID token: %w", err)
		}
		if err := ParsePublicKey(claims.ClientPublicKey, key); err != nil {
			return iData, cData, res, fmt.Errorf("parse cpk: %w", err)
		}
		iData, authenticated = claims.identityData(), iData.XUID != ""
		// The OIDC token does not include the numerical XBL title ID (extraData.titleId) that is present in the
		// legacy Mojang chain. If the chain is present and valid, we verify it and use its title ID so callers
		// can keep using IdentityData.TitleID.
		if iData.XUID != "" {
			if legacyID, _, legacyAuthed, err := parseLegacyChain(req.Certificate.Chain, t); err == nil && legacyAuthed {
				if legacyID.TitleID != "" && legacyID.XUID == iData.XUID {
					iData.TitleID = legacyID.TitleID
				}
			}
		}
		if err := iData.Validate(); err != nil {
			return iData, cData, res, fmt.Errorf("validate identity data: %w", err)
		}
	} else {
		legacyID, legacyKey, legacyAuthed, err := parseLegacyChain(req.Certificate.Chain, t)
		if err != nil {
			return iData, cData, res, err
		}
		iData, key, authenticated = legacyID, legacyKey, legacyAuthed
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
	return iData, cData, AuthResult{PublicKey: key, XBOXLiveAuthenticated: authenticated}, nil
}

// parseLegacyChain verifies the legacy Mojang chain and returns IdentityData from extraData,
// the public key used for verification (and for client data), and a bool indicating if the chain was
// authenticated by Xbox Live.
func parseLegacyChain(chain []string, now time.Time) (IdentityData, *ecdsa.PublicKey, bool, error) {
	key := &ecdsa.PublicKey{}
	tok, err := jwt.ParseSigned(chain[0], []jose.SignatureAlgorithm{jose.ES384})
	if err != nil {
		return IdentityData{}, nil, false, fmt.Errorf("parse token 0: %w", err)
	}

	// The first token holds the client's public key in the x5u (it's self signed).
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	raw, _ := tok.Headers[0].ExtraHeaders["x5u"]
	if err := parseAsKey(raw, key); err != nil {
		return IdentityData{}, nil, false, fmt.Errorf("parse x5u: %w", err)
	}

	var (
		identityClaims identityClaims
		authenticated  bool
	)
	iss := "Mojang"

	switch len(chain) {
	case 1:
		// Player was not authenticated with XBOX Live, meaning the one token in here is self-signed.
		if err := parseFullClaim(chain[0], key, &identityClaims); err != nil {
			return IdentityData{}, nil, false, err
		}
		if err := identityClaims.Validate(jwt.Expected{Time: now}); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("validate token 0: %w", err)
		}
	case 3:
		// Player was (or should be) authenticated with XBOX Live, meaning the chain is exactly 3 tokens long.
		var c jwt.Claims
		if err := parseFullClaim(chain[0], key, &c); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("parse token 0: %w", err)
		}
		if err := c.Validate(jwt.Expected{Time: now}); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("validate token 0: %w", err)
		}
		authenticated = bytes.Equal(key.X.Bytes(), mojangKey.X.Bytes()) && bytes.Equal(key.Y.Bytes(), mojangKey.Y.Bytes())

		if err := parseFullClaim(chain[1], key, &c); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("parse token 1: %w", err)
		}
		if err := c.Validate(jwt.Expected{Time: now, Issuer: iss}); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("validate token 1: %w", err)
		}
		if err := parseFullClaim(chain[2], key, &identityClaims); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("parse token 2: %w", err)
		}
		if err := identityClaims.Validate(jwt.Expected{Time: now, Issuer: iss}); err != nil {
			return IdentityData{}, nil, false, fmt.Errorf("validate token 2: %w", err)
		}
		if authenticated != (identityClaims.ExtraData.XUID != "") {
			return IdentityData{}, nil, false, fmt.Errorf("identity data must have an XUID when logged into XBOX Live only")
		}
	default:
		return IdentityData{}, nil, false, fmt.Errorf("unexpected login chain length %v", len(chain))
	}
	return identityClaims.ExtraData, key, authenticated, nil
}

// parseLoginRequest parses the structure of a login request from the data passed and returns it.
func parseLoginRequest(requestData []byte) (*request, error) {
	buf := bytes.NewBuffer(requestData)
	var chainLength int32
	if err := binary.Read(buf, binary.LittleEndian, &chainLength); err != nil {
		return nil, fmt.Errorf("read chain length: %w", err)
	}
	if chainLength <= 0 {
		return nil, fmt.Errorf("invalid chain length: %d", chainLength)
	}
	chainData := buf.Next(int(chainLength))

	r := struct {
		request
		Certificate string `json:"Certificate"`
		Chain       chain  `json:"chain"`
	}{}
	if err := json.Unmarshal(chainData, &r); err != nil {
		return nil, fmt.Errorf("decode chain data: %w", err)
	}

	if r.Certificate != "" {
		if err := json.Unmarshal([]byte(r.Certificate), &r.request.Certificate); err != nil {
			return nil, fmt.Errorf("decode certificate: %w", err)
		}
	} else {
		r.request.Certificate.Chain = r.Chain
	}

	// First check if the chain actually has any elements in it.
	if len(r.request.Certificate.Chain) == 0 {
		return nil, fmt.Errorf("decode chain: no elements")
	}

	// Then check if the authentication type is guest mode.
	if r.AuthenticationType == 1 {
		return nil, fmt.Errorf("guest authentication is not supported")
	}

	var rawLength int32
	if err := binary.Read(buf, binary.LittleEndian, &rawLength); err != nil {
		return nil, fmt.Errorf("read raw token length: %w", err)
	}
	r.request.RawToken = string(buf.Next(int(rawLength)))
	if n := buf.Len(); n != 0 {
		return nil, fmt.Errorf("%d unread bytes", n)
	}
	return &r.request, nil
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
// login chain. The multiplayer token is used as the Token field in the connection request.
func Encode(loginChain string, data ClientData, key *ecdsa.PrivateKey, token string, legacy bool) []byte {
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
		Token:  token,
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
// The token parameter is optional and can be an empty string for offline logins that don't require
// a multiplayer token.
func EncodeOffline(identityData IdentityData, data ClientData, key *ecdsa.PrivateKey, token string, legacy bool) []byte {
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
		Token:              token,
		Legacy:             legacy,
	}
	// We create another token this time, which is signed the same as the claim we just inserted in the chain,
	// just now it contains client data.
	req.RawToken, _ = jwt.Signed(signer).Claims(data).Serialize()

	return encodeRequest(req)
}

// tokenClaims holds the claims for the multiplayer token from the first chain,
// which contains the fields related to the identity of the player.
type tokenClaims struct {
	jwt.Claims

	// IdentityProviderType is seemingly the underlying identity provider
	// used to sign in to the authorization service. It is always 'PlayFab'.
	IdentityProviderType string `json:"ipt"`
	// PlayFabID is the PlayFab entity ID for the authenticated player.
	// It is the ID for the master player account of the player, which
	// is shared across multiple PlayFab titles published by Mojang.
	PlayFabID string `json:"mid"`
	// PlayFabTitleID is the title ID specific to PlayFab.
	// It is typically '20CA2' for the base version of the game.
	PlayFabTitleID string `json:"tid"`
	// ClientPublicKey is the public key of the client used to sign the client data
	// and to initialise the encryption in the handshake.
	ClientPublicKey string `json:"cpk"`
	// XUID is the ID of the authenticated player specific to Xbox Live.
	XUID string `json:"xid"`
	// DisplayName is the in-game name for the authenticated player.
	DisplayName string `json:"xname"`
}

// identityData converts the OIDC tokenClaims into IdentityData.
// Fields that exist in the legacy chain's extraData are filled to keep behavior consistent.
func (tc tokenClaims) identityData() IdentityData {
	return IdentityData{
		XUID:           tc.XUID,
		Identity:       identityFromXUID(tc.XUID).String(),
		DisplayName:    tc.DisplayName,
		PlayFabID:      tc.PlayFabID,
		PlayFabTitleID: tc.PlayFabTitleID,
	}
}

// identityFromXUID returns the UUID derived from the player's XUID claimed
// from the new multiplayer token.
func identityFromXUID(xuid string) uuid.UUID {
	// See [github.com/google/uuid.NewHash], This takes 'pocket-auth-1-uuid:' as
	// the name-space instead of UUID and uses the player's XUID to compute a v3 UUID.
	hash := md5.New()
	hash.Write([]byte("pocket-auth-1-xuid:"))
	hash.Write([]byte(xuid))
	s := hash.Sum(nil)
	var id uuid.UUID
	copy(id[:], s)
	id[6] = (id[6] & 0x0f) | 0x30 // Version 3
	id[8] = (id[8] & 0x3f) | 0x80 // RFC 4122 variant
	return id
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
