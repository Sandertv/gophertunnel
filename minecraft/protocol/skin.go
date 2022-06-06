package protocol

import (
	"fmt"
)

// Skin represents the skin of a player as sent over network. The skin holds a texture and a model, and
// optional animations which may be present when the skin is created using persona or bought from the
// marketplace.
type Skin struct {
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string
	// PlayFabID is the PlayFab ID produced for the skin. PlayFab is the company that hosts the Marketplace,
	// skins and other related features from the game. This ID is the ID of the skin used to store the skin
	// inside of PlayFab.
	PlayFabID string
	// SkinResourcePatch is a JSON encoded object holding some fields that point to the geometry that the
	// skin has.
	// The JSON object that this holds specifies the way that the geometry of animations and the default skin
	// of the player are combined.
	SkinResourcePatch []byte
	// SkinImageWidth and SkinImageHeight hold the dimensions of the skin image. Note that these are not the
	// dimensions in bytes, but in pixels.
	SkinImageWidth, SkinImageHeight uint32
	// SkinData is a byte slice of SkinImageWidth * SkinImageHeight bytes. It is an RGBA ordered byte
	// representation of the skin pixels.
	SkinData []byte
	// Animations is a list of all animations that the skin has.
	Animations []SkinAnimation
	// CapeImageWidth and CapeImageHeight hold the dimensions of the cape image. Note that these are not the
	// dimensions in bytes, but in pixels.
	CapeImageWidth, CapeImageHeight uint32
	// CapeData is a byte slice of 64*32*4 bytes. It is a RGBA ordered byte representation of the cape
	// colours, much like the SkinData.
	CapeData []byte
	// SkinGeometry is a JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry []byte
	// TODO: Find out what value AnimationData holds and when it does hold something.
	AnimationData []byte
	// GeometryDataEngineVersion ...
	GeometryDataEngineVersion []byte
	// PremiumSkin specifies if this is a skin that was purchased from the marketplace.
	PremiumSkin bool
	// PersonaSkin specifies if this is a skin that was created using the in-game skin creator.
	PersonaSkin bool
	// PersonaCapeOnClassicSkin specifies if the skin had a Persona cape (in-game skin creator cape) equipped
	// on a classic skin.
	PersonaCapeOnClassicSkin bool
	// PrimaryUser ...
	PrimaryUser bool
	// CapeID is a unique identifier that identifies the cape. It usually holds a UUID in it.
	CapeID string
	// FullID is an ID that represents the skin in full. The actual functionality is unknown: The client
	// does not seem to send a value for this.
	FullID string
	// SkinColour is a hex representation (including #) of the base colour of the skin. An example of the
	// colour sent here is '#b37b62'.
	SkinColour string
	// ArmSize is the size of the arms of the player's model. This is either 'wide' (generally for male skins)
	// or 'slim' (generally for female skins).
	ArmSize string
	// PersonaPieces is a list of all persona pieces that the skin is composed of.
	PersonaPieces []PersonaPiece
	// PieceTintColours is a list of specific tint colours for (some of) the persona pieces found in the list
	// above.
	PieceTintColours []PersonaPieceTintColour
	// Trusted specifies if the skin is 'trusted'. No code should rely on this field, as any proxy or client
	// can easily change it.
	Trusted bool
}

// WriteSerialisedSkin writes a Skin x to Writer w. WriteSerialisedSkin panics if the fields of the skin
// have invalid values, usually indicating that the dimensions of the skin images are incorrect.
func WriteSerialisedSkin(w *Writer, x *Skin) {
	if err := x.validate(); err != nil {
		panic(err)
	}
	w.String(&x.SkinID)
	w.String(&x.PlayFabID)
	w.ByteSlice(&x.SkinResourcePatch)
	w.Uint32(&x.SkinImageWidth)
	w.Uint32(&x.SkinImageHeight)
	w.ByteSlice(&x.SkinData)
	l := uint32(len(x.Animations))
	w.Uint32(&l)
	for _, anim := range x.Animations {
		Animation(w, &anim)
	}
	w.Uint32(&x.CapeImageWidth)
	w.Uint32(&x.CapeImageHeight)
	w.ByteSlice(&x.CapeData)
	w.ByteSlice(&x.SkinGeometry)
	w.ByteSlice(&x.GeometryDataEngineVersion)
	w.ByteSlice(&x.AnimationData)
	w.String(&x.CapeID)
	w.String(&x.FullID)
	w.String(&x.ArmSize)
	w.String(&x.SkinColour)
	l = uint32(len(x.PersonaPieces))
	w.Uint32(&l)
	for _, piece := range x.PersonaPieces {
		SkinPiece(w, &piece)
	}
	l = uint32(len(x.PieceTintColours))
	w.Uint32(&l)
	for _, tint := range x.PieceTintColours {
		WriteSkinPieceTint(w, &tint)
	}

	w.Bool(&x.PremiumSkin)
	w.Bool(&x.PersonaSkin)
	w.Bool(&x.PersonaCapeOnClassicSkin)
	w.Bool(&x.PrimaryUser)
}

