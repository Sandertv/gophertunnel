module github.com/sandertv/gophertunnel

go 1.24

require (
	github.com/df-mc/go-playfab v0.0.0-00010101000000-000000000000
	github.com/go-gl/mathgl v1.2.0
	github.com/go-jose/go-jose/v4 v4.1.0
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.6.0
	github.com/klauspost/compress v1.17.11
	github.com/muhammadmuzzammil1998/jsonc v1.0.0
	github.com/pelletier/go-toml v1.9.5
	github.com/sandertv/go-raknet v1.14.3-0.20250305181847-6af3e95113d6
	golang.org/x/net v0.35.0
	golang.org/x/oauth2 v0.25.0
	golang.org/x/text v0.22.0
)

require github.com/df-mc/go-xsapi v0.0.0-20240902102602-e7c4bffb955f // indirect

replace github.com/df-mc/go-xsapi => github.com/lactyy/go-xsapi v0.0.0-20240911052022-1b9dffef64ab

replace github.com/df-mc/go-playfab => github.com/lactyy/go-playfab v0.0.0-20240911042657-037f6afe426f
