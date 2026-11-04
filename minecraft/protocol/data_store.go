package protocol

const (
	DataStoreChangeTypeUpdate = iota
	DataStoreChangeTypeChange
	DataStoreChangeTypeRemoval
)

const (
	DataStoreControlDouble = iota
	DataStoreControlBoolean
	DataStoreControlString
)

// DataStoreChangeEntry represents a single entry in the data store changes array.
type DataStoreChangeEntry struct {
	// ChangeType is the type of change. It is one of the DataStoreChangeType constants.
	ChangeType uint32
	// Update is set if ChangeType is DataStoreChangeTypeUpdate.
	Update DataStoreUpdate
	// Change is set if ChangeType is DataStoreChangeTypeChange.
	Change DataStoreChange
	// Removal is set if ChangeType is DataStoreChangeTypeRemoval.
	Removal DataStoreRemoval
}

// Marshal encodes/decodes a DataStoreChangeEntry.
func (x *DataStoreChangeEntry) Marshal(r IO) {
	r.Varuint32(&x.ChangeType)
	switch x.ChangeType {
	case DataStoreChangeTypeUpdate:
		Single(r, &x.Update)
	case DataStoreChangeTypeChange:
		Single(r, &x.Change)
	case DataStoreChangeTypeRemoval:
		r.String(&x.Removal.VariableName)
	default:
		r.UnknownEnumOption(x.ChangeType, "data store change type")
	}
}

// DataStoreValue is a primitive value of a data store variable.
type DataStoreValue struct {
	// ControlType is the type of the value. It is one of the DataStoreControl constants.
	ControlType uint32
	// DoubleValue is the value if ControlType is DataStoreControlDouble.
	DoubleValue float64
	// BoolValue is the value if ControlType is DataStoreControlBoolean.
	BoolValue bool
	// StringValue is the value if ControlType is DataStoreControlString.
	StringValue string
}

// Marshal encodes/decodes a DataStoreValue.
func (x *DataStoreValue) Marshal(r IO) {
	r.Varuint32(&x.ControlType)
	switch x.ControlType {
	case DataStoreControlDouble:
		r.Float64(&x.DoubleValue)
	case DataStoreControlBoolean:
		r.Bool(&x.BoolValue)
	case DataStoreControlString:
		r.String(&x.StringValue)
	default:
		r.UnknownEnumOption(x.ControlType, "data store control type")
	}
}

// DataStoreUpdate represents an update to a data store variable.
type DataStoreUpdate struct {
	// VariableName is the name of the variable being updated.
	VariableName string
	// Value is the new value of the variable.
	Value DataStoreValue
	// UpdateCount is the number of times the variable has been updated.
	UpdateCount uint32
}

// Marshal encodes/decodes a DataStoreUpdate.
func (x *DataStoreUpdate) Marshal(r IO) {
	r.String(&x.VariableName)
	Single(r, &x.Value)
	r.Uint32(&x.UpdateCount)
}

// DataStoreChange represents a change to the value of a data store variable.
type DataStoreChange struct {
	// VariableName is the name of the variable that changed.
	VariableName string
	// UpdateCount is the number of times the variable has been updated.
	UpdateCount uint32
	// Value is the new value of the variable.
	Value DataStoreValue
}

// Marshal encodes/decodes a DataStoreChange.
func (x *DataStoreChange) Marshal(r IO) {
	r.String(&x.VariableName)
	r.Uint32(&x.UpdateCount)
	Single(r, &x.Value)
}

// DataStoreRemoval represents the removal of a data store variable.
type DataStoreRemoval struct {
	// VariableName is the name of the variable being removed.
	VariableName string
}
