package resource

import "github.com/google/uuid"

// Documentation on this may be found here:
// https://learn.microsoft.com/en-us/minecraft/creator/reference/content/addonsreference/examples/addonmanifest

// Manifest contains all the basic information about the pack that Minecraft needs to identify it.
type Manifest struct {
	// FormatVersion defines the current version of the manifest. This is currently always 2.
	FormatVersion int `json:"format_version"`
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

	// worldTemplate holds a value indicating if the pack holds an entire world template or not.
	worldTemplate bool
}

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
	Version [3]int `json:"version"`
	// MinimumGameVersion is the minimum version of the game that this resource pack was written for.
	MinimumGameVersion [3]int `json:"min_engine_version"`
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
	Version [3]int `json:"version"`
}

// Dependency describes a pack that this pack depends on in order to work.
type Dependency struct {
	// UUID is the unique identifier of the pack that this pack depends on. It needs to be the exact same UUID
	// that the pack has defined in the header section of it's manifest file.
	UUID string `json:"uuid"`
	// Version is the specific version of the pack that the pack depends on. Should match the version the
	// other pack has in its manifest file.
	Version [3]int `json:"version"`
}

// Capability is a particular feature that the pack utilises of that isn't necessarily enabled by default.
//
//	experimental_custom_ui: Allows HTML files in the pack to be used for custom UI, and scripts in the pack
//	                        to call and manipulate custom UI.
//	chemistry:              Allows the pack to add, change or replace Chemistry functionality.
type Capability string

// Metadata contains additional information about the pack that is otherwise optional.
type Metadata struct {
	// Author is the name of the author(s) of the pack.
	Author string `json:"authors,omitempty"`
	// License is the license applied to the pack.
	License string `json:"license,omitempty"`
	// URL is the home website of the creator of the pack.
	URL string `json:"url,omitempty"`
}
