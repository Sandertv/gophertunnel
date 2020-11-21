package protocol

// ExperimentData holds data on an experiment that is either enabled or disabled.
type ExperimentData struct {
	// Name is the name of the experiment.
	Name string
	// Enabled specifies if the experiment is enabled. Vanilla typically always sets this to true for any
	// experiments sent.
	Enabled bool
}

// Experiment reads an ExperimentData x from IO r.
func Experiment(r IO, x *ExperimentData) {
	r.String(&x.Name)
	r.Bool(&x.Enabled)
}
