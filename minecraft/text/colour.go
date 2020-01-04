package text

import (
	"fmt"
	"strings"
)

// FormatFunc represents a function that may be called on a list of values to format the values and apply
// a colour or format on them, such as making the text bold.
// The formatting on the values passed is done using Minecraft formatting codes. To turn it into a string that
// may be printed, use text.ANSI(string).
type FormatFunc func(a ...interface{}) string

// ANSI converts all Minecraft text formatting codes in the values passed to ANSI formatting codes, so that
// it may be displayed properly in the terminal.
func ANSI(a ...interface{}) string {
	str := make([]string, len(a))
	for i, v := range a {
		str[i] = minecraftReplacer.Replace(fmt.Sprint(v))
	}
	return strings.Join(str, " ")
}

// Minecraft converts all ANSI text formatting codes in the values passed to Minecraft formatting codes, so
// that it may be displayed in-game.
func Minecraft(a ...interface{}) string {
	str := make([]string, len(a))
	for i, v := range a {
		str[i] = ansiReplacer.Replace(fmt.Sprint(v))
	}
	return strings.Join(str, " ")
}

// Black returns a black formatter.
func Black() FormatFunc {
	return formatFunc(black, reset)
}

// DarkBlue returns a dark blue formatter.
func DarkBlue() FormatFunc {
	return formatFunc(darkBlue, reset)
}

// DarkGreen returns a dark green formatter.
func DarkGreen() FormatFunc {
	return formatFunc(darkGreen, reset)
}

// DarkAqua returns a dark aqua formatter.
func DarkAqua() FormatFunc {
	return formatFunc(darkAqua, reset)
}

// DarkRed returns a dark red formatter.
func DarkRed() FormatFunc {
	return formatFunc(darkRed, reset)
}

// DarkPurple returns a dark purple formatter.
func DarkPurple() FormatFunc {
	return formatFunc(darkPurple, reset)
}

// Gold returns a gold formatter.
func Gold() FormatFunc {
	return formatFunc(gold, reset)
}

// Grey returns a grey formatter.
func Grey() FormatFunc {
	return formatFunc(grey, reset)
}

// DarkGrey returns a dark grey formatter.
func DarkGrey() FormatFunc {
	return formatFunc(darkGrey, reset)
}

// Blue returns a blue formatter.
func Blue() FormatFunc {
	return formatFunc(blue, reset)
}

// Green returns a green formatter.
func Green() FormatFunc {
	return formatFunc(green, reset)
}

// Aqua returns an aqua formatter.
func Aqua() FormatFunc {
	return formatFunc(aqua, reset)
}

// Red returns a red formatter.
func Red() FormatFunc {
	return formatFunc(red, reset)
}

// Purple returns a purple formatter.
func Purple() FormatFunc {
	return formatFunc(purple, reset)
}

// Yellow returns a yellow formatter.
func Yellow() FormatFunc {
	return formatFunc(yellow, reset)
}

// White returns a white formatter.
func White() FormatFunc {
	return formatFunc(white, reset)
}

// DarkYellow returns a dark yellow formatter.
func DarkYellow() FormatFunc {
	return formatFunc(darkYellow, reset)
}

// Obfuscated returns an obfuscated formatter.
func Obfuscated() FormatFunc {
	return formatFunc(obfuscated, reset)
}

// Bold returns a bold formatter.
func Bold() FormatFunc {
	return formatFunc(bold, reset)
}

// Strikethrough returns a strikethrough formatter.
func Strikethrough() FormatFunc {
	return formatFunc(strikethrough, reset)
}

// Italic returns an italic formatter.
func Italic() FormatFunc {
	return formatFunc(italic, reset)
}

// formatFunc produces a FormatFunc for a given format code and reset code passed. It will start each of the
// arguments passed to the FormatFunc with the format code and end with the reset code.
func formatFunc(formatCode, reset string) FormatFunc {
	return func(a ...interface{}) string {
		str := &strings.Builder{}
		str.WriteString(formatCode)
		for i, v := range a {
			if i != 0 {
				str.WriteByte(' ')
			}
			str.WriteString(fmt.Sprint(v))
			if i != len(a)-1 {
				str.WriteString(formatCode)
			} else {
				str.WriteString(reset)
			}
		}
		return str.String()
	}
}
