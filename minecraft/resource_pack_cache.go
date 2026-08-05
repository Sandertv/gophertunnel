package minecraft

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

// ResourcePackCacheKey identifies a resource pack advertised in the ResourcePacksInfo packet. Like the
// vanilla client's own pack cache, pack content is assumed not to change without a version bump; the
// advertised size is included as an extra guard.
type ResourcePackCacheKey struct {
	UUID    uuid.UUID
	Version string
	Size    uint64
}

// Matches reports whether pack has the UUID, version and size the key holds.
func (key ResourcePackCacheKey) Matches(pack *resource.Pack) bool {
	return pack.UUID() == key.UUID && pack.Version() == key.Version && uint64(pack.Len()) == key.Size
}

// ResourcePackCache allows a Dialer to reuse resource packs downloaded earlier. Cache failures are
// non-fatal: a nil pack or an error from Load falls back to a normal download, and errors from Store are
// only logged.
type ResourcePackCache interface {
	// Load returns the pack stored under key, or nil if it is not cached.
	Load(ctx context.Context, key ResourcePackCacheKey) (*resource.Pack, error)
	// Store stores a pack under key for a later Load.
	Store(ctx context.Context, key ResourcePackCacheKey, pack *resource.Pack) error
}

// DirResourcePackCache is a ResourcePackCache that stores resource packs as files in a directory. Entries
// are never evicted: the caller owns the directory and its lifecycle.
type DirResourcePackCache struct {
	// Dir is the directory packs are stored in. It is created when the first pack is stored.
	Dir string
}

// Load returns the pack stored under key, or nil if no file exists for it.
func (cache DirResourcePackCache) Load(_ context.Context, key ResourcePackCacheKey) (*resource.Pack, error) {
	pack, err := resource.ReadPath(cache.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return pack, err
}

// Store writes the pack to a file under key, replacing any previous entry.
func (cache DirResourcePackCache) Store(_ context.Context, key ResourcePackCacheKey, pack *resource.Pack) error {
	if err := os.MkdirAll(cache.Dir, 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(cache.Dir, "pack-*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()
	if _, err := io.Copy(temp, io.NewSectionReader(pack, 0, int64(pack.Len()))); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), cache.path(key))
}

// path returns the file a pack with key is stored at. The version is escaped as it comes from the server.
func (cache DirResourcePackCache) path(key ResourcePackCacheKey) string {
	return filepath.Join(cache.Dir, key.UUID.String()+"_"+url.PathEscape(key.Version)+".mcpack")
}
