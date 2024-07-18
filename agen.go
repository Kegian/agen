package agen

import "go.uber.org/zap"

func Init(opts ...Option) error {
	settings := settings{}
	for _, o := range opts {
		if err := o(&settings); err != nil {
			return err
		}
	}

	if settings.config == nil {
		err := LoadConfig(&Config)
		if err != nil {
			return err
		}
	} else {
		err := LoadConfig(settings.config)
		if err != nil {
			return err
		}
		err = LoadConfig(settings.config.Config())
		if err != nil {
			return err
		}
		Config = settings.config.Config()
	}

	err := InitLogger(&Config.Log)
	if err != nil {
		return err
	}

	if Config.Sentry.Enabled {
		if err := InitSentry(Config.Environment, &Config.Sentry); err != nil {
			return err
		}
	}

	for _, c := range CurrentConfigs() {
		zap.L().Debug(c)
	}

	return nil
}

func Sync() error {
	if err := zap.L().Sync(); err != nil {
		return err
	}
	return nil
}

type settings struct {
	config Configurable
}

type Option func(*settings) error

func WithConfig(cfg Configurable) Option {
	return func(opts *settings) error {
		opts.config = cfg
		return nil
	}
}
