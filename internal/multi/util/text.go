package util

import c "go.minekube.com/common/minecraft/component"

func TextInternalError(message string, err error) *c.Text {
	return &c.Text{
		Content: message,
		S: c.Style{
			Color:      ColorRed,
			HoverEvent: c.ShowText(&c.Text{Content: "Internal error: " + err.Error(), S: StyleColorRed}),
		},
	}
}

func TextError(message string) *c.Text {
	return &c.Text{
		Content: message,
		S:       StyleColorRed,
	}
}

func TextWarn(message string) *c.Text {
	return &c.Text{
		Content: message,
		S:       StyleColorOrange,
	}
}

func TextSuccess(values ...string) *c.Text {
	components := make([]c.Component, 0, len(values))
	colors := []c.Style{StyleColorLightGreen, StyleColorCyan}
	for i, v := range values {
		components = append(components, &c.Text{
			Content: v,
			S:       colors[i%len(colors)],
		})
	}

	if len(components) == 0 {
		return &c.Text{}
	}

	first := components[0].(*c.Text)
	extra := make([]c.Component, 0, len(components)-1)
	extra = append(extra, components[1:]...)
	first.Extra = extra
	return first
}
