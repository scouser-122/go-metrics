package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()
var Sugar = Log.Sugar()

func Initialize(level string, environment string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	var cfg zap.Config
	switch environment {
	case "dev":
		cfg = zap.NewDevelopmentConfig()
	case "prod":
		cfg = zap.NewProductionConfig()
	default:
		return fmt.Errorf("unknown environment for logging config: %s", environment)
	}
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	Sugar = Log.Sugar()
	return nil
}
