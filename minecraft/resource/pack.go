package resource

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"github.com/muhammadmuzzammil1998/jsonc"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Pack is a container of a resource pack parsed from a directory or a .zip archive (or .mcpack). It holds
// methods that may be used to get information about the resource pack.
type Pack struct {
	// manifest is the manifest of the resource pack. It contains information about the pack such as the name,
	// version and description.
	manifest *Manifest

	// downloadURL is the URL that the resource pack can be downloaded from. If the string is empty, then the
	// resource pack will be downloaded over RakNet rather than HTTP.
	downloadURL string
	// content is a bytes.Reader that contains the full content of the zip file. It is used to send the full
	// data to a client.
	content *bytes.Reader
	// contentKey is the key used to encrypt the files. The client uses this to decrypt the resource pack if encrypted.
	// If nothing is encrypted, this field can be left as an empty string.
	contentKey string

	// checksum is the SHA256 checksum of the full content of the file. It is sent to the client so that it
	// can 'verify' the download.
	checksum [32]byte
}

// ReadPath compiles a resource pack found at the path passed. The resource pack must either be a zip archive
// (extension does not matter, could be .zip or .mcpack), or a directory containing a resource pack. In the
// case of a directory, the directory is compiled into an archive and the pack is parsed from that.
// ReadPath operates assuming the resource pack has a 'manifest.json' file in it. If it does not, the function
// will fail and return an error.
func ReadPath(path string) (*Pack, error) {
	return compile(path)
}

// ReadURL downloads a resource pack found at the URL passed and compiles it. The resource pack must be a valid
// zip archive where the manifest.json file is inside a subdirectory rather than the root itself. If the resource
// pack is not a valid zip or there is no manifest.json file, an error is returned.
func ReadURL(url string) (*Pack, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download resource pack: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download resource pack: %v (%d)", resp.Status, resp.StatusCode)
	}
	pack, err := Read(resp.Body)
	if err != nil {
		return nil, err
	}
	pack.downloadURL = url
	return pack, nil
}

// MustReadPath compiles a resource pack found at the path passed. The resource pack must either be a zip
// archive (extension does not matter, could be .zip or .mcpack), or a directory containing a resource pack.
// In the case of a directory, the directory is compiled into an archive and the pack is parsed from that.
// ReadPath operates assuming the resource pack has a 'manifest.json' file in it. If it does not, the function
// will fail and return an error.
// Unlike ReadPath, MustReadPath does not return an error and panics if an error occurs instead.
func MustReadPath(path string) *Pack {
	pack, err := compile(path)
	if err != nil {
		panic(err)
	}
	return pack
}

// MustReadURL downloads a resource pack found at the URL passed and compiles it. The resource pack must be a valid
// zip archive where the manifest.json file is inside a subdirectory rather than the root itself. If the resource
// pack is not a valid zip or there is no manifest.json file, an error is returned.
// Unlike ReadURL, MustReadURL does not return an error and panics if an error occurs instead.
func MustReadURL(url string) *Pack {
	pack, err := ReadURL(url)
	if err != nil {
		panic(err)
	}
	return pack
}

// Read parses an archived resource pack written to a raw byte slice passed. The data must be a valid
// zip archive and contain a pack manifest in order for the function to succeed.
// Read saves the data to a temporary archive.
func Read(r io.Reader) (*Pack, error) {
	temp, err := createTempFile()
	if err != nil {
		return nil, fmt.Errorf("create temp zip archive: %w", err)
	}
	_, _ = io.Copy(temp, r)
	if err := temp.Close(); err != nil {
		return nil, fmt.Errorf("close temp zip archive: %w", err)
	}
	pack, parseErr := ReadPath(temp.Name())
	if err := os.Remove(temp.Name()); err != nil {
		return nil, fmt.Errorf("remove temp zip archive: %w", err)
	}
	return pack, parseErr
}

// Name returns the name of the resource pack.
func (pack *Pack) Name() string {
	return pack.manifest.Header.Name
}

