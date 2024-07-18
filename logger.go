package agen

import (
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level    string `cfg:"LOG_LEVEL" default:"info"`
	Encoding string `cfg:"LOG_ENCODING" default:"json"`
}

func InitLogger(cfg *LoggerConfig) error {
	var (
		encoding string
		isPretty = false
	)
	switch strings.ToLower(cfg.Encoding) {
	case "json":
		encoding = "json"
	case "console":
		encoding = "console"
	case "pretty":
		encoding = "console"
		isPretty = true
	default:
		encoding = "json"
	}

	level, err := zap.ParseAtomicLevel(strings.ToLower(cfg.Level))
	if err != nil {
		return err
	}

	zapCfg := zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if encoding == "console" {
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	}

	if isPretty {
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.EncoderConfig.EncodeTime = prettyTimeEncoder
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return err
	}

	_ = zap.ReplaceGlobals(logger)

	return nil
}

const prettyTimeLayout = "2006-01-02 15:04:05"

func prettyTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, prettyTimeLayout)
		return
	}

	enc.AppendString(t.Format(prettyTimeLayout))
}
