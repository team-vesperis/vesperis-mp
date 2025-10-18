package util

import (
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	. "go.minekube.com/common/minecraft/component"
)

var ColorLightGreen, _ = color.Hex("#38ff38")

var StyleColorLightGreen = Style{
	Color: ColorLightGreen,
}

var ColorGreen, _ = color.Hex("#1fac1f")

var StyleColorGreen = Style{
	Color: ColorGreen,
}

var ColorOrange, _ = color.Hex("#ff9100")

var StyleColorOrange = Style{
	Color: ColorOrange,
}

var ColorRed, _ = color.Hex("#ff3733")

var StyleColorRed = Style{
	Color: ColorRed,
}

func TextInternalError(message string, err error) *component.Text {
	return &component.Text{
		Content: message,
		S:       StyleInternalError(err),
	}
}

func StyleInternalError(err error) component.Style {
	return component.Style{
		Color:      ColorRed,
		HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: StyleColorRed}),
	}
}

var ColorCyan, _ = color.Hex("#12dbe2")

var StyleColorCyan = Style{
	Color: ColorCyan,
}

var ColorGray, _ = color.Hex("#999999")

var StyleColorGray = Style{
	Color: ColorGray,
}

var StyleMysterious = Style{
	Color:  ColorGray,
	Italic: True,
}
