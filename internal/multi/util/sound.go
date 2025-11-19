package util

import (
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/edition/java/sound"
)

var SoundThunder sound.Sound = sound.NewSound("entity.lightning_bolt.thunder", sound.SourceUI).WithPitch(0.8)

func PlayThunderSound(p proxy.Player) {
	sound.Play(p, SoundThunder, p)
}

var LevelUpSound sound.Sound = sound.NewSound("entity.level.up", sound.SourcePlayer).WithVolume(0.3)

func PlayLevelUpSound(p proxy.Player) {
	sound.Play(p, LevelUpSound, p)
}
