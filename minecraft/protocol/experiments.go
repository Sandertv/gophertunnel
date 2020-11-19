package protocol

type ExperimentData struct {
	Name    string
	Enabled bool
}

func Experiments(r IO, x *[]ExperimentData) {
	var count int32
	r.Int32(&count)
	*x = make([]ExperimentData, count)
	for i := int32(0); i < count; i++ {
		e := ExperimentData{}
		r.String(&e.Name)
		r.Bool(&e.Enabled)
		*x = append(*x, e)
	}
}
