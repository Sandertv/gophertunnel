module github.com/sandertv/gophertunnel

go 1.24.1

require (
	github.com/go-gl/mathgl v1.2.0
	github.com/go-jose/go-jose/v4 v4.1.1
	github.com/golang/snappy v1.0.0
	github.com/google/uuid v1.6.0
	github.com/klauspost/compress v1.18.0
	github.com/muhammadmuzzammil1998/jsonc v1.0.0
	github.com/pelletier/go-toml v1.9.5
	github.com/sandertv/go-raknet v1.14.2
	golang.org/x/net v0.41.0
	golang.org/x/oauth2 v0.30.0
	golang.org/x/text v0.26.0
)

require golang.org/x/crypto v0.39.0 // indirect

replace github.com/sandertv/go-raknet => github.com/tedacmc/tedac-raknet v0.0.6
