package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (lm *ListenerManager) onChatMessage(e *proxy.PlayerChatEvent) {
	p := e.Player()
	e.SetAllowed(false)
	for _, mp := range lm.mm.GetAllOnlinePlayers(true) {
		if mp.GetProxy() == nil {
			continue
		}

		tr := lm.tm.BuildTask(tasks.NewMessageTask(mp.GetId(), mp.GetProxy().GetId(), util.ComponentToString(util.TextAlternatingColors(util.ColorList(util.ColorLightBlue, util.ColorWhite), "["+p.Username()+"]", ": "+e.Message()))))
		if !tr.IsSuccessful() {
			lm.l.Info("player chat event send message task to online player error", "originalPlayerId", p.ID(), "targetPlayerId", mp.GetId(), "error", tr.GetInfo())
		}
	}
}