// SerialisedSkin reads a Skin x from Reader r.
func SerialisedSkin(r *Reader, x *Skin) {
	var animationCount, count uint32

	r.String(&x.SkinID)
	r.String(&x.PlayFabID)
	r.ByteSlice(&x.SkinResourcePatch)
	r.Uint32(&x.SkinImageWidth)
	r.Uint32(&x.SkinImageHeight)
	r.ByteSlice(&x.SkinData)
	r.Uint32(&animationCount)

	x.Animations = make([]SkinAnimation, animationCount)
	for i := uint32(0); i < animationCount; i++ {
		Animation(r, &x.Animations[i])
	}
	r.Uint32(&x.CapeImageWidth)
	r.Uint32(&x.CapeImageHeight)
	r.ByteSlice(&x.CapeData)
	r.ByteSlice(&x.SkinGeometry)
	r.ByteSlice(&x.GeometryDataEngineVersion)
	r.ByteSlice(&x.AnimationData)
	r.String(&x.CapeID)
	r.String(&x.FullID)
	r.String(&x.ArmSize)
	r.String(&x.SkinColour)

	r.Uint32(&count)
	x.PersonaPieces = make([]PersonaPiece, count)
	for i := uint32(0); i < count; i++ {
		SkinPiece(r, &x.PersonaPieces[i])
	}
	r.Uint32(&count)
	x.PieceTintColours = make([]PersonaPieceTintColour, count)
	for i := uint32(0); i < count; i++ {
		SkinPieceTint(r, &x.PieceTintColours[i])
	}
	if err := x.validate(); err != nil {
		r.InvalidValue(fmt.Sprintf("Skin %v", x.SkinID), "serialised skin", err.Error())
	}

	r.Bool(&x.PremiumSkin)
	r.Bool(&x.PersonaSkin)
	r.Bool(&x.PersonaCapeOnClassicSkin)
	r.Bool(&x.PrimaryUser)
}

// validate checks the skin and makes sure every one of its values are correct. It checks the image dimensions
// and makes sure they match the image size of the skin, cape and the skin's animations.
func (skin Skin) validate() error {
	if skin.SkinImageHeight*skin.SkinImageWidth*4 != uint32(len(skin.SkinData)) {
		return fmt.Errorf("expected size of skin is %vx%v (%v bytes total), but got %v bytes", skin.SkinImageWidth, skin.SkinImageHeight, skin.SkinImageHeight*skin.SkinImageWidth*4, len(skin.SkinData))
	}
	if skin.CapeImageHeight*skin.CapeImageWidth*4 != uint32(len(skin.CapeData)) {
		return fmt.Errorf("expected size of cape is %vx%v (%v bytes total), but got %v bytes", skin.CapeImageWidth, skin.CapeImageHeight, skin.CapeImageHeight*skin.CapeImageWidth*4, len(skin.CapeData))
	}
	for i, animation := range skin.Animations {
		if animation.ImageHeight*animation.ImageWidth*4 != uint32(len(animation.ImageData)) {
			return fmt.Errorf("expected size of animation %v is %vx%v (%v bytes total), but got %v bytes", i, animation.ImageWidth, animation.ImageHeight, animation.ImageHeight*animation.ImageWidth*4, len(animation.ImageData))
		}
	}
	return nil
}

const (
	SkinAnimationHead = iota + 1
	SkinAnimationBody32x32
	SkinAnimationBody128x128

	ExpressionTypeLinear = iota
	ExpressionTypeBlinking
)

