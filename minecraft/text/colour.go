package text

import (
	"fmt"
	"golang.org/x/net/html"
	"regexp"
	"strings"
)

// cleaner represents the regex used to clean Minecraft formatting codes from a string.
var cleaner = regexp.MustCompile("§[0-9a-u]")

// Clean removes all Minecraft formatting codes from the string passed.
func Clean(s string) string {
	return cleaner.ReplaceAllString(s, "")
}

// ANSI converts all Minecraft text formatting codes in the values passed to ANSI formatting codes, so that
// it may be displayed properly in the terminal.
func ANSI(a ...any) string {
	str := make([]string, len(a))
	for i, v := range a {
		str[i] = minecraftReplacer.Replace(fmt.Sprint(v))
	}
	return strings.Join(str, " ")
}

// Colourf colours the format string using HTML tags after first escaping all parameters passed and
// substituting them in the format string. The following colours and formatting may be used:
//
//	black, dark-blue, dark-green, dark-aqua, dark-red, dark-purple, gold, grey, dark-grey, blue, green, aqua,
//	red, purple, yellow, white, dark-yellow, quartz, iron, netherite, redstone, copper, gold, emerald, diamond,
//	lapis, amethyst, obfuscated, bold (b), and italic (i).
//
// These HTML tags may also be nested, like so:
// `<red>Hello <bold>World</bold>!</red>`
func Colourf(format string, a ...any) string {
	str := fmt.Sprintf(format, a...)

	e := &enc{w: &strings.Builder{}, first: true}
	t := html.NewTokenizer(strings.NewReader(str))
	for {
		if t.Next() == html.ErrorToken {
			break
		}
		e.process(t.Token())
	}
	return e.w.String()
}

// enc holds the state of a string to be processed for colour substitution.
type enc struct {
	w           *strings.Builder
	formatStack []string
	first       bool
}

// process handles a single html.Token and either writes the string to the strings.Builder, adds a colour to
// the stack or removes a colour from the stack.
func (e *enc) process(tok html.Token) {
	if e.first {
		e.w.WriteString(Reset)
		e.first = false
	}
	switch tok.Type {
	case html.TextToken:
		e.writeText(tok.Data)
	case html.StartTagToken:
		if format, ok := strMap[tok.Data]; ok {
			e.formatStack = append(e.formatStack, format)
			return
		}
		// Not a known colour, so just write the token as a string.
		e.writeText("<" + tok.Data + ">")
	case html.EndTagToken:
		for i, format := range e.formatStack {
			if f, ok := strMap[tok.Data]; ok && f == format {
				e.formatStack = append(e.formatStack[:i], e.formatStack[i+1:]...)
				return
			}
		}
		// Not a known colour, so just write the token as a string.
		e.writeText("</" + tok.Data + ">")
	}
}

// writeText writes text to the encoder by encasing it in the current format stack.
func (e *enc) writeText(s string) {
	for _, format := range e.formatStack {
		e.w.WriteString(format)
	}
	e.w.WriteString(s)
	if len(e.formatStack) != 0 {
		e.w.WriteString(Reset)
	}
}
