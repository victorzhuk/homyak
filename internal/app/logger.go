package app

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/victorzhuk/homyak/internal/logger"
)

const (
	logKeyTimestamp = "timestamp"
	logKeyMessage   = "message"
)

func SetupLogger(cfg *config) {
	zc := zap.NewProductionEncoderConfig()
	zc.TimeKey = logKeyTimestamp
	zc.MessageKey = logKeyMessage
	zc.EncodeTime = zapcore.ISO8601TimeEncoder
	zc.EncodeLevel = zapcore.CapitalColorLevelEncoder

	stdout := zapcore.AddSync(os.Stdout)
	ce := zapcore.NewJSONEncoder(zc)
	if cfg.APP.Env == EnvLocal {
		ce = zapcore.NewConsoleEncoder(zc)
	}

	lvl := zapcore.InfoLevel
	if cfg.APP.IsDebug {
		lvl = zapcore.DebugLevel
	}
	core := zapcore.NewCore(ce, stdout, lvl)
	logger.Init(core)
}
