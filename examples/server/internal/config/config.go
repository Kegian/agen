package config

import "github.com/Kegian/agen"

type Config struct {
	*agen.BaseConfig

	CustomSetting int64 `cfg:"CUSTOM_SETTING,ALIAS__NAME" default:"5"`
}

var Cfg = Config{}
