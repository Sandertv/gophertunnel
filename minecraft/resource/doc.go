// Package resource implements the compiling of resource packs found in files, directories or raw byte data.
// It ensures the data in the resource pack is valid (for example, it checks if the manifest is present and
// holds correct data) and extracts information which may be obtained by calling the exported methods of a
// *resource.Pack.
package resource
