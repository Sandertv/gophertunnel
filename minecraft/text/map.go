package text

import "strings"

const (
	Black      = "§0"
	DarkBlue   = "§1"
	DarkGreen  = "§2"
	DarkAqua   = "§3"
	DarkRed    = "§4"
	DarkPurple = "§5"
	Orange     = "§6"
	Grey       = "§7"
	DarkGrey   = "§8"
	Blue       = "§9"
	Green      = "§a"
	Aqua       = "§b"
	Red        = "§c"
	Purple     = "§d"
	Yellow     = "§e"
	White      = "§f"
	DarkYellow = "§g"
	Quartz     = "§h"
	Iron       = "§i"
	Netherite  = "§j"
	Obfuscated = "§k"
	Bold       = "§l"
	Redstone   = "§m"
	Copper     = "§n"
	Italic     = "§o"
	Gold       = "§p"
	Emerald    = "§q"
	Reset      = "§r"
	Diamond    = "§s"
	Lapis      = "§t"
	Amethyst   = "§u"
	Resin      = "§v"
)

const (
	ansiBlack      = "\x1b[38;5;16m"
	ansiDarkBlue   = "\x1b[38;5;19m"
	ansiDarkGreen  = "\x1b[38;5;34m"
	ansiDarkAqua   = "\x1b[38;5;37m"
	ansiDarkRed    = "\x1b[38;5;124m"
	ansiDarkPurple = "\x1b[38;5;127m"
	ansiOrange     = "\x1b[38;5;214m"
	ansiGrey       = "\x1b[38;5;145m"
	ansiDarkGrey   = "\x1b[38;5;59m"
	ansiBlue       = "\x1b[38;5;63m"
	ansiGreen      = "\x1b[38;5;83m"
	ansiAqua       = "\x1b[38;5;87m"
	ansiRed        = "\x1b[38;5;203m"
	ansiPurple     = "\x1b[38;5;207m"
	ansiYellow     = "\x1b[38;5;227m"
	ansiWhite      = "\x1b[38;5;231m"
	ansiDarkYellow = "\x1b[38;5;226m"
	ansiQuartz     = "\x1b[38;5;224m"
	ansiIron       = "\x1b[38;5;251m"
	ansiNetherite  = "\x1b[38;5;234m"
	ansiRedstone   = "\x1b[38;5;1m"
	ansiCopper     = "\x1b[38;5;216m"
	ansiGold       = "\x1b[38;5;220m"
	ansiEmerald    = "\x1b[38;5;71m"
	ansiDiamond    = "\x1b[38;5;122m"
	ansiLapis      = "\x1b[38;5;4m"
	ansiAmethyst   = "\x1b[38;5;171m"
	ansiResin      = "\x1b[38;5;172m"

	ansiObfuscated = ""
	ansiBold       = "\x1b[1m"
	ansiItalic     = "\x1b[3m"
	ansiReset      = "\x1b[m"
)

var m = map[string]string{
	Black:      ansiBlack,
	DarkBlue:   ansiDarkBlue,
	DarkGreen:  ansiDarkGreen,
	DarkAqua:   ansiDarkAqua,
	DarkRed:    ansiDarkRed,
	DarkPurple: ansiDarkPurple,
	Orange:     ansiOrange,
	Grey:       ansiGrey,
	DarkGrey:   ansiDarkGrey,
	Blue:       ansiBlue,
	Green:      ansiGreen,
	Aqua:       ansiAqua,
	Red:        ansiRed,
	Purple:     ansiPurple,
	Yellow:     ansiYellow,
	White:      ansiWhite,
	DarkYellow: ansiDarkYellow,
	Quartz:     ansiQuartz,
	Iron:       ansiIron,
	Netherite:  ansiNetherite,
	Redstone:   ansiRedstone,
	Copper:     ansiCopper,
	Gold:       ansiGold,
	Emerald:    ansiEmerald,
	Diamond:    ansiDiamond,
	Lapis:      ansiLapis,
	Amethyst:   ansiAmethyst,
	Resin:      ansiResin,

	Obfuscated: ansiObfuscated,
	Bold:       ansiBold,
	Reset:      ansiReset,
	Italic:     ansiItalic,
}

var strMap = map[string]string{
	"black":       Black,
	"dark-blue":   DarkBlue,
	"dark-green":  DarkGreen,
	"dark-aqua":   DarkAqua,
	"dark-red":    DarkRed,
	"dark-purple": DarkPurple,
	"orange":      Orange,
	"grey":        Grey,
	"dark-grey":   DarkGrey,
	"blue":        Blue,
	"green":       Green,
	"aqua":        Aqua,
	"red":         Red,
	"purple":      Purple,
	"yellow":      Yellow,
	"white":       White,
	"dark-yellow": DarkYellow,
	"quartz":      Quartz,
	"iron":        Iron,
	"netherite":   Netherite,
	"obfuscated":  Obfuscated,
	"bold":        Bold,
	"b":           Bold,
	"redstone":    Redstone,
	"copper":      Copper,
	"gold":        Gold,
	"emerald":     Emerald,
	"italic":      Italic,
	"i":           Italic,
	"diamond":     Diamond,
	"lapis":       Lapis,
	"amethyst":    Amethyst,
	"resin":       Resin,
}

// minecraftReplacer and ansiReplacer are used to translate ANSI formatting codes to Minecraft formatting
// codes and vice versa.
var minecraftReplacer *strings.Replacer

func init() {
	var minecraftToANSI []string
	for minecraftCode, ansiCode := range m {
		minecraftToANSI = append(minecraftToANSI, minecraftCode, ansiCode)
	}
	minecraftReplacer = strings.NewReplacer(minecraftToANSI...)
}
