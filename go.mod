module github.com/sandertv/gophertunnel

go 1.22

toolchain go1.22.1

require (
	github.com/go-gl/mathgl v1.1.0
	github.com/go-jose/go-jose/v3 v3.0.3
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.6.0
	github.com/klauspost/compress v1.17.9
	github.com/muhammadmuzzammil1998/jsonc v1.0.0
	github.com/pelletier/go-toml v1.9.5
	github.com/sandertv/go-raknet v1.14.1
	golang.org/x/net v0.26.0
	golang.org/x/oauth2 v0.21.0
	golang.org/x/text v0.16.0
)

require (
	github.com/df-mc/atomic v1.10.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20220909182711-5c715a9e8561 // indirect
	golang.org/x/image v0.17.0 // indirect
)

replace github.com/sandertv/go-raknet => github.com/tedacmc/tedac-raknet v0.0.4
