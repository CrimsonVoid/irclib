package irclib

const (
	// Color      = "\x03"
	Bold          = "\x02"
	Italic        = "\x09"
	Reset         = "\x0F"
	StrikeThrough = "\x13"
	Underline     = "\x15"
	Reverse       = "\x16"
	Underline2    = "\x1F"
)

const (
	White = "\x03" + string(iota) // Also color reset in most clients
	Black
	Blue
	Green
	LightRed
	Red
	Magenta
	Orange
	Yellow
	LightGreen
	Cyan
	LightCyan
	LightBlue
	LightMagenta
	Gray
	LightGray
)
