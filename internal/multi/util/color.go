package util

import (
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	. "go.minekube.com/common/minecraft/component"
)

var ColorLightGreen, _ = color.Hex("#2ec52e")

var StyleColorLightGreen = Style{
	Color: ColorLightGreen,
}

var ColorOrange, _ = color.Hex("#ff8c00")

var StyleColorOrange = Style{
	Color: ColorOrange,
}

var ColorRed, _ = color.Hex("#f72421")

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

var ColorCyan, _ = color.Hex("#0da8ad")

var StyleColorCyan = Style{
	Color: ColorCyan,
}

var ColorGray, _ = color.Hex("#7a7878")

var StyleColorGray = Style{
	Color: ColorGray,
}

var StyleMysterious = Style{
	Color:  ColorGray,
	Italic: True,
}
