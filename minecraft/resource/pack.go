package resource

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Pack is a container of a resource pack parsed from a directory or a .zip archive (or .mcpack). It holds
// methods that may be used to get information about the resource pack.
type Pack struct {
	// manifest is the manifest of the resource pack. It contains information about the pack such as the name,
	// version and description.
	manifest *Manifest

	// content is a bytes.Reader that contains the full content of the zip file. It is used to send the full
	// data to a client.
	content *bytes.Reader

	// checksum is the SHA256 checksum of the full content of the file. It is sent to the client so that it
	// can 'verify' the download.
	checksum [32]byte
}

// Compile compiles a resource pack found at the path passed. The resource pack must either be a zip archive
// (extension does not matter, could be .zip or .mcpack), or a directory containing a resource pack. In the
// case of a directory, the directory is compiled into an archive and the pack is parsed from that.
// Compile operates assuming the resource pack has a 'manifest.json' file in it. If it does not, the function
// will fail and return an error.
func Compile(path string) (*Pack, error) {
	return compile(path)
}

// MustCompile compiles a resource pack found at the path passed. The resource pack must either be a zip
// archive (extension does not matter, could be .zip or .mcpack), or a directory containing a resource pack.
// In the case of a directory, the directory is compiled into an archive and the pack is parsed from that.
// Compile operates assuming the resource pack has a 'manifest.json' file in it. If it does not, the function
// will fail and return an error.
// Unlike Compile, MustCompile does not return an error and panics if an error occurs instead.
func MustCompile(path string) *Pack {
	pack, err := compile(path)
	if err != nil {
		panic(err)
	}
	return pack
}

// FromBytes parses an archived resource pack written to a raw byte slice passed. The data must be a valid
// zip archive and contain a pack manifest in order for the function to succeed.
// FromBytes saves the data to a temporary archive.
func FromBytes(data []byte) (*Pack, error) {
	tempFile, err := ioutil.TempFile("", "resource_pack_archive-*.mcpack")
	if err != nil {
		return nil, fmt.Errorf("error creating temp zip archive: %v", err)
	}
	_, _ = tempFile.Write(data)
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing temp zip archive: %v", err)
	}
	pack, parseErr := Compile(tempFile.Name())
	if err := os.Remove(tempFile.Name()); err != nil {
		return nil, fmt.Errorf("error removing temp zip archive: %v", err)
	}
	return pack, parseErr
}

// Name returns the name of the resource pack.
func (pack *Pack) Name() string {
	return pack.manifest.Header.Name
}

// UUID returns the UUID of the resource pack.
func (pack *Pack) UUID() string {
	return pack.manifest.Header.UUID
}

// Description returns the description of the resource pack.
func (pack *Pack) Description() string {
	return pack.manifest.Header.Description
}

// Version returns the string version of the resource pack. It is guaranteed to have 3 digits in it, joined
// by a dot.
func (pack *Pack) Version() string {
	return strconv.Itoa(pack.manifest.Header.Version[0]) + "." + strconv.Itoa(pack.manifest.Header.Version[1]) + "." + strconv.Itoa(pack.manifest.Header.Version[2])
}

// Modules returns all modules that the resource pack exists out of. Resource packs usually have only one
// module, but may have more depending on their functionality.
func (pack *Pack) Modules() []Module {
	return pack.manifest.Modules
}

// Dependencies returns all dependency resource packs that must be loaded in order for this resource pack to
// function correctly.
func (pack *Pack) Dependencies() []Dependency {
	return pack.manifest.Dependencies
}

// HasScripts checks if any of the modules of the resource pack have the type 'client_data', meaning they have
// scripts in them.
func (pack *Pack) HasScripts() bool {
	for _, module := range pack.manifest.Modules {
		if module.Type == "client_data" {
			// The module has the client_data type, meaning it holds client scripts.
			return true
		}
	}
	return false
}

// HasBehaviours checks if any of the modules of the resource pack have either the type 'data' or
// 'client_data', meaning they contain behaviours (or scripts).
func (pack *Pack) HasBehaviours() bool {
	for _, module := range pack.manifest.Modules {
		if module.Type == "client_data" || module.Type == "data" {
			// The module has the client_data or data type, meaning it holds behaviours.
			return true
		}
	}
	return false
}

// HasTextures checks if any of the modules of the resource pack have the type 'resources', meaning they have
// textures in them.
func (pack *Pack) HasTextures() bool {
	for _, module := range pack.manifest.Modules {
		if module.Type == "resources" {
			// The module has the resources type, meaning it holds textures.
			return true
		}
	}
	return false
}

// HasWorldTemplate checks if the resource compiled holds a level.dat in it, indicating that the resource is
// a world template.
func (pack *Pack) HasWorldTemplate() bool {
	return pack.manifest.worldTemplate
}

// Checksum returns the SHA256 checksum made from the full, compressed content of the resource pack archive.
// It is transmitted as a string over network.
func (pack *Pack) Checksum() [32]byte {
	return pack.checksum
}

// Len returns the total length in bytes of the content of the archive that contained the resource pack.
func (pack *Pack) Len() int {
	return pack.content.Len()
}

// DataChunkCount returns the amount of chunks the data of the resource pack is split into if each chunk has
// a specific length.
func (pack *Pack) DataChunkCount(length int) int {
	count := pack.Len() / length
	if pack.Len()%length != 0 {
		count++
	}
	return count
}

