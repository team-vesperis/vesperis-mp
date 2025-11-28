package util

import (
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
)

// https://lospec.com/palette-list/ribe-64

var ColorLightGreen, _ = color.Hex("#97ca63")

var StyleColorLightGreen = c.Style{
	Color: ColorLightGreen,
}

var ColorGreen, _ = color.Hex("#72a84d")

var StyleColorGreen = c.Style{
	Color: ColorGreen,
}

var ColorOrange, _ = color.Hex("#f6841c")

var StyleColorOrange = c.Style{
	Color: ColorOrange,
}

var ColorRed, _ = color.Hex("#e05340")

var StyleColorRed = c.Style{
	Color: ColorRed,
}

var ColorLightBlue, _ = color.Hex("#7bc2d4")

var StyleColorLightBlue = c.Style{
	Color: ColorLightBlue,
}

var ColorGray, _ = color.Hex("#a7a6b4")

var StyleColorGray = c.Style{
	Color: ColorGray,
}

var StyleMysterious = c.Style{
	Color:  ColorGray,
	Italic: c.True,
}

var ColorPink, _ = color.Hex("#ed5277")

var ColorPurple, _ = color.Hex("#7a4b61")

func ColorList(colors ...*color.RGB) []c.Style {
	var l []c.Style
	for _, col := range colors {
		l = append(l, c.Style{Color: col})
	}

	return l
}
