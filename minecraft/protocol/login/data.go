package login

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/text/language"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// IdentityData contains identity data of the player logged in. It is found in one of the JWT claims signed
// by Mojang, and can thus be trusted.
type IdentityData struct {
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account. It is empty if the user is not logged into its XBL account.
	XUID string
	// Identity is the UUID of the player, which will also remain consistent for as long as the user is logged
	// into its XBOX Live account.
	Identity string `json:"identity"`
	// DisplayName is the username of the player, which may be changed by the user. It should for that reason
	// not be used as a key to store information.
	DisplayName string `json:"displayName"`
	// TitleID is a numerical ID present only if the user is logged into XBL. It holds the title ID (XBL
	// related) of the version that the player is on. Some of these IDs may be found below.
	// Win10: 896928775
	// Mobile: 1739947436
	// Nintendo: 2047319603
	// Note that these IDs are protected using XBOX Live, making the spoofing of this data very difficult.
	TitleID string `json:"titleId,omitempty"`
}

// checkUsername is used to check if a username is valid according to the Microsoft specification: "You can
// use up to 15 characters: Aa-Zz, 0-9, and single spaces. It cannot start with a number and cannot start or
// end with a space."
var checkUsername = regexp.MustCompile("[A-Za-z0-9 ]").MatchString

// Validate validates the identity data. It returns an error if any data contained in the IdentityData is
// invalid.
func (data IdentityData) Validate() error {
	if _, err := strconv.ParseInt(data.XUID, 10, 64); err != nil && len(data.XUID) != 0 {
		return fmt.Errorf("XUID must be parseable as an int64, but got %v", data.XUID)
	}
	if id, err := uuid.Parse(data.Identity); err != nil || id == uuid.Nil {
		return fmt.Errorf("UUID must be parseable as a valid UUID, but got %v", data.Identity)
	}
	if len(data.DisplayName) == 0 || len(data.DisplayName) > 15 {
		return fmt.Errorf("DisplayName must not be empty or longer than 15 characters, but got %v characters", len(data.DisplayName))
	}
	if data.DisplayName[0] == ' ' || data.DisplayName[len(data.DisplayName)-1] == ' ' {
		return fmt.Errorf("DisplayName may not have a space as first/last character, but got %v", data.DisplayName)
	}
	if data.DisplayName[0] >= '0' && data.DisplayName[0] <= '9' {
		return fmt.Errorf("DisplayName may not have a number as first character, but got %v", data.DisplayName)
	}
	if !checkUsername(data.DisplayName) {
		return fmt.Errorf("DisplayName must only contain numbers, letters and spaces, but got %v", data.DisplayName)
	}
	// We check here if the name contains at least 2 spaces after each other, which is not allowed. The name
	// is only allowed to have single spaces.
	if strings.Contains(data.DisplayName, "  ") {
		return fmt.Errorf("DisplayName must only have single spaces, but got %v", data.DisplayName)
	}
	return nil
}

