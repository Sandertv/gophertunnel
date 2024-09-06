module github.com/sandertv/gophertunnel

go 1.23.0

require (
	github.com/df-mc/go-nethernet v0.0.0-20240902102242-528de5c8686f
	github.com/df-mc/go-playfab v0.0.0-20240902102459-2f8b5cd02173
	github.com/df-mc/go-xsapi v0.0.0-20240902102602-e7c4bffb955f
	github.com/go-gl/mathgl v1.1.0
	github.com/go-jose/go-jose/v3 v3.0.3
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.6.0
	github.com/klauspost/compress v1.17.9
	github.com/muhammadmuzzammil1998/jsonc v1.0.0
	github.com/pelletier/go-toml v1.9.5
	github.com/sandertv/go-raknet v1.14.1
	golang.org/x/net v0.27.0
	golang.org/x/oauth2 v0.21.0
	golang.org/x/text v0.17.0
	nhooyr.io/websocket v1.8.11
)

require (
	github.com/andreburgaud/crypt2go v1.8.0 // indirect
	github.com/coder/websocket v1.8.12 // indirect
	github.com/pion/datachannel v1.5.9 // indirect
	github.com/pion/dtls/v3 v3.0.2 // indirect
	github.com/pion/ice/v4 v4.0.1 // indirect
	github.com/pion/interceptor v0.1.30 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.14 // indirect
	github.com/pion/rtp v1.8.9 // indirect
	github.com/pion/sctp v1.8.33 // indirect
	github.com/pion/sdp/v3 v3.0.9 // indirect
	github.com/pion/srtp/v3 v3.0.3 // indirect
	github.com/pion/stun/v3 v3.0.0 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v4 v4.0.0 // indirect
	github.com/pion/webrtc/v4 v4.0.0-beta.29.0.20240826201411-3147b45f9db5 // indirect
	github.com/wlynxg/anet v0.0.3 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/image v0.17.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
)

replace (
	github.com/df-mc/go-nethernet => github.com/lactyy/go-nethernet v0.0.0-20240902104417-681fd9263f4a
	github.com/df-mc/go-playfab => github.com/lactyy/go-playfab v0.0.0-20240906070923-01f9987eafb6
	github.com/df-mc/go-xsapi => github.com/lactyy/go-xsapi v0.0.0-20240902120723-5a844e61607e
	github.com/pion/sctp => github.com/lactyy/sctp v0.0.0-20240822210319-2eae0bcbc9f3
)
