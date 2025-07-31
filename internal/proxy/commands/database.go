package commands

import (
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

// The database command is used for testing parts of the database.
func (cm CommandManager) databaseCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("get").
			Then(brigodier.Argument("key", brigodier.SingleWord).
				Executes(command.Command(func(c *command.Context) error {
					v, err := cm.db.GetData(c.String("key"))
					val, _ := v.(string)

					if err != nil {
						c.SendMessage(&component.Text{
							Content: "Error getting value from database: " + err.Error(),
							S:       component.Style{Color: color.Red},
						})
					} else {

						c.SendMessage(&component.Text{
							Content: "Returned value: " + val,
							S:       component.Style{Color: color.Green},
						})

					}

					return nil
				})))).
		Then(brigodier.Literal("set").
			Then(brigodier.Argument("key", brigodier.SingleWord).
				Then(brigodier.Argument("value", brigodier.SingleWord).
					Executes(command.Command(func(c *command.Context) error {
						err := cm.db.SetData(c.String("key"), c.String("value"))
						if err != nil {
							c.SendMessage(&component.Text{
								Content: "Error setting value in database: " + err.Error(),
								S:       component.Style{Color: color.Red},
							})
						} else {
							c.SendMessage(&component.Text{
								Content: "Successfully set value in database",
								S:       component.Style{Color: color.Green},
							})
						}

						return nil
					})))))
}
