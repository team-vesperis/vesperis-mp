package util

import (
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
)

var ColorLightGreen, _ = color.Hex("#38ff38")

var StyleColorLightGreen = c.Style{
	Color: ColorLightGreen,
}

var ColorGreen, _ = color.Hex("#1fac1f")

var StyleColorGreen = c.Style{
	Color: ColorGreen,
}

var ColorOrange, _ = color.Hex("#ff9100")

var StyleColorOrange = c.Style{
	Color: ColorOrange,
}

var ColorRed, _ = color.Hex("#ff3733")

var StyleColorRed = c.Style{
	Color: ColorRed,
}

var ColorCyan, _ = color.Hex("#12dbe2")

var StyleColorCyan = c.Style{
	Color: ColorCyan,
}

var ColorGray, _ = color.Hex("#999999")

var StyleColorGray = c.Style{
	Color: ColorGray,
}

var StyleMysterious = c.Style{
	Color:  ColorGray,
	Italic: c.True,
}
