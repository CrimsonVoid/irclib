package styles

import (
	"fmt"
)

type Painter interface {
	// Paint a style with Print-style formatting
	Paint(a ...interface{}) string
	// Paint a style with Printf-style formatting
	Paintf(string, ...interface{}) string
	// Paint a style with Println-style formatting
	Paintln(a ...interface{}) string
}

// Retuns an anonymous function which formats text with Sprintf
func MkPainter(p Painter) func(string, ...interface{}) string {
	return func(format string, a ...interface{}) string {
		return p.Paintf(format, a...)
	}
}

type Style struct {
	Fg, Bg     Color
	Attributes []Attrib
}

// Sets the `Style` and formats the text with Sprint
func (self *Style) Paint(a ...interface{}) string {
	return self.setStyles() + fmt.Sprint(a...) + string(Reset)
}

// Sets the `Style` and formats the text with Sprintf
func (self *Style) Paintf(format string, a ...interface{}) string {
	return self.setStyles() + fmt.Sprintf(format, a...) + string(Reset)
}

// Sets the `Style` and formats the text with Sprintln
func (self *Style) Paintln(a ...interface{}) string {
	// Do not use `fmt.Println()` becuase '\n' needs to be after `Reset`
	return self.Paint(a...) + "\n"
}

// Returns a string with the just the appropriate colors and attributes
func (self *Style) setStyles() string {
	styles := setColors(self.Fg, self.Bg)

	for _, s := range self.Attributes {
		styles += string(s)
	}

	return styles
}
