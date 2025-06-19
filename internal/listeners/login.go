package listeners

import (
	"fmt"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/ban"
	"github.com/team-vesperis/vesperis-mp/internal/mp/datasync"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onLogin(event *proxy.LoginEvent) {
	player := event.Player()

	if ban.IsPlayerBanned(player) {
		reason := ban.GetBanReason(player)

		if ban.IsPlayerPermanentlyBanned(player) {
			event.Deny(&component.Text{
				Content: "You are permanently banned from VesperisMC.",
				S: component.Style{
					Color: color.Red,
				},
				Extra: []component.Component{
					&component.Text{
						Content: "\n\nReason: " + reason,
						S: component.Style{
							Color: color.Gray,
						},
					},
				},
			})

		} else {
			timeComponent := component.Text{}
			duration := time.Until(ban.GetBanExpiration(player))

			if duration.Seconds() < 1 {
				timeComponent = component.Text{
					Content: "\n\nYour ban has just expired. Please try again in a moment.",
					S: component.Style{
						Color: color.Aqua,
					},
				}

			} else {
				hours := int(duration.Hours())
				days := hours / 24
				hours = hours % 24
				minutes := int(duration.Minutes()) % 60
				seconds := int(duration.Seconds()) % 60

				timeComponent = component.Text{
					Content: "\n\nYou are still banned for " + fmt.Sprintf("%d days, %d hours, %d minutes and %d seconds", days, hours, minutes, seconds),
					S: component.Style{
						Color: color.Aqua,
					},
				}
			}

			event.Deny(&component.Text{
				Content: "You are temporarily banned from VesperisMC",
				S: component.Style{
					Color: color.Red,
				},
				Extra: []component.Component{
					&component.Text{
						Content: "\n\nReason: " + reason,
						S: component.Style{
							Color: color.Gray,
						},
					},
					&timeComponent,
				},
			})
		}
	}

	p, _, _, _ := datasync.FindPlayerWithUUID(player.ID().String())
	if p != "" {
		event.Deny(&component.Text{
			Content: "You are already connected.",
			S: component.Style{
				Color: color.Red,
			},
		})
	}

}
