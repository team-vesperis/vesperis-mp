package config

import (
	"errors"
	"slices"
)

type Mode string

const (
	Mode_Default    Mode = "default"
	Mode_Kubernetes Mode = "kubernetes"
)

var ErrIncorrectMode = errors.New("incorrect mode")

var AllowedModes = []Mode{
	Mode_Default,
	Mode_Kubernetes,
}

// check config which mode is enabled. if wrong mode is set, change to default
func (c *Config) getMode() (Mode, error) {
	s := c.v.GetString("mode")
	m, err := GetMode(s)
	if err != nil {
		c.v.Set("mode", "default")
		return m, err
	}

	return m, nil
}

func GetMode(s string) (Mode, error) {
	m := Mode(s)
	if !slices.Contains(AllowedModes, m) {
		return Mode(""), ErrIncorrectMode
	}

	return m, nil
}
