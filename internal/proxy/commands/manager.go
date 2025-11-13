package commands

import (
	"errors"
	"strings"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type CommandManager struct {
	m  *command.Manager
	l  *logger.Logger
	db *database.Database
	mm *manager.MultiManager
	tm *task.TaskManager
}

func Init(p *proxy.Proxy, l *logger.Logger, db *database.Database, mm *manager.MultiManager, tm *task.TaskManager) (*CommandManager, error) {
	cm := &CommandManager{
		m:  p.Command(),
		l:  l,
		db: db,
		mm: mm,
		tm: tm,
	}

	cm.registerCommands()
	return cm, nil
}

var (
	ComponentOnlyPlayersCommand = &component.Text{
		Content: "Only players can use this command.",
		S: component.Style{
			Color: color.Red,
		},
	}

	ErrOnlyPlayersCommand = errors.New("only players can use this command")

	ComponentOnlyPlayersSubCommand = &component.Text{
		Content: "Only players can use this sub command.",
		S: component.Style{
			Color: color.Red,
		},
	}

	ErrOnlyPlayersSubCommand = errors.New("only players can use this sub command")
)

func (cm *CommandManager) registerCommands() {
	cm.m.Register(cm.databaseCommand("database"))
	cm.m.Register(cm.databaseCommand("db"))
	cm.m.Register(cm.vanishCommand("vanish"))
	cm.m.Register(cm.vanishCommand("v"))
	cm.m.Register(cm.messageCommand("message"))
	cm.m.Register(cm.messageCommand("msg"))
	cm.m.Register(cm.kickCommand("kick"))
	cm.m.Register(cm.transferCommand("transfer"))
	cm.m.Register(cm.banCommand("ban"))
}

var (
	ErrTargetNotFound  = errors.New("target not found")
	TextTargetNotFound = util.TextWarn("Target not found.")

	ErrTargetIsOffline  = errors.New("target is offline")
	TextTargetIsOffline = util.TextWarn("Target is offline.")
)

func (cm *CommandManager) getMultiPlayerFromTarget(t string) (*multi.Player, error) {
	// target can be a player name or an uuid
	id, err := uuid.Parse(t)
	if err != nil {
		// try to get the id from a player name
		l := cm.mm.GetAllMultiPlayers()

		id = uuid.Nil
		for _, mp := range l {
			if t == mp.GetUsername() {
				id = mp.GetId()
				break
			}
		}

		if id == uuid.Nil {
			return nil, ErrTargetNotFound
		}
	}

	mp, err := cm.mm.GetMultiPlayer(id)
	if err != nil {
		if err == database.ErrDataNotFound {
			return nil, ErrTargetNotFound
		}

		return nil, err
	}

	return mp, nil
}

// suggests all multiplayers, online and offline.
// vanished players are hidden from non-privileged players
// Optional:
// 1. offline players can be hidden
// 2. own username and uuid can be hidden
func (cm *CommandManager) SuggestAllMultiPlayers(hideOfflinePlayers, hideOwn bool) brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		if len(r) < 1 {
			b.Suggest("type a username or UUID...")
			return b.Build()
		}

		// use list to get all names and ids
		var l []*multi.Player
		if hideOfflinePlayers {
			l = cm.mm.GetAllOnlinePlayers()
		} else {
			l = cm.mm.GetAllMultiPlayers()
		}

		hide_vanished := false
		own_username := ""
		p, ok := c.Source.(proxy.Player)
		if ok {
			mp, err := cm.mm.GetMultiPlayer(p.ID())
			if err != nil {
				cm.l.Error("suggest all multiplayers get multiplayer error", "error", err)
				return b.Build()
			}

			if hideOwn {
				own_username = mp.GetUsername()
			}

			if !mp.GetPermissionInfo().IsPrivileged() {
				hide_vanished = true
			}
		}

		for _, mp := range l {
			if hide_vanished && mp.IsVanished() {
				continue
			}

			name := mp.GetUsername()
			if name == own_username {
				continue
			}

			if strings.HasPrefix(strings.ToLower(name), r) {
				b.Suggest(name)
			}

			id := mp.GetId().String()
			if len(r) > 3 && strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}

func (cm *CommandManager) SuggestAllMultiProxies(hideOwn bool) brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase
		l := cm.mm.GetAllMultiProxies()

		for _, mp := range l {
			if hideOwn && mp == cm.mm.GetOwnerMultiProxy() {
				continue
			}

			id := mp.GetId().String()
			if strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}

func (cm *CommandManager) SuggestAllMultiBackends(hideOwn bool) brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		l := cm.mm.GetAllMultiBackends()

		var hiddenBackend *multi.Backend
		if hideOwn {
			hiddenBackendName := c.String("backendId")
			if hiddenBackendName != "" {
				id, err := uuid.Parse(hiddenBackendName)
				if err == nil {
					mp, err := cm.mm.GetMultiBackend(id)
					if err == nil {
						hiddenBackend = mp
					}
				}
			}
		}

		for _, mb := range l {
			if mb == hiddenBackend {
				continue
			}

			id := mb.GetId().String()
			if strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}

func (cm *CommandManager) SuggestAllMultiBackendsUnderProxy(hideOwn bool) brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		ownProxyName := c.String("proxyId")
		if ownProxyName == "" {
			return b.Build()
		}

		id, err := uuid.Parse(ownProxyName)
		if err != nil {
			return b.Build()
		}

		mp, err := cm.mm.GetMultiProxy(id)
		if err != nil {
			return b.Build()
		}

		l := cm.mm.GetAllMultiBackendsUnderMultiProxy(mp)

		var hiddenBackend *multi.Backend
		if hideOwn {
			hiddenBackendName := c.String("backendId")
			if hiddenBackendName != "" {
				id, err := uuid.Parse(hiddenBackendName)
				if err == nil {
					mp, err := cm.mm.GetMultiBackend(id)
					if err == nil {
						hiddenBackend = mp
					}
				}
			}
		}

		for _, mb := range l {
			if mb == hiddenBackend {
				continue
			}

			id := mb.GetId().String()
			if strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}

func (cm *CommandManager) getGatePlayerFromSource(source command.Source) proxy.Player {
	p, ok := source.(proxy.Player)
	if !ok {
		return nil
	}

	return p
}

func (cm *CommandManager) requireAdmin() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		p := cm.getGatePlayerFromSource(context.Source)

		if p != nil {
			mp, err := cm.mm.GetMultiPlayer(p.ID())
			if err != nil {
				return false
			}

			if mp.GetPermissionInfo().GetRole() == multi.RoleAdmin {
				return true
			}
		}

		return false
	})
}

func (cm *CommandManager) requireAdminOrModerator() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		p := cm.getGatePlayerFromSource(context.Source)

		if p != nil {
			mp, err := cm.mm.GetMultiPlayer(p.ID())
			if err != nil {
				return false
			}

			role := mp.GetPermissionInfo().GetRole()
			if role == multi.RoleAdmin || role == multi.RoleModerator {
				return true
			}
		}

		return false
	})
}

func (cm *CommandManager) requirePrivileged() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		p := cm.getGatePlayerFromSource(context.Source)

		if p != nil {
			mp, err := cm.mm.GetMultiPlayer(p.ID())
			if err != nil {
				return false
			}

			return mp.GetPermissionInfo().IsPrivileged()
		}

		return false
	})
}
