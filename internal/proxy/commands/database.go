package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

// The database command is used for testing parts of the database.
func (cm CommandManager) databaseCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("get").
			Then(brigodier.Argument("key", brigodier.SingleWord).
				Executes(command.Command(func(c *command.Context) error {
					var val string
					err := cm.db.GetData(c.String("key"), &val)
					if err != nil {
						if err == database.ErrDataNotFound {
							c.SendMessage(util.TextWarn("No value found in the database."))
							return nil
						}
						c.SendMessage(util.TextInternalError("Could not get value from database.", err))
						return err
					}

					c.SendMessage(util.TextAlternatingColors("Returned value: ", val))
					return nil
				})))).
		Then(brigodier.Literal("set").
			Then(brigodier.Argument("key", brigodier.SingleWord).
				Then(brigodier.Argument("value", brigodier.StringPhrase).
					Executes(command.Command(func(c *command.Context) error {
						err := cm.db.SetData(c.String("key"), c.String("value"))
						if err != nil {
							c.SendMessage(util.TextInternalError("Could not set value from database.", err))
							return err
						}

						c.SendMessage(util.TextAlternatingColors("Successfully set value in database."))
						return nil
					})))))
}
