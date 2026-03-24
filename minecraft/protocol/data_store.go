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

const (
	DataStorePropertyTypeNone   = 0
	DataStorePropertyTypeBool   = 1
	DataStorePropertyTypeInt64  = 2
	DataStorePropertyTypeString = 4
	DataStorePropertyTypeMap    = 6
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
	r.Uint32(&x.ChangeType)
	switch x.ChangeType {
	case DataStoreChangeTypeUpdate:
		Single(r, &x.Update)
	case DataStoreChangeTypeChange:
		r.String(&x.Change.DataStoreName)
		r.String(&x.Change.Property)
		r.Uint32(&x.Change.UpdateCount)
		MarshalDataStorePropertyValue(r, &x.Change.NewValue)
	case DataStoreChangeTypeRemoval:
		r.String(&x.Removal.DataStoreName)
	default:
		r.UnknownEnumOption(x.ChangeType, "data store change type")
	}
}

// DataStoreUpdate represents an update to a data store property.
type DataStoreUpdate struct {
	// DataStoreName is the name of the data store.
	DataStoreName string
	// Property is the property being updated.
	Property string
	// Path is the path within the property.
	Path string
	// ControlType is the type of the data value. It is one of the DataStoreControl constants.
	ControlType uint32
	// DoubleValue is the value if ControlType is DataStoreControlDouble.
	DoubleValue float64
	// BoolValue is the value if ControlType is DataStoreControlBoolean.
	BoolValue bool
	// StringValue is the value if ControlType is DataStoreControlString.
	StringValue string
	// PropertyUpdateCount is the update count for the property.
	PropertyUpdateCount uint32
	// PathUpdateCount is the update count for the path.
	PathUpdateCount uint32
}

// Marshal encodes/decodes a DataStoreUpdate.
func (x *DataStoreUpdate) Marshal(r IO) {
	r.String(&x.DataStoreName)
	r.String(&x.Property)
	r.String(&x.Path)
	r.Uint32(&x.ControlType)
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
	r.Uint32(&x.PropertyUpdateCount)
	r.Uint32(&x.PathUpdateCount)
}

// DataStorePropertyValue represents a typed property value in a data store.
type DataStorePropertyValue struct {
	// Type is the property value type. It is one of the DataStorePropertyType constants.
	Type int32
	// BoolValue is the value if Type is DataStorePropertyTypeBool.
	BoolValue bool
	// Int64Value is the value if Type is DataStorePropertyTypeInt64.
	Int64Value int64
	// StringValue is the value if Type is DataStorePropertyTypeString.
	StringValue string
	// MapValue is the value if Type is DataStorePropertyTypeMap.
	MapValue []DataStoreMapEntry
}

// DataStoreMapEntry represents a key-value pair in a data store map property.
type DataStoreMapEntry struct {
	// Key is the map key.
	Key string
	// Value is the map value.
	Value DataStorePropertyValue
}

// MarshalDataStorePropertyValue encodes/decodes a DataStorePropertyValue.
func MarshalDataStorePropertyValue(r IO, x *DataStorePropertyValue) {
	r.Int32(&x.Type)
	switch x.Type {
	case DataStorePropertyTypeNone:
		// No data.
	case DataStorePropertyTypeBool:
		r.Bool(&x.BoolValue)
	case DataStorePropertyTypeInt64:
		r.Int64(&x.Int64Value)
	case DataStorePropertyTypeString:
		r.String(&x.StringValue)
	case DataStorePropertyTypeMap:
		FuncSlice(r, &x.MapValue, func(entry *DataStoreMapEntry) {
			r.String(&entry.Key)
			MarshalDataStorePropertyValue(r, &entry.Value)
		})
	default:
		r.UnknownEnumOption(x.Type, "data store property type")
	}
}

// DataStoreChange represents a change to a data store property value.
type DataStoreChange struct {
	// DataStoreName is the name of the data store.
	DataStoreName string
	// Property is the property that changed.
	Property string
	// UpdateCount is the update count.
	UpdateCount uint32
	// NewValue is the new property value.
	NewValue DataStorePropertyValue
}

// DataStoreRemoval represents a removal from a data store.
type DataStoreRemoval struct {
	// DataStoreName is the name of the data store being removed.
	DataStoreName string
}
