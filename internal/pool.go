package internal

import (
	"bytes"
	"github.com/klauspost/compress/flate"
	"io/ioutil"
	"sync"
)

// DecompressPool is a sync.Pool for io.ReadCloser flate readers. These are pooled for connections.
var DecompressPool = sync.Pool{
	New: func() interface{} {
		return flate.NewReader(bytes.NewReader(nil))
	},
}

// CompressPool is a sync.Pool for writeCloseResetter flate readers. These are pooled for connections.
var CompressPool = sync.Pool{
	New: func() interface{} {
		w, _ := flate.NewWriter(ioutil.Discard, 6)
		return w
	},
}

// BufferPool is a sync.Pool for buffers used to write compressed data to.
var BufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 256))
	},
}
