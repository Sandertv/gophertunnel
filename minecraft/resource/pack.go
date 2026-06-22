package resource

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/df-mc/jsonc"
	"github.com/google/uuid"
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
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("open resource pack path: %w", err)
	}
	if info.IsDir() {
		return compileDir(path)
	}
	return compileZipPath(path)
}

func ReadBytes(data []byte) (*Pack, error) {
	return compile(data, true)
}

// ReadURL downloads a resource pack found at the URL passed and compiles it. The resource pack must be a valid
// zip archive where the manifest.json file is inside a subdirectory rather than the root itself. If the resource
// pack is not a valid zip or there is no manifest.json file, an error is returned.
func ReadURL(url string) (*Pack, error) {
	return ReadURLContext(context.Background(), url)
}

// ReadURLContext downloads a resource pack found at the URL passed and compiles it. The request is canceled
// when ctx is done.
func ReadURLContext(ctx context.Context, url string) (*Pack, error) {
	return readURLContext(ctx, url, 0)
}

// ReadURLContextLimit downloads a resource pack found at the URL passed and compiles it, reading at most maxSize
// bytes from the response body. The request is canceled when ctx is done.
func ReadURLContextLimit(ctx context.Context, url string, maxSize uint64) (*Pack, error) {
	if maxSize == 0 {
		return nil, errors.New("download resource pack: max size must be greater than 0")
	}
	if maxSize > math.MaxInt64 {
		return nil, fmt.Errorf("download resource pack: max size %d exceeds supported limit", maxSize)
	}
	return readURLContext(ctx, url, int64(maxSize))
}

func readURLContext(ctx context.Context, url string, maxSize int64) (*Pack, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create resource pack request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download resource pack: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download resource pack: %v (%d)", resp.Status, resp.StatusCode)
	}
	var r io.Reader = resp.Body
	if maxSize > 0 {
		if resp.ContentLength > maxSize {
			return nil, fmt.Errorf("download resource pack: response size %d exceeds limit %d", resp.ContentLength, maxSize)
		}
		r = io.LimitReader(resp.Body, maxSize+1)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read resource pack: %w", err)
	}
	if maxSize > 0 && int64(len(data)) > maxSize {
		return nil, fmt.Errorf("download resource pack: response size exceeds limit %d", maxSize)
	}
	pack, err := compile(data, false)
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
	pack, err := ReadPath(path)
	if err != nil {
		panic(err)
	}
	return pack
}

func MustReadBytes(data []byte) *Pack {
	pack, err := ReadBytes(data)
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
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read resource pack: %w", err)
	}
	return compile(data, true)
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

// Size returns the total size in bytes of the archive of the resource pack.
func (pack *Pack) Size() int {
	return int(pack.content.Size())
}

// Len returns the total size in bytes of the archive of the resource pack.
// Deprecated: Use Size instead. Kept for backwards compatibility with older integrations.
func (pack *Pack) Len() int {
	return pack.Size()
}

