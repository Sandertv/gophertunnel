package login

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/device"
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
	// logged in with the XBOX Live account..
	XUID string
	// Identity is the UUID of the player, which will also remain consistent for as long as the user is logged
	// into its XBOX Live account.
	Identity string `json:"identity"`
	// DisplayName is the username of the player, which may be changed by the user. It should for that reason
	// not be used as a key to store information.
	DisplayName string `json:"displayName"`
}

// checkUsername is used to check if a username is valid according to the Microsoft specification: "You can
// use up to 15 characters: Aa-Zz, 0-9, and single spaces. It cannot start with a number and cannot start or
// end with a space."
var checkUsername = regexp.MustCompile("[A-Z0-9 ]").MatchString

// Validate validates the identity data. It returns an error if any data contained in the IdentityData is
// invalid.
func (data IdentityData) Validate() error {
	if _, err := strconv.ParseInt(data.XUID, 10, 64); err != nil && len(data.XUID) != 0 {
		return fmt.Errorf("XUID must be parseable as an int64, but got %v", data.XUID)
	}
	if _, err := uuid.Parse(data.Identity); err != nil {
		return fmt.Errorf("UUID must be parseable as a valid UUID, but got %v", data.Identity)
	}
	if len(data.DisplayName) == 0 || len(data.DisplayName) > 15 {
		return fmt.Errorf("DisplayName must not be empty or longer than 15 characters, but got %v characters", len(data.DisplayName))
	}
	if data.DisplayName[0] == ' ' || (data.DisplayName[0] >= '0' && data.DisplayName[0] <= '9') {
		return fmt.Errorf("DisplayName may not have a space or number as first/last character, but got %v", data.DisplayName)
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
	// CapeData is a base64 encoded string of cape data. This is usually an empty string, as skins typically
	// don't carry capes themselves.
	CapeData string
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
	DeviceOS device.OS
	// DeviceID is a UUID specific to the device. A different user will have the same UUID for this.
	DeviceID string `json:"DeviceId"`
	// GameVersion is the game version of the player that attempted to join, for example '1.11.0'.
	GameVersion string
	// GUIScale is the GUI scale of the player. It is by default 0, and is otherwise -1 for a smaller GUI
	// scale than usual.
	GUIScale int `json:"GuiScale"`
	// LanguageCode is the language code of the player. It looks like 'en_UK'. It follows the ISO language
	// codes, but hyphens ('-') are replaced with underscores. ('_')
	LanguageCode string
	// PlatformOfflineID is either a UUID or an empty string ...
	PlatformOfflineID string `json:"PlatformOfflineId"`
	// PlatformOnlineID is either a UUID or an empty string ...
	PlatformOnlineID string `json:"PlatformOnlineId"`
	// PremiumSkin indicates if the skin the player held was a premium skin, meaning it was obtained through
	// payment.
	PremiumSkin bool
	// SelfSignedID is a UUID that remains consistent through restarts of the game and new game sessions.
	SelfSignedID string `json:"SelfSignedId"`
	// ServerAddress is the exact address the player used to join the server with. This may be either an
	// actual address, or a hostname. ServerAddress also has the port in it, in the shape of
	// 'address:port`.
	ServerAddress string
	// SkinData is a base64 encoded byte slice of 64*32*4, 64*64*4 or 128*128*4 bytes. It is a RGBA ordered
	// byte representation of the skin colours.
	SkinData string
	// SkinGeometry is a base64 JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry string
	// SkinGeometryName is the geometry name of the skin geometry above. This name must be equal to one of the
	// outer names found in the SkinGeometry, so that the client can find the correct geometry data.
	SkinGeometryName string
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string `json:"SkinId"`
	// ThirdPartyName is the username of the player. This username should not be used however. The DisplayName
	// sent in the IdentityData should be preferred over this.
	ThirdPartyName string
	// UIProfile is the UI profile used. For the 'Pocket' UI, this is 1. For the 'Classic' UI, this is 0.
	UIProfile int
}

// checkVersion is used to check if a version is an actual valid version. It must only contain numbers and
// dots.
var checkVersion = regexp.MustCompile("[0-9.]").MatchString

// Validate validates the client data. It returns an error if any of the fields checked did not carry a valid
// value.
func (data ClientData) Validate() error {
	if err := base64DecLength(data.CapeData, 64*32*4, 0); err != nil {
		return fmt.Errorf("CapeData invalid: %v", err)
	}
	if data.DeviceOS <= 0 || data.DeviceOS > 12 {
		return fmt.Errorf("DeviceOS must carry a value between 1 and 12, but got %v", data.DeviceOS)
	}
	if _, err := uuid.Parse(data.DeviceID); err != nil {
		return fmt.Errorf("DeviceID must be parseable as a valid UUID, but got %v", data.DeviceID)
	}
	if !checkVersion(data.GameVersion) {
		return fmt.Errorf("GameVersion must only contain dots and numbers, but got %v", data.GameVersion)
	}
	if data.GUIScale != -1 && data.GUIScale != 0 {
		return fmt.Errorf("GUIScale must be either -1 or 0, but got %v", data.GUIScale)
	}
	if _, err := language.Parse(strings.Replace(data.LanguageCode, "_", "-", 1)); err != nil {
		return fmt.Errorf("LanguageCode must be a valid BCP-47 ISO language code, but got %v", data.LanguageCode)
	}
	if _, err := uuid.Parse(data.PlatformOfflineID); err != nil && len(data.PlatformOfflineID) != 0 {
		return fmt.Errorf("PlatformOfflineID must be parseable as a valid UUID or empty, but got %v", data.PlatformOfflineID)
	}
	if _, err := uuid.Parse(data.PlatformOnlineID); err != nil && len(data.PlatformOnlineID) != 0 {
		return fmt.Errorf("PlatformOnlineID must be parseable as a valid UUID or empty, but got %v", data.PlatformOnlineID)
	}
	if _, err := uuid.Parse(data.SelfSignedID); err != nil {
		return fmt.Errorf("SelfSignedID must be parseable as a valid UUID, but got %v", data.SelfSignedID)
	}
	if _, err := net.ResolveUDPAddr("udp", data.ServerAddress); err != nil {
		return fmt.Errorf("ServerAddress must be resolveable as a UDP address, but got %v", data.ServerAddress)
	}
	if err := base64DecLength(data.SkinData, 32*64*4, 64*64*4, 128*128*4); err != nil {
		return fmt.Errorf("SkinData is invalid: %v", err)
	}
	if geomData, err := base64.StdEncoding.DecodeString(data.SkinGeometry); err != nil {
		return fmt.Errorf("SkinGeometry was not a valid base64 string: %v", err)
	} else {
		m := make(map[string]interface{})
		if err := json.Unmarshal(geomData, &m); err != nil {
			return fmt.Errorf("SkinGeometry base64 decoded was not a valid JSON string: %v", err)
		}
	}
	if data.SkinID == "" {
		return fmt.Errorf("SkinID must not be an empty string")
	}
	if data.UIProfile != 0 && data.UIProfile != 1 {
		return fmt.Errorf("UIProfile must be either 0 or 1, but got %v", data.UIProfile)
	}

	return nil
}

// base64DecLength decodes the base64 data passed and checks if its length is one of the valid lengths
// passed. If either of these checks fails, an error is returned.
func base64DecLength(base64Data string, validLengths ...int) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("error decoding base64 data: %v", err)
	}
	actualLength := len(data)
	for _, length := range validLengths {
		if length == actualLength {
			return nil
		}
	}
	return fmt.Errorf("invalid size: got %v, but expected one of %v", actualLength, validLengths)
}