// ClientData is a container of client specific data of a Login packet. It holds data such as the skin of a
// player, but also its language code and device information.
type ClientData struct {
	// AnimatedImageData is a list of image data for animations. Each of the elements in this slice holds the
	// image data of a single frame of the animation.
	AnimatedImageData []SkinAnimation
	// CapeData is a base64 encoded string of cape data. This is usually an empty string, as skins typically
	// don't carry capes themselves.
	CapeData string
	// CapeID is an ID which, like the SkinID, identifies a skin. Usually this is either empty for no skin or
	// some ID containing a UUID in it.
	CapeID string `json:"CapeId"`
	// CapeImageHeight and CapeImageWidth are the dimensions of the cape's image.
	CapeImageHeight, CapeImageWidth int
	// CapeOnClassicSkin specifies if the cape that the player has equipped is part of a classic skin, which
	// usually points to one of the older MineCon capes.
	CapeOnClassicSkin bool
	// ClientRandomID is a random client ID number generated for the client. It usually remains consistent
	// through sessions and through game restarts.
	ClientRandomID int64 `json:"ClientRandomId"`
	// CurrentInputMode is the input mode used by the client. It is 1 for mobile and win10, but is different
	// for console input.
	CurrentInputMode int
	// DefaultInputMode is the default input mode used by the device.
	DefaultInputMode int
	// DeviceModel is a string indicating the device model used by the player. At the moment, it appears that
	// this name is always '(Standard system devices) System devices'.
	DeviceModel string
	// DeviceOS is a numerical ID indicating the OS of the device.
	DeviceOS protocol.DeviceOS
	// DeviceID is usually a UUID specific to the device. A different user will have the same UUID for this.
	// DeviceID is not guaranteed to always be a UUID. It is a base64 encoded string under some circumstances.
	DeviceID string `json:"DeviceId"`
	// GameVersion is the game version of the player that attempted to join, for example '1.11.0'.
	GameVersion string
	// GUIScale is the GUI scale of the player. It is by default 0, and is otherwise -1 or -2 for a smaller
	// GUI scale than usual.
	GUIScale int `json:"GuiScale"`
	// IsEditorMode is a value to dictate if the player is in editor mode.
	IsEditorMode bool
	// LanguageCode is the language code of the player. It looks like 'en_UK'. It follows the ISO language
	// codes, but hyphens ('-') are replaced with underscores. ('_')
	LanguageCode string
	// PersonaSkin specifies if the skin was a persona skin, meaning that it was created through the in-game
	// skin creator.
	PersonaSkin bool
	// PlatformOfflineID is either a UUID or an empty string ...
	PlatformOfflineID string `json:"PlatformOfflineId"`
	// PlatformOnlineID is either a uint64 or an empty string ...
	PlatformOnlineID string `json:"PlatformOnlineId"`
	// PlatformUserID holds a UUID which is only sent if the DeviceOS is of type device.XBOX. Its function
	// is not exactly clear.
	PlatformUserID string `json:"PlatformUserId,omitempty"`
	// PremiumSkin indicates if the skin the player held was a premium skin, meaning it was obtained through
	// payment.
	PremiumSkin bool
	// SelfSignedID is a UUID that remains consistent through restarts of the game and new game sessions.
	SelfSignedID string `json:"SelfSignedId"`
	// ServerAddress is the exact address the player used to join the server with. This may be either an
	// actual address, or a hostname. ServerAddress also has the port in it, in the shape of
	// 'address:port`.
	ServerAddress string
	// TODO: Find out what value SkinAnimationData holds and when it does hold something.
	SkinAnimationData string
	// SkinData is a base64 encoded byte slice of 64*32*4, 64*64*4 or 128*128*4 bytes. It is a RGBA ordered
	// byte representation of the skin colours.
	SkinData string
	// SkinGeometry is a base64 JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry string `json:"SkinGeometryData"`
	// SkinGeometryVersion is the version for SkinGeometry.
	SkinGeometryVersion string `json:"SkinGeometryDataEngineVersion"`
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string `json:"SkinId"`
	// PlayFabID is the PlayFab ID produced for the player's skin. PlayFab is
	// the company that hosts the Marketplace, skins and other related features
	// from the game. This ID is the ID of the skin used to store the skin
	// inside of PlayFab. PlayFabID is a hex encoded string, usually consisting
	// of 16 characters.
	PlayFabID string `json:"PlayFabId"`
	// SkinImageHeight and SkinImageWidth are the dimensions of the skin's image data.
	SkinImageHeight, SkinImageWidth int
	// SkinResourcePatch is a base64 encoded string which holds JSON data. The content of the JSON data points
	// to the assets that should be used to shape the skin. An example with a head animation can be found
	// below.
	// {
	//   "geometry" : {
	//      "animated_face" : "geometry.animated_face_persona_d1625e47f4c9399f_0_1",
	//      "default" : "geometry.persona_d1625e47f4c9399f_0_1"
	//   }
	// }
	// A skin resource patch must be present at all times. The minimum required data that the field must hold
	// is {"geometry": {"default": "geometry.persona_d1625e47f4c9399f_0_1"}}
	SkinResourcePatch string
	// SkinColour is a hex representation (including #) of the base colour of the skin. An example of the
	// colour sent here is '#b37b62'.
	SkinColour string `json:"SkinColor"`
	// ArmSize is the size of the arms of the player's model. This is either 'wide' (generally for male skins)
	// or 'slim' (generally for female skins).
	ArmSize string
	// PersonaPieces is a list of all persona pieces that the skin is composed of.
	PersonaPieces []PersonaPiece
	// PieceTintColours is a list of specific tint colours for (some of) the persona pieces found in the list
	// above.
	PieceTintColours []PersonaPieceTintColour `json:"PieceTintColors"`
	// ThirdPartyName is the username of the player. This username should not be used however. The DisplayName
	// sent in the IdentityData should be preferred over this.
	ThirdPartyName string
	// ThirdPartyNameOnly specifies if the user only has a third party name. It should always be assumed to be
	// false, because the third party name is not XBOX Live Auth protected, meaning it can be tempered with
	// and the username changed.
	// Although this field is obviously here for a reason, allowing this is too dangerous and should never be
	// done.
	ThirdPartyNameOnly bool
	// UIProfile is the UI profile used. For the 'Pocket' UI, this is 1. For the 'Classic' UI, this is 0.
	UIProfile int
	// TrustedSkin is a boolean indicating if the skin the client is using is trusted.
	TrustedSkin bool
	// OverrideSkin is a boolean that does not make sense to be here. The current usage of this field is unknown.
	OverrideSkin bool
	// CompatibleWithClientSideChunkGen is a boolean indicating if the client's hardware is capable of using the client
	// side chunk generation system.
	CompatibleWithClientSideChunkGen bool
}

