package opt

import "github.com/sandertv/gophertunnel/minecraft/protocol/login"

// Credentials returns an option containing credentials to login to XBOX Live. The login passed must be an
// email address, or otherwise valid value to log into a Live account.
func Credentials(login, password string) Opt {
	return Opt{name: "credentials", value: Creds{Login: login, Password: password}}
}

// ClientData returns an option containing specific client data sent during the initial login of the client.
// If not set, the data is autofilled with default values.
func ClientData(data login.ClientData) Opt {
	return Opt{name: "client_data", value: data}
}

// Creds is a struct containing credentials to login to XBOX Live.
type Creds struct {
	// Login is the email address used to login to the Live account.
	Login    string
	Password string
}

// Opt is an option that may be passed to a minecraft.Dial() call. These options provide specific information
// which may be used to alter behaviour of the client.
type Opt struct {
	name  string
	value interface{}
}

// Map maps a slice of options to a map[string]interface{}. The map is indexed by the name of the options.
func Map(opts []Opt) map[string]interface{} {
	options := make(map[string]interface{})
	for _, opt := range opts {
		options[opt.name] = opt.value
	}
	return options
}
