package nethernet

type Credentials struct {
	ExpirationInSeconds int         `json:"ExpirationInSeconds"`
	ICEServers          []ICEServer `json:"TurnAuthServers"`
}

type ICEServer struct {
	Username string   `json:"Username"`
	Password string   `json:"Password"`
	URLs     []string `json:"Urls"`
}
