package listeners

import (
	"encoding/hex"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

var rp proxy.ResourcePackInfo

func (lm *ListenerManager) initResourcePack() error {
	var val string
	err := lm.db.GetData("resourcepack_hash", &val)
	if err != nil {
		lm.l.Error("get resourcepack hash from database error", "error", err)
		return err
	}

	hash, err := hex.DecodeString(val)
	if err != nil {
		lm.l.Error("decode string to bytes error", "error", err)
		return err
	}

	var url string
	err = lm.db.GetData("resourcepack_url", &url)
	if err != nil {
		lm.l.Error("get resourcepack url from database error", "error", err)
		return err
	}

	prompt := util.TextWarn("Vesperis requires you to enable our resourcepack.")

	rp = proxy.ResourcePackInfo{
		ID:          uuid.New(),
		URL:         url,
		Hash:        hash,
		Prompt:      prompt,
		ShouldForce: true,
	}

	return nil
}

func (lm *ListenerManager) sendResourcePack(e *proxy.ServerPostConnectEvent) {
	go e.Player().SendResourcePack(rp)
}