// DataChunkCount returns the amount of chunks the data of the resource pack is split into if each chunk has
// a specific length.
func (pack *Pack) DataChunkCount(length int) int {
	if length <= 0 {
		return 0
	}
	packSize := pack.Size()
	count := packSize / length
	if packSize%length != 0 {
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

// ReadFile reads a specific file from the Pack's content and returns its content as a byte slice.
func (p *Pack) ReadFile(filePath string) ([]byte, error) {
	// Create a new zip reader from the content of the bytes.Reader
	zipReader, err := zip.NewReader(p.content, int64(p.content.Size()))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Iterate over the files in the archive to find the file
	for _, file := range zipReader.File {
		// Check if the current file is the one we're looking for
		if strings.EqualFold(file.Name, filePath) {
			// Open the file
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			defer rc.Close()

			// Read the file content
			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read file content: %w", err)
			}

			return content, nil
		}
	}

	return nil, fmt.Errorf("file %s not found in the resource pack", filePath)
}

// WithContentKey creates a copy of the pack and sets the encryption key to the key provided, after which the
// new Pack is returned.
func (pack Pack) WithContentKey(key string) *Pack {
	pack.contentKey = key
	return &pack
}

// WithDownloadURL creates a copy of the pack and sets the HTTP download URL
// used in ResourcePacksInfo.
func (pack Pack) WithDownloadURL(url string) *Pack {
	pack.downloadURL = url
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

// compileDir compiles a resource pack from a directory.
func compileDir(root string) (*Pack, error) {
	pr := dirPackReader{base: root}
	m, err := readManifest(pr)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := createArchive(buf, root); err != nil {
		return nil, fmt.Errorf("create zip: %w", err)
	}

	data := buf.Bytes()
	return &Pack{
		manifest: m,
		checksum: sha256.Sum256(data),
		content:  bytes.NewReader(data),
	}, nil
}

// compileZipPath compiles a resource pack from a zip file.
func compileZipPath(p string) (*Pack, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read resource pack file: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	m, err := readManifest(zipPackReader{zr})
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	return &Pack{
		manifest: m,
		checksum: sha256.Sum256(data),
		content:  bytes.NewReader(data),
	}, nil
}

// compile compiles the resource pack from the bytes passed, either a zip archive or a directory, and returns a
// resource pack if successful.
func compile(data []byte, unwrapNested bool) (*Pack, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	// Check if this is a nested zip (only contains a single .zip file)
	// We only unwrap one level to avoid recursive nesting issues
	if unwrapNested && len(zr.File) == 1 && strings.HasSuffix(strings.ToLower(zr.File[0].Name), ".zip") {
		nestedFile, err := zr.File[0].Open()
		if err != nil {
			return nil, fmt.Errorf("open nested zip %s: %w", zr.File[0].Name, err)
		}
		defer nestedFile.Close()

		nestedData, err := io.ReadAll(nestedFile)
		if err != nil {
			return nil, fmt.Errorf("read nested zip %s: %w", zr.File[0].Name, err)
		}

		// Replace data with the unwrapped nested zip and re-open it
		data = nestedData
		zr, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, fmt.Errorf("open unwrapped zip: %w", err)
		}
	}

	pr := zipPackReader{zr}

	// Read the manifest to ensure that it exists and is valid.
	manifest, err := readManifest(pr)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	// Compute the SHA256 checksum and create a reader for the content
	return &Pack{
		manifest: manifest,
		checksum: sha256.Sum256(data),
		content:  bytes.NewReader(data),
	}, nil
}

// createArchive creates a zip archive from the files in the path passed and writes
// to the writer passed.
func createArchive(w io.Writer, path string) error {
	writer := zip.NewWriter(w)

	if err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		rel, err := filepath.Rel(path, filePath)
		if err != nil {
			return fmt.Errorf("find relative path: %w", err)
		}
		if rel == "." {
			return nil
		}

		// Make sure to replace backslashes with forward slashes as Go zip only allows that.
		rel = strings.ReplaceAll(rel, `\`, `/`)

		if d.IsDir() {
			// This is a directory: Go zip requires you add forward slashes at the end to create directories.
			_, _ = writer.Create(rel + "/")
			return nil
		}

		wr, err := writer.Create(rel)
		if err != nil {
			return fmt.Errorf("create zip entry: %w", err)
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open resource pack file %q: %w", filePath, err)
		}

		if _, err := io.Copy(wr, file); err != nil {
			_ = file.Close()
			return fmt.Errorf("copy %q: %w", filePath, err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("close resource pack file %q: %w", filePath, err)
		}

		return nil
	}); err != nil {
		_ = writer.Close()
		return fmt.Errorf("build zip archive: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close zip writer: %w", err)
	}

	return nil
}

type packReader interface {
	find(fileName string) (io.ReadCloser, error)
}

// zipPackReader wraps around a zip.Reader to provide file finding functionality.
type zipPackReader struct {
	*zip.Reader
}

// dirPackReader walks the directory tree rooted at base.
type dirPackReader struct {
	base string
}

// find attempts to find a file in a zip reader. If found, it returns an Open()ed reader of the file that may
// be used to read data from the file.
func (r zipPackReader) find(fileName string) (io.ReadCloser, error) {
	for _, f := range r.File {
		if filepath.Base(f.Name) != fileName {
			continue
		}
		fileReader, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip file %v: %w", f.Name, err)
		}
		return fileReader, nil
	}
	return nil, fmt.Errorf("'%v' not found in zip", fileName)
}

func (r dirPackReader) find(fileName string) (io.ReadCloser, error) {
	p := filepath.Join(r.base, fileName)
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("'%v' not found in directory", fileName)
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("'%v' is a directory, not a file", fileName)
	}
	return os.Open(p)
}

// readManifest reads the manifest from the resource pack located at the path passed. If not found in the root
// of the resource pack, it will also attempt to find it deeper down into the archive.
func readManifest(pr packReader) (*Manifest, error) {
	// Try to find the manifest file in the zip.
	manifestFile, err := pr.find("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("load manifest: %w", err)
	}
	defer manifestFile.Close()

	// Read all data from the manifest file so that we can decode it into a Manifest struct.
	allData, err := io.ReadAll(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("read manifest file: %w", err)
	}
	manifest := &Manifest{}
	if err := jsonc.UnmarshalLenient(allData, manifest); err != nil {
		return nil, fmt.Errorf("decode manifest JSON: %w (data: %v)", err, string(allData))
	}

	if rc, err := pr.find("level.dat"); err == nil {
		_ = rc.Close()
		manifest.worldTemplate = true
	}

	return manifest, nil
}
