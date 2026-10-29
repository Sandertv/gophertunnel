# Bedrock Protocol Packet Generator

This script generates Go packet definitions from the official Mojang [bedrock-protocol-docs](https://github.com/Mojang/bedrock-protocol-docs) JSON documentation.

## Usage

```bash
cd scripts
go run main.go
```

This will:
1. Fetch packet definitions from the Mojang bedrock-protocol-docs repository
2. Generate Go source files in `minecraft/protocol/packet/__generated__/`

## Generated Output

The generator creates:
- **Individual packet files** (e.g., `set_time.go`, `transfer.go`) - one per packet
- **`id.go`** - Packet ID constants
- **`doc.go`** - Package documentation

Each file includes a header with:
- Generation timestamp
- Minecraft version (e.g., `1.21.130`)
- Protocol version (e.g., `897`)
- Packet ID (decimal and hex)

## What Gets Generated

For each packet, the generator creates:
- A Go struct with all fields
- Field comments from the JSON docs
- An `ID()` method returning the packet's constant ID
- A `Marshal(io protocol.IO)` method for serialization/deserialization

## `$ref` Resolution

The generator automatically resolves `$ref` references to their definitions. For simple wrapper types (single-property definitions like `ActorRuntimeID`), it inlines the actual type:

```json
"Target Runtime ID": {
    "$ref": "#/definitions/3541243607"
}
```

Where definition `3541243607` contains:
```json
{
    "title": "ActorRuntimeID",
    "properties": {
        "Actor Runtime ID": {
            "x-underlying-type": "uint64",
            "x-serialization-options": ["Compression"]
        }
    }
}
```

Becomes:
```go
EntityRuntimeID uint64
// ...
io.Varuint64(&pk.EntityRuntimeID)
```