package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/oauth2"
)

// TokenSource holds an oauth2.TokenSource which uses device auth to get a code. The user authenticates using
// a code. TokenSource prints the authentication code and URL to os.Stdout. To use a different io.Writer, use
// WriterTokenSource. TokenSource automatically refreshes tokens.
var TokenSource = AndroidConfig.WriterTokenSource(os.Stdout)

// WriterTokenSource returns a new oauth2.TokenSource which, like TokenSource, uses device auth to get a code.
// Unlike TokenSource, WriterTokenSource allows passing an io.Writer to which information on the auth URL and
// code are printed. WriterTokenSource automatically refreshes tokens.
func WriterTokenSource(w io.Writer) oauth2.TokenSource {
	return AndroidConfig.WriterTokenSource(w)
}

func (conf Config) WriterTokenSource(w io.Writer) oauth2.TokenSource {
	return &tokenSource{w: w, conf: conf}
}

// tokenSource implements the oauth2.TokenSource interface. It provides a method to get an oauth2.Token using
// device auth through a call to RequestLiveToken.
//
// NOTE: tokenSource requires a Config field to be set, otherwise the device auth
// flow will send an invalid request and fail. Prefer constructing via [Config.WriterTokenSource] (or
// [AndroidConfig.WriterTokenSource]) rather than instantiating tokenSource directly.
type tokenSource struct {
	w    io.Writer
	t    oauth2.TokenSource
	mu   sync.Mutex
	conf Config
}

// Token attempts to return a Live Connect token using the RequestLiveToken function.
func (src *tokenSource) Token() (*oauth2.Token, error) {
	if src.conf.ClientID == "" {
		panic(fmt.Errorf("minecraft/auth: tokenSource: missing ClientID; construct via Config.WriterTokenSource (or AndroidConfig.WriterTokenSource)"))
	}

	src.mu.Lock()
	defer src.mu.Unlock()
	if src.t == nil {
		t, err := src.conf.RequestLiveTokenWriter(src.w)
		if err != nil {
			return nil, err
		}
		src.t = src.conf.TokenSource(context.Background(), t)
	}
	tok, err := src.t.Token()
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// RefreshTokenSource returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires. Note that this function must be used over oauth2.ReuseTokenSource
// due to that function not refreshing with the correct scopes.
func RefreshTokenSource(t *oauth2.Token) oauth2.TokenSource {
	return RefreshTokenSourceWriter(t, os.Stdout)
}

// RefreshTokenSource returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires. Note that this function must be used over oauth2.ReuseTokenSource
// due to that function not refreshing with the correct scopes.
func (conf Config) RefreshTokenSource(t *oauth2.Token) oauth2.TokenSource {
	return conf.RefreshTokenSourceWriter(t, os.Stdout)
}

// RefreshTokenSourceWriter returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires. It requests from io.Writer if the oauth2.Token is invalid.
// Note that this function must be used over oauth2.ReuseTokenSource due to that
// function not refreshing with the correct scopes.
func RefreshTokenSourceWriter(t *oauth2.Token, w io.Writer) oauth2.TokenSource {
	return AndroidConfig.RefreshTokenSourceWriter(t, w)
}

// RefreshTokenSourceWriter returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires. It requests from io.Writer if the oauth2.Token is invalid.
// Note that this function must be used over oauth2.ReuseTokenSource due to that
// function not refreshing with the correct scopes.
func (conf Config) RefreshTokenSourceWriter(t *oauth2.Token, w io.Writer) oauth2.TokenSource {
	var src oauth2.TokenSource
	if t != nil {
		src = conf.TokenSource(context.Background(), t)
	}
	return &tokenSource{w: w, conf: conf, t: src}
}

// RequestLiveToken does a login request for Microsoft Live Connect using device auth. A login URL will be
// printed to the stdout with a user code which the user must use to submit.
// RequestLiveToken is the equivalent of RequestLiveTokenWriter(os.Stdout).
func RequestLiveToken() (*oauth2.Token, error) {
	return RequestLiveTokenWriter(os.Stdout)
}

// RequestLiveToken does a login request for Microsoft Live Connect using device auth. A login URL will be
// printed to the stdout with a user code which the user must use to submit.
// RequestLiveToken is the equivalent of RequestLiveTokenWriter(os.Stdout).
func (conf Config) RequestLiveToken() (*oauth2.Token, error) {
	return conf.RequestLiveTokenWriter(os.Stdout)
}

// RequestLiveTokenWriter does a login request for Microsoft Live Connect using device auth. A login URL will
// be printed to the io.Writer passed with a user code which the user must use to submit.
// Once fully authenticated, an oauth2 token is returned which may be used to login to XBOX Live.
func RequestLiveTokenWriter(w io.Writer) (*oauth2.Token, error) {
	return AndroidConfig.RequestLiveTokenWriter(w)
}

// RequestLiveTokenWriter does a login request for Microsoft Live Connect using device auth. A login URL will
// be printed to the io.Writer passed with a user code which the user must use to submit.
// Once fully authenticated, an oauth2 token is returned which may be used to login to XBOX Live.
func (conf Config) RequestLiveTokenWriter(w io.Writer) (*oauth2.Token, error) {
	return conf.RequestLiveTokenContext(context.Background(), w)
}

func RequestLiveTokenContext(ctx context.Context, w io.Writer) (*oauth2.Token, error) {
	return AndroidConfig.RequestLiveTokenContext(ctx, w)
}

func (conf Config) RequestLiveTokenContext(ctx context.Context, w io.Writer) (*oauth2.Token, error) {
	d, err := conf.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("start device auth: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Authenticate at %v using the code %v.\n", d.VerificationURI, d.UserCode)

	token, err := conf.DeviceAccessToken(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("poll device token: %w", err)
	}
	_, _ = w.Write([]byte("Authentication successful.\n"))
	return token, nil
}