// PersonaPiece represents a piece of a persona skin. All pieces are sent separately.
type PersonaPiece struct {
	// Default specifies if the piece is one of the default pieces. This is true when the piece is one of
	// those that a Steve or Alex skin have.
	Default bool `json:"IsDefault"`
	// PackID is a UUID that identifies the pack that the persona piece belongs to.
	PackID string `json:"PackId"`
	// PieceId is a UUID that identifies the piece itself, which is unique for each separate piece.
	PieceID string `json:"PieceId"`
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
	// ProductID is a UUID that identifies the piece when it comes to purchases. It is empty for pieces that
	// have the 'IsDefault' field set to true.
	ProductID string `json:"ProductId"`
}

// PersonaPieceTintColour describes the tint colours of a specific piece of a persona skin.
type PersonaPieceTintColour struct {
	// Colours is an array of four colours written in hex notation (note, that unlike the SkinColor field in
	// the ClientData struct, this is actually ARGB, not just RGB).
	// The colours refer to different parts of the skin piece. The 'persona_eyes' may have the following
	// colours: ["#ffa12722","#ff2f1f0f","#ff3aafd9","#0"]
	// The first hex colour represents the tint colour of the iris, the second hex colour represents the
	// eyebrows and the third represents the sclera. The fourth is #0 because there are only 3 parts of the
	// persona_eyes skin piece.
	Colours [4]string `json:"Colors"`
	// PieceType is the type of the persona skin piece that this tint colour concerns. The piece type must
	// always be present in the persona pieces list, but not each piece type has a tint colour sent.
	// Pieces that do have a tint colour that I was able to find immediately are listed below.
	// - persona_mouth
	// - persona_eyes
	// - persona_hair
	PieceType string
}

// SkinAnimation is an animation that may be present. It is applied on top of the skin default and is cycled
// through client-side.
type SkinAnimation struct {
	// Frames is the amount of frames of the animation. The number of Frames here specifies how many
	// frames may be found in the Image data.
	Frames float64
	// Image is a base64 encoded byte slice of ImageWidth * ImageHeight bytes. It is an RGBA ordered byte
	// representation of the animation image pixels. The ImageData contains FrameCount images in it, which
	// each represent one stage of the animation. The actual part of the skin that this field holds
	// depends on the Type, where SkinAnimationHead holds only the head and its hat, whereas the other
	// animations hold the entire body of the skin.
	Image string
	// ImageHeight and ImageWidth are the dimensions of the animated image. Note that the size of this
	// image is not always 32/64/128.
	ImageHeight, ImageWidth int
	// Type is the type of the animation, which defines what part of the body the Image data holds. It is
	// one of the following:
	// 0 -> 'None', doesn't typically occur.
	// 1 -> Face animation.
	// 2 -> 32x32 Body animation.
	// 3 -> 128x128 Body animation.
	Type int
	// ExpressionType is the type of expression made by the skin, which is one of the following:
	// 0 -> Linear.
	// 1 -> Blinking.
	AnimationExpression int
}

