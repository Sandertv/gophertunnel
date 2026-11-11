package resource

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Documentation on this may be found here:
// https://learn.microsoft.com/en-us/minecraft/creator/reference/content/addonsreference/examples/addonmanifest

// Manifest contains all the basic information about the pack that Minecraft needs to identify it.
type Manifest struct {
	// FormatVersion defines the current version of the manifest, usually 1, 2 or 3.
	FormatVersion FormatVersion `json:"format_version"`
	// Header is the header of a resource pack. It contains information that applies to the entire resource
	// pack, such as the name of the resource pack.
	Header Header `json:"header"`
	// Modules describes the modules that comprise the pack. Each entry here defines one of the kinds of
	// contents of the pack.
	Modules []Module `json:"modules"`
	// Dependencies describes the packs that this pack depends on in order to work.
	Dependencies []Dependency `json:"dependencies,omitempty"`
	// Capabilities are the different features that the pack makes use of that aren't necessarily enabled by
	// default. For a list of options, see below.
	Capabilities []Capability `json:"capabilities,omitempty"`
	// Metadata contains additional information about the pack that is otherwise optional.
	Metadata *Metadata `json:"metadata,omitempty"`

	// worldTemplate holds a value indicating if the pack holds an entire world template or not.
	worldTemplate bool
}

// Capability is a particular feature that the pack utilises of that isn't necessarily enabled by default.
//
//	experimental_custom_ui: Allows HTML files in the pack to be used for custom UI, and scripts in the pack
//	                        to call and manipulate custom UI.
//	chemistry:              Allows the pack to add, change or replace Chemistry functionality.
type Capability string

// Header is the header of a resource pack. It contains information that applies to the entire resource pack,
// such as the name of the resource pack.
type Header struct {
	// Name is the name of the pack as it appears within Minecraft.
	Name string `json:"name"`
	// Description is a short description of the pack. It will appear in the game below the name of the pack.
	Description string `json:"description"`
	// UUID is a unique identifier this pack from any other pack.
	UUID uuid.UUID `json:"uuid"`
	// Version is the version of the pack, which can be used to identify changes in the pack.
	Version Version `json:"version"`
	// MinimumGameVersion is the minimum version of the game that this resource pack was written for.
	MinimumGameVersion Version `json:"min_engine_version"`
}

// Module describes a module that comprises the pack. Each module defines one of the kinds of contents of the
// pack.
type Module struct {
	// UUID is a unique identifier for the module in the same format as the pack's UUID in the header. This
	// should be different from the pack's UUID, and different for every module.
	UUID string `json:"uuid"`
	// Description is a short description of the module. This is not user-facing at the moment.
	Description string `json:"description"`
	// Type is the type of the module. Can be any of the following: resources, data, client_data, interface or
	// world_template.
	Type string `json:"type"`
	// Version is the version of the module in the same format as the pack's version in the header. This can
	// be used to further identify changes in the pack.
	Version Version `json:"version"`
}

// Dependency describes a pack that this pack depends on in order to work.
type Dependency struct {
	// UUID is the unique identifier of the pack that this pack depends on. It needs to be the exact same UUID
	// that the pack has defined in the header section of it's manifest file.
	UUID string `json:"uuid"`
	// Version is the specific version of the pack that the pack depends on. Should match the version the
	// other pack has in its manifest file.
	Version Version `json:"version"`
}

// Metadata contains additional information about the pack that is otherwise optional.
type Metadata struct {
	// Author is the name of the author(s) of the pack.
	Authors []string `json:"authors,omitempty"`
	// License is the license applied to the pack.
	License string `json:"license,omitempty"`
	// URL is the home website of the creator of the pack.
	URL string `json:"url,omitempty"`
}

// FormatVersion is a manifest's format version as major.minor. Vanilla reads it as a semantic version
// written as a number, string or [major, minor, patch] array; an unreadable value decodes as 0 rather than
// failing the whole manifest, since nothing here depends on it.
type FormatVersion float64

// UnmarshalJSON accepts numeric, string and array manifest format versions.
func (f *FormatVersion) UnmarshalJSON(b []byte) error {
	*f = 0
	var number float64
	if err := json.Unmarshal(b, &number); err == nil {
		*f = FormatVersion(number)
		return nil
	}
	var v Version
	if err := v.UnmarshalJSON(b); err == nil {
		if parsed, err := strconv.ParseFloat(strconv.Itoa(v[0])+"."+strconv.Itoa(v[1]), 64); err == nil {
			*f = FormatVersion(parsed)
		}
	}
	return nil
}

// Version may be present in the manifest as [1,0,0] or "1.0.0".
type Version [3]int

// semVersion is vanilla's version grammar: missing minor and patch parts are zero, and pre-release and
// build suffixes follow a complete triple.
var semVersion = regexp.MustCompile(`^(0|[1-9]\d{0,4})(?:\.(0|[1-9]\d{0,4})(?:\.(0|[1-9]\d{0,4})(?:-(?:(?:(?:0[0-9]*[A-Za-z-])|[1-9A-Za-z-])[0-9A-Za-z-]*|0)(?:\.(?:(?:(?:0[0-9]*[A-Za-z-])|[1-9A-Za-z-])[0-9A-Za-z-]*|0))*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?)?)?$`)

// parseVersion reads a version string as vanilla does, also accepting the leading zeros older
// gophertunnel releases accepted. Pre-release and build suffixes are not part of the result.
func parseVersion(s string) (Version, bool) {
	s = strings.TrimSpace(s)
	var v Version
	if match := semVersion.FindStringSubmatch(s); match != nil {
		for i, part := range match[1:] {
			if part != "" {
				v[i], _ = strconv.Atoi(part)
			}
		}
		return v, true
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return v, false
	}
	for i, part := range parts {
		n, err := strconv.ParseUint(part, 10, 31)
		if err != nil {
			return v, false
		}
		v[i] = int(n)
	}
	return v, true
}

func (v *Version) UnmarshalJSON(b []byte) error {
	// Parse common array format [1,0,0]
	var arr [3]int
	if err := json.Unmarshal(b, &arr); err == nil {
		*v = arr
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		parsed, ok := parseVersion(s)
		if !ok {
			return fmt.Errorf("invalid version %q", s)
		}
		*v = parsed
		return nil
	}

	return fmt.Errorf("invalid version: %s", string(b))
}
