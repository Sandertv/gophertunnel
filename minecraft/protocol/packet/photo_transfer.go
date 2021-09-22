package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PhotoTypePortfolio uint8 = iota
	PhotoTypePhotoItem
	PhotoTypeBook
)

// PhotoTransfer is sent by the server to transfer a photo (image) file to the client. It is typically used
// to transfer photos so that the client can display it in a portfolio in Education Edition.
// While previously usable in the default Bedrock Edition, the displaying of photos in books was disabled and
// the packet now has little use anymore.
type PhotoTransfer struct {
	// PhotoName is the name of the photo to transfer. It is the exact file name that the client will download
	// the photo as, including the extension of the file.
	PhotoName string
	// PhotoData is the raw data of the photo image. The format of this data may vary: Formats such as JPEG or
	// PNG work, as long as PhotoName has the correct extension.
	PhotoData []byte
	// BookID is the ID of the book that the photo is associated with. If the PhotoName in a book with this ID
	// is set to PhotoName, it will display the photo (provided Education Edition is used).
	// The photo image is downloaded to a sub-folder with this book ID.
	BookID string
	// PhotoType is one of the three photo types above.
	PhotoType byte
	// SourceType is the source photo type. It is one of the three photo types above.
	SourceType byte
	// OwnerEntityUniqueID is the entity unique ID of the photo's owner.
	OwnerEntityUniqueID int64
	// NewPhotoName is the new name of the photo.
	NewPhotoName string
}

// ID ...
func (*PhotoTransfer) ID() uint32 {
	return IDPhotoTransfer
}

// Marshal ...
func (pk *PhotoTransfer) Marshal(w *protocol.Writer) {
	w.String(&pk.PhotoName)
	w.ByteSlice(&pk.PhotoData)
	w.String(&pk.BookID)
	w.Uint8(&pk.PhotoType)
	w.Uint8(&pk.SourceType)
	w.Int64(&pk.OwnerEntityUniqueID)
	w.String(&pk.NewPhotoName)
}

// Unmarshal ...
func (pk *PhotoTransfer) Unmarshal(r *protocol.Reader) {
	r.String(&pk.PhotoName)
	r.ByteSlice(&pk.PhotoData)
	r.String(&pk.BookID)
	r.Uint8(&pk.PhotoType)
	r.Uint8(&pk.SourceType)
	r.Int64(&pk.OwnerEntityUniqueID)
	r.String(&pk.NewPhotoName)
}
