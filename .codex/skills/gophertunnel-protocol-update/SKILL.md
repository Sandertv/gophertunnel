---
name: gophertunnel-protocol-update
description: Use when updating, reviewing, or debugging gophertunnel for a new Minecraft Bedrock protocol version, especially packet codecs, protocol constants, Cloudburst/PMMP/Mojang doc comparisons, live packet decode errors, or PR preparation for HashimTheArab/gophertunnel.
---

# Gophertunnel Protocol Update

## Goal

Update gophertunnel protocol code from scratch with source-backed evidence. Prefer correct wire compatibility over matching any single upstream implementation.

## Source Order

1. Identify the exact target protocol and Minecraft version first.
2. Check Mojang `bedrock-protocol-docs` for current schemas and `x-ordinal-index` order.
3. Check Cloudburst for the exact target codec version and packet serializer.
4. Check PMMP BedrockProtocol current implementation for independent codec evidence.
5. Use LeviLamina/BDS symbols for C++ field names and layout hints, but verify that the commit matches the target protocol before treating it as current.
6. Use live BDS/client packet bytes when sources disagree or decode still fails.

Do not pull changes from earlier protocol updates by accident. Cloudburst often updates previous codec classes too; verify that the class or registration is specifically used by the target protocol.

## Cross-Check Rules

- Treat Cloudburst read and write paths as separate evidence. If they differ, do not blindly copy the write path; for gophertunnel decode correctness, the read path and live payloads are usually more important.
- Treat Mojang docs as high-value for field order and logical types, but not infallible. Confirm with at least one implementation or live bytes for risky changes.
- Treat PMMP as strong independent evidence when it matches Mojang docs or live bytes, even if Cloudburst differs.
- For `map<K,V>` schemas, check the byte layout. A gophertunnel slice is fine if each entry serializes `key` then `value` in map order.
- For optional fields, preserve exact ordinal order. Missing an optional marker shifts all later fields.
- For `Compression` on enum/integer fields, the logical Go type may remain small (`byte`, enum) while the wire encoding uses varint/varuint.
- Distinguish actor unique IDs from runtime IDs:
  - actor unique ID: signed `varint64` zigzag
  - runtime entity ID: unsigned `varuint64`
  Use live bytes to resolve ambiguity.
- PrimitiveShapes `PrimitiveShapeDataPayload.Attached To Entity ID` is a runtime actor ID on the wire. Mojang
  `bedrock-protocol-docs` has historically described it as runtime ID while linking the schema node to
  `ActorUniqueID`; trust the runtime-ID wording plus independent implementation/BDS evidence here. PMMP encodes
  this field with `getActorRuntimeId`/`putActorRuntimeId`, LeviLamina names the field
  `std::optional<ActorRuntimeID> mAttachedToId`, and live BDS ignored signed/unique-style encoding while accepting
  unsigned runtime encoding. In gophertunnel, model it as `Optional[uint64]` and marshal with `Varuint64`, not
  signed `Varint64`.
- For `oneOf`/variant selectors, encode the selector as documented, usually varuint, not as a bool unless the source explicitly says bool.
- For enum sentinels like `UNDEFINED`, check whether they existed historically at changing numeric indexes before calling them a new semantic value.
- For contiguous protocol constants, prefer the surrounding gophertunnel style. `iota` is fine for dense, ordered wire-value ranges when every value is consecutive and future additions can be inserted in order.
- Keep non-contiguous or out-of-band protocol values explicit. Use hex for flag-like/high-bit values such as `0x8000000`, and separate them from the contiguous `iota` block with a blank line.
- PMMP may keep compatibility branches that gophertunnel should not support. Prefer the target protocol's exact wire format unless the user explicitly asks for backwards compatibility.

## Implementation Workflow

1. Start with `git status --short --branch`; never revert unrelated user changes.
2. Gather references before editing. Capture commit SHAs or stable links for Mojang, Cloudburst, PMMP, and any LeviLamina/BDS evidence used.
3. Diff the current gophertunnel files against source evidence packet by packet.
4. Make the smallest code change that fixes the target protocol. Avoid broad refactors and generated churn.
5. Respect user-specific constraints. For Hashim's gophertunnel work, use normal follow-up commits, do not amend unless asked, do not add "committed by Codex", and do not add test files when the user says tests are unnecessary.
6. Run targeted verification first:
   - `go test ./minecraft/protocol`
   - `go test ./minecraft/protocol/packet`
   - `git diff --check`
   Add broader checks only when the change touches shared behavior.
7. If Lunar or another downstream project consumes a pseudo-version branch, verify which branch/commit its `go.mod` points at before claiming live testing will include the new fix.
8. Push the requested branch only after tests pass and status is understood.

## Debugging Decode Errors

- Remember that a "remaining bytes" dump is often the unread tail after partial decode, not the full packet.
- When a packet leaves a large unread tail, first look for a missing field, missing optional marker, or wrong variant selector before assuming a nested type is corrupt.
- For repeated binary structures, compare field order from Mojang `x-ordinal-index`, PMMP read/write order, and Cloudburst read order.
- Add temporary packet logging only in the downstream/debug repo unless the user asks to keep it. Do not commit temporary live-capture instrumentation by default.

## PR Output

When asked for a PR description, include:

- concise summary of packet/codecs changed
- exact source links per change
- test commands run
- known evidence conflicts, if any

Use direct GitHub links with commit SHAs where possible.