// SkinAnimation represents an animation that may be added to a skin. The client plays the animation itself,
// without the server having to do so.
// The rate at which these animations play appears to be decided by the client.
type SkinAnimation struct {
	// ImageWidth and ImageHeight hold the dimensions of the animation image. Note that these are not the
	// dimensions in bytes, but in pixels.
	ImageWidth, ImageHeight uint32
	// ImageData is a byte slice of ImageWidth * ImageHeight bytes. It is an RGBA ordered byte representation
	// of the animation image pixels. The ImageData contains FrameCount images in it, which each represent one
	// stage of the animation. The actual part of the skin that this field holds depends on the AnimationType,
	// where SkinAnimationHead holds only the head and its hat, whereas the other animations hold the entire
	// body of the skin.
	ImageData []byte
	// AnimationType is the type of the animation, which is one of the types found above. The data that
	// ImageData contains depends on this type.
	AnimationType uint32
	// FrameCount is the amount of frames that the skin animation holds. The number of frames here is the
	// amount of images that may be found in the ImageData field.
	FrameCount float32
	// ExpressionType is the type of expression made by the skin, which is one the types found above.
	ExpressionType uint32
}

// Animation reads/writes a SkinAnimation x using IO r.
func Animation(r IO, x *SkinAnimation) {
	r.Uint32(&x.ImageWidth)
	r.Uint32(&x.ImageHeight)
	r.ByteSlice(&x.ImageData)
	r.Uint32(&x.AnimationType)
	r.Float32(&x.FrameCount)
	r.Uint32(&x.ExpressionType)
}

// PersonaPiece represents a piece of a persona skin. All pieces are sent separately.
type PersonaPiece struct {
	// PieceId is a UUID that identifies the piece itself, which is unique for each separate piece.
	PieceID string
	// PieceType holds the type of the piece. Several types I was able to find immediately are listed below.
	// - persona_skeleton
	// - persona_body
	// - persona_skin
	// - persona_bottom
	// - persona_feet
	// - persona_top
	// - persona_mouth
	// - persona_hair
	// - persona_eyes
	// - persona_facial_hair
	PieceType string
	// PackID is a UUID that identifies the pack that the persona piece belongs to.
	PackID string
	// Default specifies if the piece is one of the default pieces. This is true when the piece is one of
	// those that a Steve or Alex skin have.
	Default bool
	// ProductID is a UUID that identifies the piece when it comes to purchases. It is empty for pieces that
	// have the 'Default' field set to true.
	ProductID string
}

// SkinPiece reads/writes a PersonaPiece x using IO r.
func SkinPiece(r IO, x *PersonaPiece) {
	r.String(&x.PieceID)
	r.String(&x.PieceType)
	r.String(&x.PackID)
	r.Bool(&x.Default)
	r.String(&x.ProductID)
}

// PersonaPieceTintColour describes the tint colours of a specific piece of a persona skin.
type PersonaPieceTintColour struct {
	// PieceType is the type of the persona skin piece that this tint colour concerns. The piece type must
	// always be present in the persona pieces list, but not each piece type has a tint colour sent.
	// Pieces that do have a tint colour that I was able to find immediately are listed below.
	// - persona_mouth
	// - persona_eyes
	// - persona_hair
	PieceType string
	// Colours is a list four colours written in hex notation (note, that unlike the SkinColour field in
	// the ClientData struct, this is actually ARGB, not just RGB).
	// The colours refer to different parts of the skin piece. The 'persona_eyes' may have the following
	// colours: ["#ffa12722","#ff2f1f0f","#ff3aafd9","#0"]
	// The first hex colour represents the tint colour of the iris, the second hex colour represents the
	// eyebrows and the third represents the sclera. The fourth is #0 because there are only 3 parts of the
	// persona_eyes skin piece.
	Colours []string
}

// WriteSkinPieceTint writes a PersonaPieceTintColour x to Writer w.
func WriteSkinPieceTint(w *Writer, x *PersonaPieceTintColour) {
	w.String(&x.PieceType)
	l := uint32(len(x.Colours))
	w.Uint32(&l)
	for _, c := range x.Colours {
		w.String(&c)
	}
}

// SkinPieceTint reads a PersonaPieceTintColour x from Reader r.
func SkinPieceTint(r *Reader, x *PersonaPieceTintColour) {
	var c uint32
	r.String(&x.PieceType)

	r.Uint32(&c)
	x.Colours = make([]string, c)
	for i := uint32(0); i < c; i++ {
		r.String(&x.Colours[i])
	}
}