// checkVersion is used to check if a version is an actual valid version. It must only contain numbers and
// dots.
var checkVersion = regexp.MustCompile("[0-9.]").MatchString

// Validate validates the client data. It returns an error if any of the fields checked did not carry a valid
// value.
func (data ClientData) Validate() error {
	if data.DeviceOS <= 0 || data.DeviceOS > 15 {
		return fmt.Errorf("DeviceOS must carry a value between 1 and 15, but got %v", data.DeviceOS)
	}
	if !checkVersion(data.GameVersion) {
		return fmt.Errorf("GameVersion must only contain dots and numbers, but got %v", data.GameVersion)
	}
	if _, err := language.Parse(strings.Replace(data.LanguageCode, "_", "-", 1)); err != nil {
		return fmt.Errorf("LanguageCode must be a valid BCP-47 ISO language code, but got %v", data.LanguageCode)
	}
	if _, err := uuid.Parse(data.PlatformOfflineID); err != nil && len(data.PlatformOfflineID) != 0 {
		return fmt.Errorf("PlatformOfflineID must be parseable as a valid UUID or empty, but got %v", data.PlatformOfflineID)
	}
	if _, err := strconv.ParseUint(data.PlatformOnlineID, 10, 64); err != nil && len(data.PlatformOnlineID) != 0 {
		return fmt.Errorf("PlatformOnlineID must be parseable as an int64 or empty, but got %v", data.PlatformOnlineID)
	}
	if _, err := uuid.Parse(data.SelfSignedID); err != nil {
		return fmt.Errorf("SelfSignedID must be parseable as a valid UUID, but got %v", data.SelfSignedID)
	}
	if _, err := net.ResolveUDPAddr("udp", data.ServerAddress); err != nil {
		return fmt.Errorf("ServerAddress must be resolveable as a UDP address, but got %v", data.ServerAddress)
	}
	if err := base64DecLength(data.SkinData, data.SkinImageHeight*data.SkinImageWidth*4); err != nil {
		return fmt.Errorf("SkinData is invalid: %w", err)
	}
	if err := base64DecLength(data.CapeData, data.CapeImageHeight*data.CapeImageWidth*4); err != nil {
		return fmt.Errorf("CapeData is invalid: %w", err)
	}
	if _, err := hex.DecodeString(data.PlayFabID); err != nil {
		return fmt.Errorf("PlayFabID must be hex string, but got %v", data.PlayFabID)
	}
	for _, anim := range data.AnimatedImageData {
		if err := base64DecLength(anim.Image, anim.ImageHeight*anim.ImageWidth*4); err != nil {
			return fmt.Errorf("invalid animated image data: %w", err)
		}
		if anim.Type < 0 || anim.Type > 3 {
			return fmt.Errorf("invalid animation type: %v", anim.Type)
		}
	}
	if geomData, err := base64.StdEncoding.DecodeString(data.SkinGeometry); err != nil {
		return fmt.Errorf("SkinGeometry was not a valid base64 string: %w", err)
	} else if len(geomData) != 0 {
		m := make(map[string]any)
		if err := json.Unmarshal(geomData, &m); err != nil {
			return fmt.Errorf("SkinGeometry base64 decoded was not a valid JSON string: %w", err)
		}
	}
	b, err := base64.StdEncoding.DecodeString(data.SkinResourcePatch)
	if err != nil {
		return fmt.Errorf("SkinResourcePatch was not a valid base64 string: %w", err)
	}
	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("SkinResourcePatch base64 decoded was not a valid JSON string: %w", err)
	}
	if data.SkinID == "" {
		return fmt.Errorf("SkinID must not be an empty string")
	}
	if data.UIProfile < 0 || data.UIProfile > 2 {
		return fmt.Errorf("UIProfile must be between 0-2, but got %v", data.UIProfile)
	}
	return nil
}

// base64DecLength decodes the base64 data passed and checks if its length is one of the valid lengths
// passed. If either of these checks fails, an error is returned.
func base64DecLength(base64Data string, validLengths ...int) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("decode base64 data: %w", err)
	}
	actualLength := len(data)
	for _, length := range validLengths {
		if length == actualLength {
			return nil
		}
	}
	return fmt.Errorf("invalid size: got %v, expected one of %v", actualLength, validLengths)
}
