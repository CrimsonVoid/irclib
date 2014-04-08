package irclib

import (
	"log"
	"os"
)

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
	CC_White        = "00" // Also color reset in most clients
	CC_Black        = "01"
	CC_Blue         = "02"
	CC_Green        = "03"
	CC_LightRed     = "04"
	CC_Red          = "05"
	CC_Magenta      = "06"
	CC_Orange       = "07"
	CC_Yellow       = "08"
	CC_LightGreen   = "09"
	CC_Cyan         = "10"
	CC_LightCyan    = "11"
	CC_LightBlue    = "12"
	CC_LightMagenta = "13"
	CC_Gray         = "14"
	CC_LightGray    = "15"
)

const (
	CC_FgWhite        = CC_Color + CC_White
	CC_FgBlack        = CC_Color + CC_Black
	CC_FgBlue         = CC_Color + CC_Blue
	CC_FgGreen        = CC_Color + CC_Green
	CC_FgLightRed     = CC_Color + CC_LightRed
	CC_FgRed          = CC_Color + CC_Red
	CC_FgMagenta      = CC_Color + CC_Magenta
	CC_FgOrange       = CC_Color + CC_Orange
	CC_FgYellow       = CC_Color + CC_Yellow
	CC_FgLightGreen   = CC_Color + CC_LightGreen
	CC_FgCyan         = CC_Color + CC_Cyan
	CC_FgLightCyan    = CC_Color + CC_LightCyan
	CC_FgLightBlue    = CC_Color + CC_LightBlue
	CC_FgLightMagenta = CC_Color + CC_LightMagenta
	CC_FgGray         = CC_Color + CC_Gray
	CC_FgLightGray    = CC_Color + CC_LightGray

	CC_BgWhite        = CC_Color + "," + CC_White
	CC_BgBlack        = CC_Color + "," + CC_Black
	CC_BgBlue         = CC_Color + "," + CC_Blue
	CC_BgGreen        = CC_Color + "," + CC_Green
	CC_BgLightRed     = CC_Color + "," + CC_LightRed
	CC_BgRed          = CC_Color + "," + CC_Red
	CC_BgMagenta      = CC_Color + "," + CC_Magenta
	CC_BgOrange       = CC_Color + "," + CC_Orange
	CC_BgYellow       = CC_Color + "," + CC_Yellow
	CC_BgLightGreen   = CC_Color + "," + CC_LightGreen
	CC_BgCyan         = CC_Color + "," + CC_Cyan
	CC_BgLightCyan    = CC_Color + "," + CC_LightCyan
	CC_BgLightBlue    = CC_Color + "," + CC_LightBlue
	CC_BgLightMagenta = CC_Color + "," + CC_LightMagenta
	CC_BgGray         = CC_Color + "," + CC_Gray
	CC_BgLightGray    = CC_Color + "," + CC_LightGray
)

var (
	consLog = log.New(os.Stdout, "", 0)
)
