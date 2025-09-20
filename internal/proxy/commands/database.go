package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	. "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

// The database command is used for testing parts of the database.
func (cm CommandManager) databaseCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("get").
			Then(brigodier.Argument("key", brigodier.SingleWord).
				Executes(command.Command(func(c *command.Context) error {
					v, err := cm.db.GetData(c.String("key"))
					if err == database.ErrDataNotFound {
						c.SendMessage(&Text{
							Content: "No value found in the database.",
							S:       util.StyleColorOrange,
						})
						return nil
					}

					val, _ := v.(string)

					if err != nil {
						c.SendMessage(&Text{
							Content: "Could not get value from database.",
							S: Style{
								Color:      util.ColorRed,
								HoverEvent: ShowText(&Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
							},
						})
						return err
					}

					c.SendMessage(&Text{
						Content: "Returned value: " + val,
						S:       util.StyleColorLightGreen,
					})

					return nil
				})))).
		Then(brigodier.Literal("set").
			Then(brigodier.Argument("key", brigodier.SingleWord).
				Then(brigodier.Argument("value", brigodier.StringPhrase).
					Executes(command.Command(func(c *command.Context) error {
						err := cm.db.SetData(c.String("key"), c.String("value"))
						if err != nil {
							c.SendMessage(&Text{
								Content: "Could not set value from database.",
								S: Style{
									Color:      util.ColorRed,
									HoverEvent: ShowText(&Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
								},
							})
							return err
						}

						c.SendMessage(&Text{
							Content: "Successfully set value in database.",
							S:       util.StyleColorLightGreen,
						})

						return nil
					})))))
}
