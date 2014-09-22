package styles

import (
	"fmt"
)

type Color int

const (
	// All colors values are 1 greater than the actual value so that the
	// default value for a new `Style` is `Clear`

	Clear        Color = 0 // Do not change existing colors
	White        Color = 1 // Also color reset in most clients
	Black        Color = 2
	Blue         Color = 3
	Green        Color = 4
	LightRed     Color = 5
	Red          Color = 6
	Magenta      Color = 7
	Orange       Color = 8
	Yellow       Color = 9
	LightGreen   Color = 10
	Cyan         Color = 11
	LightCyan    Color = 12
	LightBlue    Color = 13
	LightMagenta Color = 14
	Gray         Color = 15
	LightGray    Color = 16
)

// Sets the foreground to `Color` and formats the text with Sprintf
func (self Color) Fg(format string, a ...interface{}) string {
	return PaintColors(self, Clear, format, a...)
}

// Sets the background to `Color` and formats the text with Sprintf
func (self Color) Bg(format string, a ...interface{}) string {
	return PaintColors(Clear, self, format, a...)
}

// Sets the foregroudn and background to `Color` and formats the text with
// Sprintf
func (self Color) Paint(format string, a ...interface{}) string {
	return PaintColors(self, self, format, a...)
}

// Sets the foreground to `fg` and the background to `bg` and formats the text
// with Sprintf
func PaintColors(fg, bg Color, format string, a ...interface{}) string {
	return setColors(fg, bg) + fmt.Sprintf(format, a...) + string(Reset)
}

// Returns a string with the just the appropriate color codes
func setColors(fg, bg Color) string {
	color := ""

	if fg != Clear && bg != Clear {
		// Adjust `fg` and `bg` for (+1) offset
		color = string(ColorA) + fmt.Sprintf("%v,%v", fg-1, bg-1)
	} else if fg != Clear {
		color = string(ColorA) + fmt.Sprintf("%v", fg-1)
	} else if bg != Clear {
		color = string(ColorA) + fmt.Sprintf(",%v", bg-1)
	}

	return color
}
