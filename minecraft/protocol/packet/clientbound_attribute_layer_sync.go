package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundAttributeLayerSync is sent by the server to synchronise attribute layers with the client.
type ClientBoundAttributeLayerSync struct {
	// PayloadType is the type of attribute layer payload. It is one of the protocol.AttributeLayerPayloadType constants.
	PayloadType uint32
	// Layers is set if PayloadType is AttributeLayerPayloadTypeUpdateLayers.
	Layers []protocol.AttributeLayerData
	// LayerName is the attribute layer name, used for UpdateSettings, UpdateEnvironment and RemoveEnvironment.
	LayerName string
	// DimensionID is the dimension ID, used for UpdateSettings, UpdateEnvironment and RemoveEnvironment.
	DimensionID int32
	// Settings is set if PayloadType is AttributeLayerPayloadTypeUpdateSettings.
	Settings protocol.AttributeLayerSettings
	// EnvironmentAttributes is set if PayloadType is AttributeLayerPayloadTypeUpdateEnvironment.
	EnvironmentAttributes []protocol.EnvironmentAttributeData
	// RemoveAttributeNames is set if PayloadType is AttributeLayerPayloadTypeRemoveEnvironment.
	RemoveAttributeNames []string
}

// ID ...
func (*ClientBoundAttributeLayerSync) ID() uint32 {
	return IDClientBoundAttributeLayerSync
}

func (pk *ClientBoundAttributeLayerSync) Marshal(io protocol.IO) {
	io.Varuint32(&pk.PayloadType)
	switch pk.PayloadType {
	case protocol.AttributeLayerPayloadTypeUpdateLayers:
		protocol.Slice(io, &pk.Layers)
	case protocol.AttributeLayerPayloadTypeUpdateSettings:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.Single(io, &pk.Settings)
	case protocol.AttributeLayerPayloadTypeUpdateEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.Slice(io, &pk.EnvironmentAttributes)
	case protocol.AttributeLayerPayloadTypeRemoveEnvironment:
		io.String(&pk.LayerName)
		io.Varint32(&pk.DimensionID)
		protocol.FuncSlice(io, &pk.RemoveAttributeNames, io.String)
	default:
		io.UnknownEnumOption(pk.PayloadType, "attribute layer payload type")
	}
}
