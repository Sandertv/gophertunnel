package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	// GitHub API URL to list JSON files in the bedrock-protocol-docs repo
	githubAPIURL = "https://api.github.com/repos/Mojang/bedrock-protocol-docs/contents/json?ref=dfd586157b9cf5fece81bacbe29248cca578c951"
	// Output directory for generated packets
	outputDir = "../minecraft/protocol/packet/__generated__"
)

// GitHubFile represents a file entry from the GitHub API
type GitHubFile struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
}

// PacketJSON represents the JSON schema structure for a packet
type PacketJSON struct {
	Schema           string                    `json:"$schema"`
	ID               string                    `json:"$id"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	Type             string                    `json:"type"`
	Properties       map[string]PropertyJSON   `json:"properties"`
	Required         []string                  `json:"required"`
	MetaProperties   map[string]any            `json:"$metaProperties"`
	Definitions      map[string]DefinitionJSON `json:"definitions"`
	MinecraftVersion string                    `json:"x-minecraft-version"`
	ProtocolVersion  int                       `json:"x-protocol-version"`
}

// PropertyJSON represents a field/property in the packet
type PropertyJSON struct {
	Type                 any                       `json:"type"`
	Title                string                    `json:"title"`
	Description          string                    `json:"description"`
	Default              any                       `json:"default"`
	UnderlyingType       string                    `json:"x-underlying-type"`
	SerializationOptions []string                  `json:"x-serialization-options"`
	OrdinalIndex         int                       `json:"x-ordinal-index"`
	Minimum              float64                   `json:"minimum"`
	Maximum              float64                   `json:"maximum"`
	Enum                 []string                  `json:"enum"`
	Items                *ItemJSON                 `json:"items"`
	Ref                  string                    `json:"$ref"`
	OneOf                []OneOfJSON               `json:"oneOf"`
	ControlValueType     string                    `json:"x-control-value-type"` // For oneOf with primitive types
	AdditionalProperties *AdditionalPropertiesJSON `json:"additionalProperties"`
}

// AdditionalPropertiesJSON represents additionalProperties (map-like) schema
type AdditionalPropertiesJSON struct {
	Type        string                  `json:"type"`
	Description string                  `json:"description"`
	Properties  map[string]PropertyJSON `json:"properties"`
}

// ItemJSON represents array item schema
type ItemJSON struct {
	Type           string `json:"type"`
	Ref            string `json:"$ref"`
	UnderlyingType string `json:"x-underlying-type"`
}

// OneOfJSON represents oneOf variants
type OneOfJSON struct {
	Ref            string `json:"$ref"`
	Type           string `json:"type"`              // For primitive oneOf: "null", "boolean", "integer", "number"
	UnderlyingType string `json:"x-underlying-type"` // For primitive oneOf: "boolean", "int32", "float"
	OrdinalIndex   int    `json:"x-ordinal-index"`
}

// DefinitionJSON represents a definition/sub-type in the packet
type DefinitionJSON struct {
	Title      string                  `json:"title"`
	Type       string                  `json:"type"`
	Ref        string                  `json:"$ref"` // Top-level $ref - this definition is an alias for another
	Properties map[string]PropertyJSON `json:"properties"`
	Required   []string                `json:"required"`
}

// PacketInfo holds processed packet information
type PacketInfo struct {
	Name             string
	GoName           string
	FileName         string
	ID               int
	Description      string
	Fields           []FieldInfo
	Enums            []EnumInfo
	Structs          []StructInfo // Nested struct types to generate
	MinecraftVersion string
	ProtocolVersion  int
	NeedsVecImport   bool
	NeedsNBTImport   bool
	NeedsUUIDImport  bool
}

// FieldInfo holds processed field information
type FieldInfo struct {
	Name             string
	GoName           string
	GoType           string
	IOMethod         string
	OrdinalIndex     int
	Description      string
	IsSlice          bool
	SliceType        string
	IsComplex        bool // Needs special handling (oneOf variants, etc.)
	ComplexReason    string
	IsNBT            bool // Special handling for NBT serialization
	IsOptional       bool // Field is not in required array - use protocol.Optional[T]
	IsStructSlice    bool // Slice of structs - use protocol.Slice
	IsOneOf          bool // Field is a oneOf variant type
	IsPrimitiveOneOf bool // oneOf with primitive types (bool, int32, float, null)
	OneOfVariants    []OneOfVariant
	ControlField     string // Field name that controls which variant is used
	ControlValueType string // Type of the control value (e.g., "uint32")
}

// OneOfVariant represents a single variant in a oneOf field
type OneOfVariant struct {
	Index      int
	StructName string // For struct variants
	StructInfo *StructInfo
	GoType     string // For primitive variants (e.g., "bool", "int32", "float32")
	IOMethod   string // For primitive variants (e.g., "Bool", "Int32", "Float32")
	IsNull     bool   // For null variant
}

// StructInfo holds information about a nested struct type to generate
type StructInfo struct {
	Name        string      // Go struct name
	Description string      // Description comment
	Fields      []FieldInfo // Struct fields
}

// EnumInfo holds enum constant information
type EnumInfo struct {
	Name      string   // Prefix for constants (e.g., "ActorEvent")
	GoType    string   // Go type (e.g., "byte")
	Values    []string // Enum values in order
	FieldName string   // Field this enum is for
}

// TypeMapping maps JSON types to Go types and IO methods
var TypeMapping = map[string]struct {
	GoType   string
	IOMethod string
}{
	"uint8":   {"uint8", "Uint8"},
	"int8":    {"int8", "Int8"},
	"uint16":  {"uint16", "Uint16"},
	"int16":   {"int16", "Int16"},
	"uint32":  {"uint32", "Uint32"},
	"int32":   {"int32", "Int32"},
	"uint64":  {"uint64", "Uint64"},
	"int64":   {"int64", "Int64"},
	"float32": {"float32", "Float32"},
	"float":   {"float32", "Float32"},
	"float64": {"float64", "Float64"},
	"double":  {"float64", "Float64"},
	"boolean": {"bool", "Bool"},
	"bool":    {"bool", "Bool"},
	"string":  {"string", "String"},
}

// WellKnownTypes maps definition titles to Go types and IO methods
// These are composite types that have special handling in the protocol package
var WellKnownTypes = map[string]struct {
	GoType      string
	IOMethod    string
	NeedsImport string // Import path needed (empty = none, "mgl32" or "nbt")
}{
	"Vec3":                 {"mgl32.Vec3", "Vec3", "mgl32"},
	"Vec2":                 {"mgl32.Vec2", "Vec2", "mgl32"},
	"BlockPos":             {"protocol.BlockPos", "BlockPos", ""},
	"ChunkPos":             {"protocol.ChunkPos", "ChunkPos", ""},
	"SubChunkPos":          {"protocol.SubChunkPos", "SubChunkPos", ""},
	"ActorRuntimeID":       {"uint64", "Varuint64", ""},            // Single-property wrapper
	"ActorUniqueID":        {"int64", "Varint64", ""},              // Single-property wrapper
	"NetworkBlockPosition": {"protocol.BlockPos", "UBlockPos", ""}, // Network block position
	"Json::Value":          {"[]byte", "ByteSlice", ""},            // JSON data as byte slice
	"Reference":            {"string", "String", ""},               // Simple string reference
	"Identifier":           {"string", "String", ""},               // Namespaced identifier string
	"mce::UUID":            {"uuid.UUID", "UUID", "uuid"},          // Minecraft UUID
}

// SpecialDefinitionIDs maps numeric definition IDs to their Go types and IO methods
// These are opaque hashed IDs that Mojang uses for certain types
var SpecialDefinitionIDs = map[string]struct {
	GoType      string
	IOMethod    string // Special format: "NBT" means io.NBT(&pk.Field, nbt.NetworkLittleEndian)
	NeedsImport string
}{
	// NBT compound data - used for entity metadata, components, etc.
	"4158325036": {"map[string]any", "NBT", "nbt"},
}

// FieldNameMapping maps JSON field names to preferred Go field names
var FieldNameMapping = map[string]string{
	"Server Address":    "Address",
	"ServerAddress":     "Address",
	"Server Port":       "Port",
	"ServerPort":        "Port",
	"Actor Runtime ID":  "EntityRuntimeID",
	"Actor Unique ID":   "EntityUniqueID",
	"Target Runtime ID": "EntityRuntimeID",
	"Target Actor ID":   "EntityUniqueID",
	"Player Name":       "SourceName",
	"Sender's XUID":     "XUID",
	"Platform Id":       "PlatformChatID",
	"Event ID":          "EventType",
	"Motif":             "Title",
	"Id":                "TrackingID", // Avoid conflict with ID() method
	"ID":                "TrackingID", // Avoid conflict with ID() method
}

var generationTime = time.Now().UTC()

// Track globally generated enum constants and struct names to avoid duplicates
var generatedEnumConstants = make(map[string]bool)
var generatedStructNames = make(map[string]bool)

func main() {
	fmt.Println("🚀 Bedrock Protocol Packet Generator")
	fmt.Println("=====================================")

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Fetch list of JSON files from GitHub
	fmt.Println("\n📥 Fetching packet list from GitHub...")
	files, err := fetchGitHubFiles()
	if err != nil {
		fmt.Printf("❌ Failed to fetch file list: %v\n", err)
		os.Exit(1)
	}

	// Filter to only packet JSON files (not enums or protocoldoc)
	var packetFiles []GitHubFile
	for _, f := range files {
		if strings.HasSuffix(f.Name, "Packet.json") {
			packetFiles = append(packetFiles, f)
		}
	}

	fmt.Printf("📦 Found %d packet definitions\n", len(packetFiles))

	// Process each packet
	var packets []PacketInfo
	for i, f := range packetFiles {
		fmt.Printf("\r⏳ Processing packet %d/%d: %s", i+1, len(packetFiles), f.Name)

		packet, err := processPacketFile(f)
		if err != nil {
			fmt.Printf("\n⚠️  Warning: Failed to process %s: %v\n", f.Name, err)
			continue
		}

		if packet != nil {
			packets = append(packets, *packet)
		}
	}
	fmt.Println()

	// Sort packets by ID
	sort.Slice(packets, func(i, j int) bool {
		return packets[i].ID < packets[j].ID
	})

	// Generate packet files
	fmt.Println("\n📝 Generating Go files...")
	for _, p := range packets {
		if err := generatePacketFile(p); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate %s: %v\n", p.FileName, err)
		}
	}

	// Generate ID constants file
	if err := generateIDFile(packets); err != nil {
		fmt.Printf("❌ Failed to generate id.go: %v\n", err)
	}

	// Generate doc.go
	if err := generateDocFile(packets); err != nil {
		fmt.Printf("❌ Failed to generate doc.go: %v\n", err)
	}

	fmt.Printf("\n✅ Successfully generated %d packet files in %s\n", len(packets), outputDir)
}

func fetchGitHubFiles() ([]GitHubFile, error) {
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil
}

func processPacketFile(file GitHubFile) (*PacketInfo, error) {
	// Fetch the JSON content
	resp, err := http.Get(file.DownloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var packet PacketJSON
	if err := json.Unmarshal(body, &packet); err != nil {
		return nil, err
	}

	// Extract packet ID from meta properties
	packetID := 0
	if id, ok := packet.MetaProperties["[cereal:packet]"]; ok {
		switch v := id.(type) {
		case float64:
			packetID = int(v)
		case int:
			packetID = v
		}
	}

	if packetID == 0 {
		return nil, fmt.Errorf("no packet ID found")
	}

	// Convert packet name to Go style
	goName := strings.TrimSuffix(packet.Title, "Packet")
	goName = toGoName(goName)

	fileName := toSnakeCase(goName) + ".go"

	info := &PacketInfo{
		Name:             packet.Title,
		GoName:           goName,
		FileName:         fileName,
		ID:               packetID,
		Description:      packet.Description,
		MinecraftVersion: packet.MinecraftVersion,
		ProtocolVersion:  packet.ProtocolVersion,
	}

	// Build required fields set for quick lookup
	requiredSet := make(map[string]bool)
	for _, req := range packet.Required {
		requiredSet[req] = true
	}

	// Process fields
	var fields []FieldInfo
	var enums []EnumInfo
	for name, prop := range packet.Properties {
		isRequired := requiredSet[name]
		field, enum := processField(name, prop, packet.Definitions, info, isRequired)
		if field != nil {
			fields = append(fields, *field)
		}
		if enum != nil {
			enums = append(enums, *enum)
		}
	}

	// Sort fields by ordinal index
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].OrdinalIndex < fields[j].OrdinalIndex
	})

	info.Fields = fields
	info.Enums = enums
	return info, nil
}

func processField(name string, prop PropertyJSON, definitions map[string]DefinitionJSON, packetInfo *PacketInfo, isRequired bool) (*FieldInfo, *EnumInfo) {
	// Check for field name mapping
	goName := name
	if mapped, ok := FieldNameMapping[name]; ok {
		goName = mapped
	}
	goName = toGoName(goName)

	field := &FieldInfo{
		Name:         name,
		GoName:       goName,
		OrdinalIndex: prop.OrdinalIndex,
		Description:  prop.Description,
		IsOptional:   !isRequired,
	}

	var enumInfo *EnumInfo

	// Handle enums - check if this property has enum values
	if len(prop.Enum) > 0 && prop.Title != "" {
		// Extract enum name from title (e.g., "enum ActorEvent" -> "ActorEvent")
		enumName := strings.TrimPrefix(prop.Title, "enum ")

		// Get the Go type for the enum based on underlying type
		enumGoType := "int32"
		if prop.UnderlyingType != "" {
			if m, ok := TypeMapping[prop.UnderlyingType]; ok {
				enumGoType = m.GoType
			}
		}

		enumInfo = &EnumInfo{
			Name:      toGoName(enumName),
			GoType:    enumGoType,
			Values:    prop.Enum,
			FieldName: goName,
		}

		// The field should use the enum's underlying type
		field.GoType = enumGoType
		mapping := mapJSONType(prop.UnderlyingType, prop.SerializationOptions)
		field.IOMethod = mapping.IOMethod
		return field, enumInfo
	}

	// Handle $ref - resolve the reference from definitions
	if prop.Ref != "" {
		resolvedType := resolveRefRecursive(prop.Ref, definitions, packetInfo, 0)
		if resolvedType != nil && !resolvedType.IsComplex {
			field.GoType = resolvedType.GoType
			field.IOMethod = resolvedType.IOMethod
			field.IsNBT = resolvedType.IsNBT
			return field, nil
		}
		// Complex type - try to generate a struct for multi-property references
		// Also generate for single-property wrappers with complex inner types (arrays, objects)
		refDef := resolveRef(prop.Ref, definitions)
		if refDef != nil && len(refDef.Properties) >= 1 {
			nestedStruct := generateStructFromDefinition(refDef, definitions, packetInfo)
			if nestedStruct != nil {
				// Add to packet's structs if not already there
				exists := false
				for _, s := range packetInfo.Structs {
					if s.Name == nestedStruct.Name {
						exists = true
						break
					}
				}
				if !exists {
					packetInfo.Structs = append(packetInfo.Structs, *nestedStruct)
				}
				field.GoType = nestedStruct.Name
				field.IOMethod = "Struct"
				return field, nil
			}
		}
		// Could not resolve reference
		if resolvedType != nil {
			field.GoType = resolvedType.GoType
			field.IOMethod = resolvedType.IOMethod
			field.IsNBT = resolvedType.IsNBT
			field.IsComplex = resolvedType.IsComplex
			field.ComplexReason = resolvedType.ComplexReason
		} else {
			field.GoType = "any"
			field.IsComplex = true
			field.ComplexReason = fmt.Sprintf("unresolved reference: %s", prop.Ref)
		}
		return field, nil
	}

	// Handle array types
	propType := ""
	if t, ok := prop.Type.(string); ok {
		propType = t
	}

	if propType == "array" {
		field.IsSlice = true
		if prop.Items != nil {
			if prop.Items.Ref != "" {
				// Array of referenced type - try to resolve
				resolved := resolveRef(prop.Items.Ref, definitions)
				if resolved != nil {
					// Check well-known types
					if wkt, ok := WellKnownTypes[resolved.Title]; ok {
						field.GoType = "[]" + wkt.GoType
						field.SliceType = wkt.GoType
						field.IOMethod = wkt.IOMethod
						switch wkt.NeedsImport {
						case "mgl32":
							packetInfo.NeedsVecImport = true
						case "nbt":
							packetInfo.NeedsNBTImport = true
						case "uuid":
							packetInfo.NeedsUUIDImport = true
						}
						return field, nil
					}
					// Single property wrapper - inline the inner type
					if len(resolved.Properties) == 1 {
						for _, innerProp := range resolved.Properties {
							// Check if inner property is also a ref
							if innerProp.Ref != "" {
								innerResolved := resolveRefRecursive(innerProp.Ref, definitions, packetInfo, 0)
								if innerResolved != nil && !innerResolved.IsComplex {
									field.GoType = "[]" + innerResolved.GoType
									field.SliceType = innerResolved.GoType
									field.IOMethod = innerResolved.IOMethod
									return field, nil
								}
							}
							mapping := mapJSONType(innerProp.UnderlyingType, innerProp.SerializationOptions)
							if mapping.GoType != "any" && mapping.IOMethod != "" {
								field.GoType = "[]" + mapping.GoType
								field.SliceType = mapping.GoType
								field.IOMethod = mapping.IOMethod
								return field, nil
							}
						}
					}
					// Multi-property reference - generate a struct type
					if len(resolved.Properties) > 1 {
						structInfo := generateStructFromDefinition(resolved, definitions, packetInfo)
						if structInfo != nil {
							// Check if struct already exists
							exists := false
							for _, s := range packetInfo.Structs {
								if s.Name == structInfo.Name {
									exists = true
									break
								}
							}
							if !exists {
								packetInfo.Structs = append(packetInfo.Structs, *structInfo)
							}
							field.GoType = "[]" + structInfo.Name
							field.SliceType = structInfo.Name
							field.IsStructSlice = true
							return field, nil
						}
					}
				}
				field.GoType = "[]any"
				field.IsComplex = true
				field.ComplexReason = "array of complex referenced type"
				return field, nil
			}
			// Prefer UnderlyingType over Type for more accurate type mapping
			itemType := prop.Items.UnderlyingType
			if itemType == "" {
				itemType = prop.Items.Type
			}
			if itemType != "" {
				itemMapping := mapJSONType(itemType, nil)
				field.GoType = "[]" + itemMapping.GoType
				field.SliceType = itemMapping.GoType
				field.IOMethod = itemMapping.IOMethod
				return field, nil
			}
		}
		field.GoType = "[]any"
		field.IsComplex = true
		field.ComplexReason = "array with unknown item type"
		return field, nil
	}

	// Handle oneOf (variant types) - generate variant structs and switch logic
	if len(prop.OneOf) > 0 {
		field.IsOneOf = true
		field.GoType = "any"

		// Check if this is a primitive oneOf (variants have type, not $ref)
		hasPrimitiveVariants := false
		for _, variant := range prop.OneOf {
			if variant.Type != "" {
				hasPrimitiveVariants = true
				break
			}
		}

		if hasPrimitiveVariants && prop.ControlValueType != "" {
			// Primitive oneOf with control value type (like GameRule's RuleValue)
			field.IsPrimitiveOneOf = true
			field.ControlValueType = prop.ControlValueType

			for _, variant := range prop.OneOf {
				v := OneOfVariant{
					Index: variant.OrdinalIndex,
				}
				switch variant.Type {
				case "null":
					v.IsNull = true
				case "boolean":
					v.GoType = "bool"
					v.IOMethod = "Bool"
				case "integer":
					if variant.UnderlyingType == "int32" {
						v.GoType = "int32"
						v.IOMethod = "Varint32"
					} else if variant.UnderlyingType == "uint32" {
						v.GoType = "uint32"
						v.IOMethod = "Varuint32"
					} else {
						v.GoType = "int32"
						v.IOMethod = "Varint32"
					}
				case "number":
					v.GoType = "float32"
					v.IOMethod = "Float32"
				}
				field.OneOfVariants = append(field.OneOfVariants, v)
			}
			return field, nil
		}

		// Process each variant with $ref and generate struct types
		for _, variant := range prop.OneOf {
			if variant.Ref == "" {
				continue
			}
			resolved := resolveRef(variant.Ref, definitions)
			if resolved == nil {
				continue
			}

			// Generate struct for this variant
			variantStruct := generateStructFromDefinition(resolved, definitions, packetInfo)
			if variantStruct != nil {
				// Add to packet's structs if not already there
				exists := false
				for _, s := range packetInfo.Structs {
					if s.Name == variantStruct.Name {
						exists = true
						break
					}
				}
				if !exists {
					packetInfo.Structs = append(packetInfo.Structs, *variantStruct)
				}

				field.OneOfVariants = append(field.OneOfVariants, OneOfVariant{
					Index:      variant.OrdinalIndex,
					StructName: variantStruct.Name,
					StructInfo: variantStruct,
				})
			}
		}

		// Find the control field - look for a field with enum values that match variant count
		// The control field is usually the one before this in ordinal order with enum values
		for otherName, otherProp := range packetInfo.Fields {
			_ = otherName // We're iterating to find enum fields
			if len(otherProp.GoType) > 0 && otherProp.OrdinalIndex < field.OrdinalIndex {
				// Check if this could be a control field (has enum, comes before)
				field.ControlField = otherProp.GoName
			}
		}

		return field, nil
	}

	// Determine the underlying type
	underlyingType := prop.UnderlyingType
	if underlyingType == "" {
		underlyingType = propType
	}

	// Check for serialization options that affect the IO method
	mapping := mapJSONType(underlyingType, prop.SerializationOptions)
	field.GoType = mapping.GoType
	field.IOMethod = mapping.IOMethod

	return field, nil
}

func resolveRef(ref string, definitions map[string]DefinitionJSON) *DefinitionJSON {
	// Refs look like "#/definitions/3541243607"
	if !strings.HasPrefix(ref, "#/definitions/") {
		return nil
	}
	defID := strings.TrimPrefix(ref, "#/definitions/")
	if def, ok := definitions[defID]; ok {
		return &def
	}
	return nil
}

// resolvedType holds the result of recursive ref resolution
type resolvedType struct {
	GoType        string
	IOMethod      string
	IsNBT         bool
	IsComplex     bool
	ComplexReason string
}

// resolveRefRecursive recursively resolves $ref references, unwrapping single-property wrappers
// until we hit a primitive type or a well-known type
func resolveRefRecursive(ref string, definitions map[string]DefinitionJSON, packetInfo *PacketInfo, depth int) *resolvedType {
	// Prevent infinite recursion
	if depth > 10 {
		return &resolvedType{GoType: "any", IsComplex: true, ComplexReason: "max recursion depth reached"}
	}

	defID := strings.TrimPrefix(ref, "#/definitions/")

	// First check if this is a special definition ID (like NBT compound)
	if special, ok := SpecialDefinitionIDs[defID]; ok {
		result := &resolvedType{
			GoType:   special.GoType,
			IOMethod: special.IOMethod,
		}
		if special.NeedsImport == "nbt" {
			packetInfo.NeedsNBTImport = true
			result.IsNBT = true
		} else if special.NeedsImport == "mgl32" {
			packetInfo.NeedsVecImport = true
		}
		return result
	}

	resolved := resolveRef(ref, definitions)
	if resolved == nil {
		return nil
	}

	// If the definition itself has a $ref, follow it (type alias)
	if resolved.Ref != "" {
		return resolveRefRecursive(resolved.Ref, definitions, packetInfo, depth+1)
	}

	// Check if it's a well-known type by title
	if wkt, ok := WellKnownTypes[resolved.Title]; ok {
		result := &resolvedType{
			GoType:   wkt.GoType,
			IOMethod: wkt.IOMethod,
		}
		switch wkt.NeedsImport {
		case "mgl32":
			packetInfo.NeedsVecImport = true
		case "nbt":
			packetInfo.NeedsNBTImport = true
			result.IsNBT = true
		case "uuid":
			packetInfo.NeedsUUIDImport = true
		}
		return result
	}

	// Handle simple type definitions with no properties (type aliases like "Reference" -> string)
	if len(resolved.Properties) == 0 && resolved.Type != "" {
		mapping := mapJSONType(resolved.Type, nil)
		if mapping.GoType != "any" && mapping.IOMethod != "" {
			return &resolvedType{
				GoType:   mapping.GoType,
				IOMethod: mapping.IOMethod,
			}
		}
	}

	// If it's a single-property wrapper, recursively resolve
	if len(resolved.Properties) == 1 {
		for propName, innerProp := range resolved.Properties {
			// If the inner property is also a ref, recursively resolve
			if innerProp.Ref != "" {
				return resolveRefRecursive(innerProp.Ref, definitions, packetInfo, depth+1)
			}

			// Check if inner property is an array or object - these need struct generation
			if t, ok := innerProp.Type.(string); ok {
				if t == "array" || t == "object" {
					// This is a wrapper around a complex type - mark as needing struct generation
					return &resolvedType{
						GoType:        "any",
						IsComplex:     true,
						ComplexReason: fmt.Sprintf("wrapper with %s property '%s'", t, propName),
					}
				}
			}

			// Otherwise, use the inner property's type info
			mapping := mapJSONType(innerProp.UnderlyingType, innerProp.SerializationOptions)
			if mapping.GoType != "any" && mapping.IOMethod != "" {
				return &resolvedType{
					GoType:   mapping.GoType,
					IOMethod: mapping.IOMethod,
				}
			}

			// Try to get type from the property's type field
			if t, ok := innerProp.Type.(string); ok && t != "" {
				mapping = mapJSONType(t, innerProp.SerializationOptions)
				if mapping.GoType != "any" && mapping.IOMethod != "" {
					return &resolvedType{
						GoType:   mapping.GoType,
						IOMethod: mapping.IOMethod,
					}
				}
			}
		}
	}

	// Multi-property reference - mark as complex, generate struct
	return &resolvedType{
		GoType:        "any",
		IsComplex:     true,
		ComplexReason: fmt.Sprintf("multi-property reference to %s", resolved.Title),
	}
}

// generateStructFromDefinition creates a StructInfo from a definition
func generateStructFromDefinition(def *DefinitionJSON, definitions map[string]DefinitionJSON, packetInfo *PacketInfo) *StructInfo {
	if def == nil || len(def.Properties) == 0 {
		return nil
	}

	// Generate struct name from title, removing common prefixes
	structName := def.Title
	// Remove "struct " prefix if present
	structName = strings.TrimPrefix(structName, "struct ")
	// Remove packet-specific prefixes for cleaner names
	structName = strings.TrimPrefix(structName, "AvailableCommandsPacket")
	structName = toGoName(structName)

	// Avoid collision with the packet name itself (use GoName which is the stripped version)
	if structName == packetInfo.GoName {
		structName = structName + "Entry"
	}

	// Build required set
	requiredSet := make(map[string]bool)
	for _, req := range def.Required {
		requiredSet[req] = true
	}

	// Process fields
	var fields []FieldInfo
	for propName, prop := range def.Properties {
		fieldInfo := processStructField(propName, prop, definitions, packetInfo, requiredSet[propName])
		if fieldInfo != nil {
			fields = append(fields, *fieldInfo)
		}
	}

	// Sort fields by ordinal index
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].OrdinalIndex < fields[j].OrdinalIndex
	})

	return &StructInfo{
		Name:   structName,
		Fields: fields,
	}
}

// processStructField processes a field within a nested struct definition
func processStructField(name string, prop PropertyJSON, definitions map[string]DefinitionJSON, packetInfo *PacketInfo, isRequired bool) *FieldInfo {
	goName := toGoName(name)

	field := &FieldInfo{
		Name:         name,
		GoName:       goName,
		OrdinalIndex: prop.OrdinalIndex,
		Description:  prop.Description,
		IsOptional:   !isRequired,
	}

	// Handle refs
	if prop.Ref != "" {
		resolved := resolveRefRecursive(prop.Ref, definitions, packetInfo, 0)
		if resolved != nil && !resolved.IsComplex {
			field.GoType = resolved.GoType
			field.IOMethod = resolved.IOMethod
			field.IsNBT = resolved.IsNBT
			return field
		}
		// Check if it's a multi-property struct we need to generate
		refDef := resolveRef(prop.Ref, definitions)
		if refDef != nil && len(refDef.Properties) > 1 {
			nestedStruct := generateStructFromDefinition(refDef, definitions, packetInfo)
			if nestedStruct != nil {
				// Add to packet's structs if not already there
				exists := false
				for _, s := range packetInfo.Structs {
					if s.Name == nestedStruct.Name {
						exists = true
						break
					}
				}
				if !exists {
					packetInfo.Structs = append(packetInfo.Structs, *nestedStruct)
				}
				field.GoType = nestedStruct.Name
				field.IOMethod = "Struct" // Will use protocol.Single
				return field
			}
		}
		field.GoType = "any"
		field.IsComplex = true
		field.ComplexReason = "unresolved struct reference"
		return field
	}

	// Handle arrays
	propType := ""
	if t, ok := prop.Type.(string); ok {
		propType = t
	}

	if propType == "array" {
		field.IsSlice = true
		if prop.Items != nil {
			if prop.Items.Ref != "" {
				// Array of referenced type
				resolved := resolveRef(prop.Items.Ref, definitions)
				if resolved != nil {
					// Check well-known types first
					if wkt, ok := WellKnownTypes[resolved.Title]; ok {
						field.GoType = "[]" + wkt.GoType
						field.SliceType = wkt.GoType
						field.IOMethod = wkt.IOMethod
						switch wkt.NeedsImport {
						case "mgl32":
							packetInfo.NeedsVecImport = true
						case "nbt":
							packetInfo.NeedsNBTImport = true
						case "uuid":
							packetInfo.NeedsUUIDImport = true
						}
						return field
					}
					// Check for simple type definition (no properties, just a type alias)
					if len(resolved.Properties) == 0 && resolved.Type != "" {
						mapping := mapJSONType(resolved.Type, nil)
						if mapping.GoType != "any" && mapping.IOMethod != "" {
							field.GoType = "[]" + mapping.GoType
							field.SliceType = mapping.GoType
							field.IOMethod = mapping.IOMethod
							return field
						}
					}
					// Check for single-property wrapper
					if len(resolved.Properties) == 1 {
						for _, innerProp := range resolved.Properties {
							if innerProp.Ref != "" {
								innerResolved := resolveRefRecursive(innerProp.Ref, definitions, packetInfo, 0)
								if innerResolved != nil && !innerResolved.IsComplex {
									field.GoType = "[]" + innerResolved.GoType
									field.SliceType = innerResolved.GoType
									field.IOMethod = innerResolved.IOMethod
									return field
								}
							}
							mapping := mapJSONType(innerProp.UnderlyingType, innerProp.SerializationOptions)
							if mapping.GoType != "any" && mapping.IOMethod != "" {
								field.GoType = "[]" + mapping.GoType
								field.SliceType = mapping.GoType
								field.IOMethod = mapping.IOMethod
								return field
							}
						}
					}
					// Multi-property reference - generate nested struct
					if len(resolved.Properties) > 1 {
						nestedStruct := generateStructFromDefinition(resolved, definitions, packetInfo)
						if nestedStruct != nil {
							exists := false
							for _, s := range packetInfo.Structs {
								if s.Name == nestedStruct.Name {
									exists = true
									break
								}
							}
							if !exists {
								packetInfo.Structs = append(packetInfo.Structs, *nestedStruct)
							}
							field.GoType = "[]" + nestedStruct.Name
							field.SliceType = nestedStruct.Name
							field.IsStructSlice = true
							return field
						}
					}
				}
				field.GoType = "[]any"
				field.IsComplex = true
				field.ComplexReason = "array of complex type"
				return field
			}
			// Array of primitive type - prefer UnderlyingType over Type
			itemType := prop.Items.UnderlyingType
			if itemType == "" {
				itemType = prop.Items.Type
			}
			if itemType != "" {
				mapping := mapJSONType(itemType, nil)
				field.GoType = "[]" + mapping.GoType
				field.SliceType = mapping.GoType
				field.IOMethod = mapping.IOMethod
				return field
			}
		}
		field.GoType = "[]any"
		field.IsComplex = true
		field.ComplexReason = "array with unknown type"
		return field
	}

	// Handle object with additionalProperties (map-like types serialized as array of key-value structs)
	if propType == "object" && prop.AdditionalProperties != nil {
		addProps := prop.AdditionalProperties
		if addProps.Properties != nil {
			keyProp, hasKey := addProps.Properties["key"]
			valueProp, hasValue := addProps.Properties["value"]
			if hasKey && hasValue {
				// Generate a key-value struct
				keyType := resolvePropertyType(keyProp, definitions, packetInfo)
				valueType := resolvePropertyType(valueProp, definitions, packetInfo)

				// Create struct name based on field name
				kvStructName := goName + "Entry"

				// Generate the key-value struct
				kvStruct := StructInfo{
					Name: kvStructName,
					Fields: []FieldInfo{
						{
							Name:     "Key",
							GoName:   "Key",
							GoType:   keyType.GoType,
							IOMethod: keyType.IOMethod,
						},
						{
							Name:     "Value",
							GoName:   "Value",
							GoType:   valueType.GoType,
							IOMethod: valueType.IOMethod,
						},
					},
				}

				// Add to packet's structs if not already there
				exists := false
				for _, s := range packetInfo.Structs {
					if s.Name == kvStructName {
						exists = true
						break
					}
				}
				if !exists {
					packetInfo.Structs = append(packetInfo.Structs, kvStruct)
				}

				field.GoType = "[]" + kvStructName
				field.SliceType = kvStructName
				field.IsStructSlice = true
				return field
			}
		}
		// Fallback for unknown additionalProperties structure
		field.GoType = "any"
		field.IsComplex = true
		field.ComplexReason = "unknown additionalProperties structure"
		return field
	}

	// Handle oneOf (variant types) in struct fields
	if len(prop.OneOf) > 0 {
		field.IsOneOf = true
		field.GoType = "any"

		// Check if this is a primitive oneOf (variants have type, not $ref)
		hasPrimitiveVariants := false
		for _, variant := range prop.OneOf {
			if variant.Type != "" {
				hasPrimitiveVariants = true
				break
			}
		}

		if hasPrimitiveVariants && prop.ControlValueType != "" {
			// Primitive oneOf with control value type (like GameRule's RuleValue)
			field.IsPrimitiveOneOf = true
			field.ControlValueType = prop.ControlValueType

			for _, variant := range prop.OneOf {
				v := OneOfVariant{
					Index: variant.OrdinalIndex,
				}
				switch variant.Type {
				case "null":
					v.IsNull = true
				case "boolean":
					v.GoType = "bool"
					v.IOMethod = "Bool"
				case "integer":
					if variant.UnderlyingType == "int32" {
						v.GoType = "int32"
						v.IOMethod = "Varint32"
					} else if variant.UnderlyingType == "uint32" {
						v.GoType = "uint32"
						v.IOMethod = "Varuint32"
					} else {
						v.GoType = "int32"
						v.IOMethod = "Varint32"
					}
				case "number":
					v.GoType = "float32"
					v.IOMethod = "Float32"
				}
				field.OneOfVariants = append(field.OneOfVariants, v)
			}
			return field
		}

		// Complex oneOf - mark as needing manual handling
		field.IsComplex = true
		field.ComplexReason = "oneOf with complex variants"
		return field
	}

	// Handle primitive types
	underlyingType := prop.UnderlyingType
	if underlyingType == "" {
		underlyingType = propType
	}

	mapping := mapJSONType(underlyingType, prop.SerializationOptions)
	field.GoType = mapping.GoType
	field.IOMethod = mapping.IOMethod

	return field
}

type typeMapping struct {
	GoType   string
	IOMethod string
}

// resolvePropertyType resolves a property to its Go type and IO method
func resolvePropertyType(prop PropertyJSON, definitions map[string]DefinitionJSON, packetInfo *PacketInfo) typeMapping {
	// Handle $ref
	if prop.Ref != "" {
		resolved := resolveRefRecursive(prop.Ref, definitions, packetInfo, 0)
		if resolved != nil && !resolved.IsComplex {
			return typeMapping{GoType: resolved.GoType, IOMethod: resolved.IOMethod}
		}
	}

	// Handle primitive type
	propType := ""
	if t, ok := prop.Type.(string); ok {
		propType = t
	}

	underlyingType := prop.UnderlyingType
	if underlyingType == "" {
		underlyingType = propType
	}

	return mapJSONType(underlyingType, prop.SerializationOptions)
}

func mapJSONType(jsonType string, serializationOpts []string) typeMapping {
	// Check for big-endian serialization
	isBigEndian := false
	isCompressed := false
	for _, opt := range serializationOpts {
		if opt == "Big-Endian" || opt == "BigEndian" {
			isBigEndian = true
		}
		if opt == "Compression" || opt == "VarInt" || opt == "Zigzag" {
			isCompressed = true
		}
	}

	// Handle big-endian types
	if isBigEndian {
		switch jsonType {
		case "int32":
			return typeMapping{"int32", "BEInt32"}
		case "uint32":
			return typeMapping{"uint32", "BEUint32"}
		case "int16":
			return typeMapping{"int16", "BEInt16"}
		case "uint16":
			return typeMapping{"uint16", "BEUint16"}
		case "int64":
			return typeMapping{"int64", "BEInt64"}
		case "uint64":
			return typeMapping{"uint64", "BEUint64"}
		case "float32", "float":
			return typeMapping{"float32", "BEFloat32"}
		case "float64", "double":
			return typeMapping{"float64", "BEFloat64"}
		}
	}

	// Handle compressed/varint types
	if isCompressed {
		switch jsonType {
		case "int32":
			return typeMapping{"int32", "Varint32"}
		case "uint32":
			return typeMapping{"uint32", "Varuint32"}
		case "int64":
			return typeMapping{"int64", "Varint64"}
		case "uint64":
			return typeMapping{"uint64", "Varuint64"}
		}
	}

	// Default mappings
	if m, ok := TypeMapping[jsonType]; ok {
		return typeMapping{m.GoType, m.IOMethod}
	}

	// Special types
	switch jsonType {
	case "integer":
		return typeMapping{"int32", "Int32"}
	case "number":
		return typeMapping{"float32", "Float32"}
	case "object":
		return typeMapping{"any", ""}
	default:
		return typeMapping{"any", ""}
	}
}

func generatePacketFile(packet PacketInfo) error {
	// Generate content first (without imports) to check what's actually used
	var contentBuilder strings.Builder

	// Generate enum constants
	for _, enum := range packet.Enums {
		generateEnumConstants(&contentBuilder, packet.GoName, enum)
	}

	// Generate nested struct types
	for _, structInfo := range packet.Structs {
		generateStructType(&contentBuilder, structInfo)
	}

	// Add description comment
	if packet.Description != "" {
		contentBuilder.WriteString(fmt.Sprintf("// %s %s\n", packet.GoName, cleanDescription(packet.Description)))
	} else {
		contentBuilder.WriteString(fmt.Sprintf("// %s is a packet with ID %d.\n", packet.GoName, packet.ID))
	}

	// Struct definition
	contentBuilder.WriteString(fmt.Sprintf("type %s struct {\n", packet.GoName))
	for _, f := range packet.Fields {
		if f.Description != "" {
			contentBuilder.WriteString(fmt.Sprintf("\t// %s %s\n", f.GoName, cleanDescription(f.Description)))
		}
		if f.IsOptional && !f.IsComplex {
			contentBuilder.WriteString(fmt.Sprintf("\t%s protocol.Optional[%s]\n", f.GoName, f.GoType))
		} else {
			contentBuilder.WriteString(fmt.Sprintf("\t%s %s\n", f.GoName, f.GoType))
		}
	}
	contentBuilder.WriteString("}\n\n")

	// ID method
	contentBuilder.WriteString("// ID ...\n")
	contentBuilder.WriteString(fmt.Sprintf("func (*%s) ID() uint32 {\n", packet.GoName))
	contentBuilder.WriteString(fmt.Sprintf("\treturn ID%s\n", packet.GoName))
	contentBuilder.WriteString("}\n\n")

	// Marshal method
	contentBuilder.WriteString(fmt.Sprintf("func (pk *%s) Marshal(io protocol.IO) {\n", packet.GoName))
	for _, f := range packet.Fields {
		if f.IsComplex {
			contentBuilder.WriteString(fmt.Sprintf("\t// TODO: %s - %s\n", f.GoName, f.ComplexReason))
			continue
		}
		// Handle oneOf variant types with switch statement
		if f.IsOneOf && len(f.OneOfVariants) > 0 {
			generateOneOfSwitch(&contentBuilder, f, packet)
			continue
		}
		// Check struct slices BEFORE IOMethod check (they don't need IOMethod)
		if f.IsStructSlice {
			contentBuilder.WriteString(fmt.Sprintf("\tprotocol.Slice(io, &pk.%s)\n", f.GoName))
			continue
		}
		if f.IOMethod == "" {
			contentBuilder.WriteString(fmt.Sprintf("\t// TODO: %s - type mapping not found for %s\n", f.GoName, f.GoType))
			continue
		}
		if f.IsOptional {
			// Optional fields use protocol.OptionalFunc or OptionalMarshaler
			if f.IOMethod == "Struct" {
				// Optional structs use protocol.OptionalMarshaler
				contentBuilder.WriteString(fmt.Sprintf("\tprotocol.OptionalMarshaler(io, &pk.%s)\n", f.GoName))
			} else {
				contentBuilder.WriteString(fmt.Sprintf("\tprotocol.OptionalFunc(io, &pk.%s, io.%s)\n", f.GoName, f.IOMethod))
			}
		} else if f.IsNBT {
			// NBT fields need special serialization with encoding type
			contentBuilder.WriteString(fmt.Sprintf("\tio.NBT(&pk.%s, nbt.NetworkLittleEndian)\n", f.GoName))
		} else if f.IOMethod == "Struct" {
			// Nested struct uses protocol.Single
			contentBuilder.WriteString(fmt.Sprintf("\tprotocol.Single(io, &pk.%s)\n", f.GoName))
		} else if f.IsSlice {
			contentBuilder.WriteString(fmt.Sprintf("\tprotocol.FuncSlice(io, &pk.%s, io.%s)\n", f.GoName, f.IOMethod))
		} else {
			contentBuilder.WriteString(fmt.Sprintf("\tio.%s(&pk.%s)\n", f.IOMethod, f.GoName))
		}
	}
	contentBuilder.WriteString("}\n")

	// Get the generated content
	content := contentBuilder.String()

	// Check which imports are actually needed by scanning the content
	needsUUID := strings.Contains(content, "uuid.UUID") || strings.Contains(content, "io.UUID")
	needsVec := strings.Contains(content, "mgl32.")
	needsNBT := strings.Contains(content, "nbt.")

	// Build the final file
	var sb strings.Builder

	// Generated file header
	sb.WriteString("// Code generated by protocol/packet generator; DO NOT EDIT.\n")
	sb.WriteString(fmt.Sprintf("// Generated at:       %s\n", generationTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("// Minecraft version:  %s\n", packet.MinecraftVersion))
	sb.WriteString(fmt.Sprintf("// Protocol version:   %d\n", packet.ProtocolVersion))
	sb.WriteString(fmt.Sprintf("// Packet ID:          %d (0x%02X)\n", packet.ID, packet.ID))
	sb.WriteString("\n")

	// Package declaration
	sb.WriteString("package packet\n\n")

	// Add imports based on actual usage
	sb.WriteString("import (\n")
	if needsUUID {
		sb.WriteString("\t\"github.com/google/uuid\"\n")
	}
	if needsVec {
		sb.WriteString("\t\"github.com/go-gl/mathgl/mgl32\"\n")
	}
	if needsNBT {
		sb.WriteString("\t\"github.com/sandertv/gophertunnel/minecraft/nbt\"\n")
	}
	sb.WriteString("\t\"github.com/sandertv/gophertunnel/minecraft/protocol\"\n")
	sb.WriteString(")\n\n")

	// Add the content
	sb.WriteString(content)

	// Write to file
	filePath := filepath.Join(outputDir, packet.FileName)
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func generateEnumConstants(sb *strings.Builder, packetName string, enum EnumInfo) {
	if len(enum.Values) == 0 {
		return
	}

	// Use just the enum name as prefix (e.g., "ActorEvent" not "ActorEventActorEvent")
	prefix := enum.Name

	// Check if the first constant would be a duplicate - if so, skip this entire enum
	firstConstName := prefix + toGoEnumValueName(enum.Values[0])
	if generatedEnumConstants[firstConstName] {
		return // Skip - this enum's constants were already generated by another packet
	}

	sb.WriteString("const (\n")

	for i, val := range enum.Values {
		constName := prefix + toGoEnumValueName(val)
		generatedEnumConstants[constName] = true
		if i == 0 {
			// First value - start with iota
			sb.WriteString(fmt.Sprintf("\t%s = iota\n", constName))
		} else {
			sb.WriteString(fmt.Sprintf("\t%s\n", constName))
		}
	}

	sb.WriteString(")\n\n")
}

// generateStructType generates a struct type definition and its Marshal method
func generateStructType(sb *strings.Builder, structInfo StructInfo) {
	// Check if this struct was already generated by another packet
	if generatedStructNames[structInfo.Name] {
		return // Skip - already generated
	}
	generatedStructNames[structInfo.Name] = true

	// Struct definition
	sb.WriteString(fmt.Sprintf("// %s is a nested struct used in the packet.\n", structInfo.Name))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structInfo.Name))
	for _, f := range structInfo.Fields {
		if f.Description != "" {
			sb.WriteString(fmt.Sprintf("\t// %s %s\n", f.GoName, cleanDescription(f.Description)))
		}
		sb.WriteString(fmt.Sprintf("\t%s %s\n", f.GoName, f.GoType))
	}
	sb.WriteString("}\n\n")

	// Marshal method
	sb.WriteString(fmt.Sprintf("func (x *%s) Marshal(r protocol.IO) {\n", structInfo.Name))
	for _, f := range structInfo.Fields {
		if f.IsComplex {
			sb.WriteString(fmt.Sprintf("\t// TODO: %s - %s\n", f.GoName, f.ComplexReason))
			continue
		}
		// Handle primitive oneOf with type switch
		if f.IsPrimitiveOneOf && len(f.OneOfVariants) > 0 {
			generatePrimitiveOneOfSwitch(sb, f, "x")
			continue
		}
		// Check struct slices BEFORE IOMethod check (they don't need IOMethod)
		if f.IsStructSlice {
			sb.WriteString(fmt.Sprintf("\tprotocol.Slice(r, &x.%s)\n", f.GoName))
			continue
		}
		if f.IOMethod == "" {
			sb.WriteString(fmt.Sprintf("\t// TODO: %s - type mapping not found\n", f.GoName))
			continue
		}
		if f.IOMethod == "Struct" {
			// Nested struct uses protocol.Single
			sb.WriteString(fmt.Sprintf("\tprotocol.Single(r, &x.%s)\n", f.GoName))
		} else if f.IsNBT {
			// NBT fields need encoding parameter
			sb.WriteString(fmt.Sprintf("\tr.NBT(&x.%s, nbt.NetworkLittleEndian)\n", f.GoName))
		} else if f.IsSlice {
			sb.WriteString(fmt.Sprintf("\tprotocol.FuncSlice(r, &x.%s, r.%s)\n", f.GoName, f.IOMethod))
		} else {
			sb.WriteString(fmt.Sprintf("\tr.%s(&x.%s)\n", f.IOMethod, f.GoName))
		}
	}
	sb.WriteString("}\n\n")
}

// generatePrimitiveOneOfSwitch generates a type switch for oneOf with primitive types
func generatePrimitiveOneOfSwitch(sb *strings.Builder, field FieldInfo, receiver string) {
	// First, write the type indicator based on the actual value type
	// Then write the value itself
	controlType := field.ControlValueType
	if controlType == "" {
		controlType = "uint32"
	}

	// Map control type to IO method
	controlIOMethod := "Uint32"
	switch controlType {
	case "uint8":
		controlIOMethod = "Uint8"
	case "uint16":
		controlIOMethod = "Uint16"
	case "uint32":
		controlIOMethod = "Uint32"
	case "int32":
		controlIOMethod = "Int32"
	}

	sb.WriteString(fmt.Sprintf("\tswitch val := %s.%s.(type) {\n", receiver, field.GoName))
	for _, v := range field.OneOfVariants {
		if v.IsNull {
			sb.WriteString("\tcase nil:\n")
			sb.WriteString(fmt.Sprintf("\t\tvar typeID %s = %d\n", controlType, v.Index))
			sb.WriteString(fmt.Sprintf("\t\tr.%s(&typeID)\n", controlIOMethod))
			sb.WriteString("\t\t_ = val // nil value, nothing more to write\n")
		} else if v.GoType != "" {
			sb.WriteString(fmt.Sprintf("\tcase %s:\n", v.GoType))
			sb.WriteString(fmt.Sprintf("\t\tvar typeID %s = %d\n", controlType, v.Index))
			sb.WriteString(fmt.Sprintf("\t\tr.%s(&typeID)\n", controlIOMethod))
			sb.WriteString(fmt.Sprintf("\t\tr.%s(&val)\n", v.IOMethod))
		}
	}
	sb.WriteString("\t}\n")
}

// generateOneOfSwitch generates a switch statement for oneOf variant fields
func generateOneOfSwitch(sb *strings.Builder, field FieldInfo, packet PacketInfo) {
	// Find the control field - usually the enum field that determines which variant to use
	controlField := ""
	for _, f := range packet.Fields {
		if f.OrdinalIndex < field.OrdinalIndex && len(packet.Enums) > 0 {
			// Check if there's an enum associated with this field
			for _, enum := range packet.Enums {
				if enum.FieldName == f.GoName {
					controlField = f.GoName
					break
				}
			}
		}
	}

	if controlField == "" {
		// Try to find a field that could be the control
		for _, f := range packet.Fields {
			if f.OrdinalIndex < field.OrdinalIndex && (strings.Contains(f.GoName, "Type") || strings.Contains(f.GoName, "Event")) {
				controlField = f.GoName
				break
			}
		}
	}

	if controlField == "" {
		sb.WriteString(fmt.Sprintf("\t// TODO: %s - oneOf variants generated but control field unknown\n", field.GoName))
		sb.WriteString("\t// Variants: ")
		for i, v := range field.OneOfVariants {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s(%d)", v.StructName, v.Index))
		}
		sb.WriteString("\n")
		return
	}

	sb.WriteString(fmt.Sprintf("\tswitch pk.%s {\n", controlField))
	for _, variant := range field.OneOfVariants {
		if variant.StructInfo == nil || len(variant.StructInfo.Fields) == 0 {
			// Empty variant (like "Empty" type)
			sb.WriteString(fmt.Sprintf("\tcase %d:\n", variant.Index))
			sb.WriteString("\t\t// Empty variant - no additional data\n")
			continue
		}
		sb.WriteString(fmt.Sprintf("\tcase %d: // %s\n", variant.Index, variant.StructName))
		sb.WriteString(fmt.Sprintf("\t\tx := pk.%s.(%s)\n", field.GoName, variant.StructName))
		sb.WriteString("\t\tx.Marshal(io)\n")
	}
	sb.WriteString("\t}\n")
}

// toGoEnumValueName converts an enum value name to Go style
func toGoEnumValueName(name string) string {
	// If the name is already PascalCase (no underscores, has mixed case), use it as-is
	if !strings.Contains(name, "_") {
		// Check if it looks like PascalCase (starts with uppercase)
		if len(name) > 0 && unicode.IsUpper(rune(name[0])) {
			return name
		}
		// Single word lowercase - capitalize first letter
		if len(name) > 0 {
			runes := []rune(name)
			runes[0] = unicode.ToUpper(runes[0])
			return string(runes)
		}
		return name
	}

	// Convert SCREAMING_SNAKE_CASE to PascalCase
	parts := strings.Split(strings.ToLower(name), "_")
	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}
	return result.String()
}

func generateIDFile(packets []PacketInfo) error {
	var sb strings.Builder

	// Get protocol info from first packet
	mcVersion := ""
	protoVersion := 0
	if len(packets) > 0 {
		mcVersion = packets[0].MinecraftVersion
		protoVersion = packets[0].ProtocolVersion
	}

	sb.WriteString("// Code generated by protocol/packet generator; DO NOT EDIT.\n")
	sb.WriteString(fmt.Sprintf("// Generated at:       %s\n", generationTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("// Minecraft version:  %s\n", mcVersion))
	sb.WriteString(fmt.Sprintf("// Protocol version:   %d\n", protoVersion))
	sb.WriteString("\n")
	sb.WriteString("package packet\n\n")
	sb.WriteString("// Packet IDs for Minecraft Bedrock Protocol\n")
	sb.WriteString("const (\n")

	// Group packets by consecutive IDs to use iota where possible
	if len(packets) > 0 {
		lastID := packets[0].ID - 1
		firstInGroup := true

		for _, p := range packets {
			if p.ID == lastID+1 && !firstInGroup {
				// Consecutive, use simple constant
				sb.WriteString(fmt.Sprintf("\tID%s\n", p.GoName))
			} else if firstInGroup {
				// First in group, use iota + offset
				sb.WriteString(fmt.Sprintf("\tID%s = iota + %d\n", p.GoName, p.ID))
				firstInGroup = false
			} else {
				// Gap in IDs, add underscores or explicit value
				gap := p.ID - lastID - 1
				for i := 0; i < gap; i++ {
					sb.WriteString("\t_\n")
				}
				sb.WriteString(fmt.Sprintf("\tID%s\n", p.GoName))
			}
			lastID = p.ID
		}
	}

	sb.WriteString(")\n")

	filePath := filepath.Join(outputDir, "id.go")
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func generateDocFile(packets []PacketInfo) error {
	// Get protocol info from first packet
	mcVersion := ""
	protoVersion := 0
	if len(packets) > 0 {
		mcVersion = packets[0].MinecraftVersion
		protoVersion = packets[0].ProtocolVersion
	}

	var sb strings.Builder
	sb.WriteString("// Code generated by protocol/packet generator; DO NOT EDIT.\n")
	sb.WriteString(fmt.Sprintf("// Generated at:       %s\n", generationTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("// Minecraft version:  %s\n", mcVersion))
	sb.WriteString(fmt.Sprintf("// Protocol version:   %d\n", protoVersion))
	sb.WriteString("\n")
	sb.WriteString("// Package packet implements all packets used by the Minecraft Bedrock protocol.\n")
	sb.WriteString("// This package was auto-generated from the official Mojang bedrock-protocol-docs.\n")
	sb.WriteString("// See: https://github.com/Mojang/bedrock-protocol-docs\n")
	sb.WriteString("package packet\n")

	filePath := filepath.Join(outputDir, "doc.go")
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

// toGoName converts a JSON field name to a Go-style exported name
func toGoName(name string) string {
	// Handle common abbreviations and special cases
	name = strings.ReplaceAll(name, "'s", "")
	name = strings.ReplaceAll(name, "?", "")
	name = strings.ReplaceAll(name, "::", "") // C++ scope operator
	name = strings.ReplaceAll(name, ":", "")  // Single colon

	// Remove parentheses and their contents (e.g., "(JSON)" -> "")
	name = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(name, "")

	// Remove brackets and their contents
	name = regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(name, "")

	// Remove angle brackets and their contents
	name = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(name, "")

	// Remove any remaining special characters that aren't valid in Go identifiers
	name = regexp.MustCompile(`[^a-zA-Z0-9\s\-_]`).ReplaceAllString(name, "")

	// Split on spaces and special characters
	parts := regexp.MustCompile(`[\s\-_]+`).Split(name, -1)

	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		// Capitalize first letter of each part
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}

	s := result.String()

	// Handle common acronyms - but be careful with word boundaries
	acronymReplacements := map[string]string{
		"Uuid": "UUID",
		"Xuid": "XUID",
		"Url":  "URL",
		"Uri":  "URI",
		"Nbt":  "NBT",
		"Lan":  "LAN",
		"Xbl":  "XBL",
		"Npc":  "NPC",
		"Db":   "DB",
	}

	for old, new := range acronymReplacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Handle "Id" -> "ID" more carefully to avoid issues like "Identifiers" -> "IDentifiers"
	s = replaceIDSafely(s)

	// If the name became empty after stripping, use a fallback
	if s == "" {
		return "Data"
	}

	return s
}

// replaceIDSafely replaces "Id" with "ID" only at word boundaries
func replaceIDSafely(s string) string {
	result := []rune(s)
	for i := 0; i < len(result)-1; i++ {
		if result[i] == 'I' && result[i+1] == 'd' {
			// Check if this is at end of string or followed by uppercase
			if i+2 >= len(result) || unicode.IsUpper(result[i+2]) || !unicode.IsLetter(result[i+2]) {
				result[i+1] = 'D'
			}
		}
	}
	return string(result)
}

// toSnakeCase converts a PascalCase name to snake_case
func toSnakeCase(name string) string {
	var result strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Add underscore before uppercase if:
			// - not at the start
			// - previous char is lowercase, OR
			// - next char exists and is lowercase (handles acronyms like NPC -> npc not n_p_c)
			if i > 0 {
				prevLower := unicode.IsLower(runes[i-1])
				nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if prevLower || nextLower {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// cleanDescription cleans up a description for use in Go comments
func cleanDescription(desc string) string {
	// Remove newlines and excessive whitespace
	desc = strings.ReplaceAll(desc, "\n", " ")
	desc = strings.ReplaceAll(desc, "\r", " ")
	desc = regexp.MustCompile(`\s+`).ReplaceAllString(desc, " ")
	desc = strings.TrimSpace(desc)

	// Ensure it starts lowercase if it's a continuation
	if len(desc) > 0 && !unicode.IsUpper(rune(desc[0])) {
		desc = strings.ToLower(string(desc[0])) + desc[1:]
	}

	return desc
}