// UUID returns the UUID of the resource pack.
func (pack *Pack) UUID() uuid.UUID {
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

// DownloadURL returns the URL that the resource pack can be downloaded from. If the string is empty, then the
// resource pack will be downloaded over RakNet rather than HTTP.
func (pack *Pack) DownloadURL() string {
	return pack.downloadURL
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

// Encrypted returns if the resource pack has been encrypted with a content key or not.
func (pack *Pack) Encrypted() bool {
	return pack.contentKey != ""
}

// ContentKey returns the encryption key used to encrypt the resource pack. If the pack is not encrypted then
// this can be empty.
func (pack *Pack) ContentKey() string {
	return pack.contentKey
}

// ReadAt reads len(b) bytes from the resource pack's archive data at offset off and copies it into b. The
// amount of bytes read n is returned.
func (pack *Pack) ReadAt(b []byte, off int64) (n int, err error) {
	return pack.content.ReadAt(b, off)
}

// WithContentKey creates a copy of the pack and sets the encryption key to the key provided, after which the
// new Pack is returned.
func (pack Pack) WithContentKey(key string) *Pack {
	pack.contentKey = key
	return &pack
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
		return nil, fmt.Errorf("open resource pack path: %w", err)
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
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	// Then we read the entire content of the zip archive into a byte slice and compute the SHA256 checksum
	// and a reader.
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read resource pack file content: %w", err)
	}
	checksum := sha256.Sum256(content)
	contentReader := bytes.NewReader(content)

	return &Pack{manifest: manifest, checksum: checksum, content: contentReader}, nil
}

// createTempArchive creates a zip archive from the files in the path passed and writes it to a temporary
// file, which is returned when successful.
func createTempArchive(path string) (*os.File, error) {
	temp, err := createTempFile()
	if err != nil {
		return nil, err
	}
	writer := zip.NewWriter(temp)
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return fmt.Errorf("find relative path: %w", err)
		}
		// Make sure to replace backslashes with forward slashes as Go zip only allows that.
		relPath = strings.Replace(relPath, `\`, "/", -1)
		// Always ignore '.' as it is not a real file/folder.
		if relPath == "." {
			return nil
		}
		s, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("read stat of file path %v: %w", filePath, err)
		}
		if s.IsDir() {
			// This is a directory: Go zip requires you add forward slashes at the end to create directories.
			_, _ = writer.Create(relPath + "/")
			return nil
		}
		f, err := writer.Create(relPath)
		if err != nil {
			return fmt.Errorf("create new zip file: %w", err)
		}
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open resource pack file %v: %w", filePath, err)
		}
		data, _ := io.ReadAll(file)
		// Write the original content into the 'zip file' so that we write compressed data to the file.
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write file data to zip: %w", err)
		}
		_ = file.Close()
		return nil
	}); err != nil {
		return nil, fmt.Errorf("build zip archive: %w", err)
	}
	_ = writer.Close()
	return temp, nil
}

// createTempFile attempts to create a temporary file and returns it.
func createTempFile() (*os.File, error) {
	// We've got a directory which we need to load. Provided we need to send compressed zip data to the
	// client, we compile it to a zip archive in a temporary file.

	// Note that we explicitly do not handle the error here. If the user config
	// dir cannot be found, 'dir' will be an empty string. os.CreateTemp will
	// then use the default temporary file directory, which might succeed in
	// this case.
	dir, _ := os.UserConfigDir()
	_ = os.MkdirAll(dir, os.ModePerm)

	temp, err := os.CreateTemp(dir, "temp_resource_pack-*.mcpack")
	if err != nil {
		return nil, fmt.Errorf("create temp resource pack file: %w", err)
	}
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
			return nil, fmt.Errorf("open zip file %v: %w", file.Name, err)
		}
		return fileReader, nil
	}
	return nil, fmt.Errorf("'%v' not found in zip", fileName)
}

// readManifest reads the manifest from the resource pack located at the path passed. If not found in the root
// of the resource pack, it will also attempt to find it deeper down into the archive.
func readManifest(path string) (*Manifest, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip reader: %w", err)
	}
	reader := packReader{ReadCloser: r}
	defer func() {
		_ = r.Close()
	}()

	// Try to find the manifest file in the zip.
	manifestFile, err := reader.find("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("load manifest: %w", err)
	}
	defer func() {
		_ = manifestFile.Close()
	}()

	// Read all data from the manifest file so that we can decode it into a Manifest struct.
	allData, err := io.ReadAll(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("read manifest file: %w", err)
	}
	manifest := &Manifest{}
	if err := jsonc.Unmarshal(allData, manifest); err != nil {
		return nil, fmt.Errorf("decode manifest JSON: %w (data: %v)", err, string(allData))
	}

	if _, err := reader.find("level.dat"); err == nil {
		manifest.worldTemplate = true
	}

	return manifest, nil
}
