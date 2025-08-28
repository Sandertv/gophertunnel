package auth

type Device struct {
	ClientID   string
	DeviceType string
	Version    string
}

var DeviceAndroid = &Device{ClientID: "0000000048183522", DeviceType: "Android", Version: "10"}
