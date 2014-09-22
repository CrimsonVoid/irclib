package styles

import (
	"fmt"
)

type Attrib string

const (
	ColorA        Attrib = "\x03"
	Bold          Attrib = "\x02"
	Italic        Attrib = "\x09"
	Reset         Attrib = "\x0F"
	StrikeThrough Attrib = "\x13"
	Underline     Attrib = "\x15"
	Reverse       Attrib = "\x16"
	Underline2    Attrib = "\x1F"
)

// Paints the text with `Attrib` and formats the text with Sprintf
func (self Attrib) Paint(format string, a ...interface{}) string {
	return string(self) + fmt.Sprintf(format, a...) + string(Reset)
}
