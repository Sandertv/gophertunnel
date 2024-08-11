package signaling

type Environment struct {
	ServiceURI string `json:"serviceUri"`
	StunURI    string `json:"stunUri"`
	TurnURI    string `json:"turnUri"`
}

func (e *Environment) EnvironmentName() string { return "signaling" }