// ReadAt reads len(b) bytes from the resource pack's archive data at offset off and copies it into b. The
// amount of bytes read n is returned.
func (pack *Pack) ReadAt(b []byte, off int64) (n int, err error) {
	return pack.content.ReadAt(b, off)
}

// Manifest returns the manifest found in the manifest.json of the resource pack. It contains information
// about the pack such as its name.
func (pack *Pack) Manifest() Manifest {
	return *pack.manifest
}

// String returns a readable representation of the resource pack. It implements the Stringer interface.
func (pack *Pack) String() string {
	return fmt.Sprintf("%v v%v (%v): %v", pack.Name(), pack.Version(), pack.UUID(), pack.Description())
}

// compile compiles the resource pack found in path, either a zip archive or a directory, and returns a
// resource pack if successful.
func compile(path string) (*Pack, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error opening resource pack path: %v", err)
	}
	if info.IsDir() {
		temp, err := createTempArchive(path)
		if err != nil {
			return nil, err
		}
		// We set the path to the temp zip archive we just made.
		path = temp.Name()

		// Make sure we close the temp file and remove it at the end. We don't need to keep it, as we read all
		// the content in a byte slice.
		_ = temp.Close()
		defer func() {
			_ = os.Remove(temp.Name())
		}()
	}
	// First we read the manifest to ensure that it exists and is valid.
	manifest, err := readManifest(path)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest: %v", err)
	}

	// Then we read the entire content of the zip archive into a byte slice and compute the SHA256 checksum
	// and a reader.
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading resource pack file content: %v", err)
	}
	checksum := sha256.Sum256(content)
	contentReader := bytes.NewReader(content)

	return &Pack{manifest: manifest, checksum: checksum, content: contentReader}, nil
}

// createTempArchive creates a zip archive from the files in the path passed and writes it to a temporary
// file, which is returned when successful.
func createTempArchive(path string) (*os.File, error) {
	// We've got a directory which we need to load. Provided we need to send compressed zip data to the
	// client, we compile it to a zip archive in a temporary file.
	temp, err := ioutil.TempFile("", "resource_pack-*.mcpack")
	if err != nil {
		return nil, fmt.Errorf("error creating temp zip file: %v", err)
	}
	writer := zip.NewWriter(temp)
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return fmt.Errorf("error finding relative path: %v", err)
		}
		// Make sure to replace backslashes with forward slashes as Go zip only allows that.
		relPath = strings.Replace(relPath, `\`, "/", -1)
		// Always ignore '.' as it is not a real file/folder.
		if relPath == "." {
			return nil
		}
		s, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("error getting stat of file path %v: %w", filePath, err)
		}
		if s.IsDir() {
			// This is a directory: Go zip requires you add forward slashes at the end to create directories.
			_, _ = writer.Create(relPath + "/")
			return nil
		}
		f, err := writer.Create(relPath)
		if err != nil {
			return fmt.Errorf("error creating new zip file: %v", err)
		}
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error opening resource pack file %v: %v", filePath, err)
		}
		data, _ := ioutil.ReadAll(file)
		// Write the original content into the 'zip file' so that we write compressed data to the file.
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("error writing file data to zip: %v", err)
		}
		_ = file.Close()
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error building zip archive: %v", err)
	}
	_ = writer.Close()
	return temp, nil
}

// packReader wraps around a zip.Reader to provide file finding functionality.
type packReader struct {
	*zip.ReadCloser
}

// find attempts to find a file in a zip reader. If found, it returns an Open()ed reader of the file that may
// be used to read data from the file.
func (reader packReader) find(fileName string) (io.ReadCloser, error) {
	for _, file := range reader.File {
		base := filepath.Base(file.Name)
		if file.Name != fileName && base != fileName {
			continue
		}
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("error opening zip file %v: %v", file.Name, err)
		}
		return fileReader, nil
	}
	return nil, fmt.Errorf("could not find '%v' in zip", fileName)
}

// readManifest reads the manifest from the resource pack located at the path passed. If not found in the root
// of the resource pack, it will also attempt to find it deeper down into the archive.
func readManifest(path string) (*Manifest, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("error opening zip reader: %v", err)
	}
	reader := packReader{ReadCloser: r}
	defer func() {
		_ = r.Close()
	}()

	// Try to find the manifest file in the zip.
	manifestFile, err := reader.find("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("error loading manifest: %v", err)
	}
	defer func() {
		_ = manifestFile.Close()
	}()

	// Read all data from the manifest file so that we can decode it into a Manifest struct.
	allData, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("error reading from manifest file: %v", err)
	}
	// Some JSON implementations (Mojang's) allow comments in JSON. We strip these out first.
	expr := regexp.MustCompile(`//.*`)
	allData = expr.ReplaceAll(allData, []byte{})

	manifest := &Manifest{}
	if err := json.Unmarshal(allData, manifest); err != nil {
		return nil, fmt.Errorf("error decoding manifest JSON: %v (data: %v)", err, string(allData))
	}
	manifest.Header.UUID = strings.ToLower(manifest.Header.UUID)

	if _, err := reader.find("level.dat"); err == nil {
		manifest.worldTemplate = true
	}

	return manifest, nil
}
