package irclib

const (
	CC_Color         = "\x03"
	CC_Bold          = "\x02"
	CC_Italic        = "\x09"
	CC_Reset         = "\x0F"
	CC_StrikeThrough = "\x13"
	CC_Underline     = "\x15"
	CC_Reverse       = "\x16"
	CC_Underline2    = "\x1F"
)

const (
	CC_White = CC_Color + string(iota) // Also color reset in most clients
	CC_Black
	CC_Blue
	CC_Green
	CC_LightRed
	CC_Red
	CC_Magenta
	CC_Orange
	CC_Yellow
	CC_LightGreen
	CC_Cyan
	CC_LightCyan
	CC_LightBlue
	CC_LightMagenta
	CC_Gray
	CC_LightGray
)
