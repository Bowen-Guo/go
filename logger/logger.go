package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var singleton *zap.SugaredLogger
var once sync.Once

func InitStdoutLogger() {
	once.Do(func() { singleton = getStdoutLogger() })
}

func GetLogger(name string) (*zap.SugaredLogger, error) {
	if singleton == nil {
		return nil, fmt.Errorf("Empty logger. InitStdoutLogger method should have been invoked first.")
	}

	if name == "" {
		return singleton, nil
	}

	return singleton.Named(name), nil
}

func getStdoutLogger() *zap.SugaredLogger {
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	zapConf := zap.NewProductionEncoderConfig()
	level := zap.NewAtomicLevel()
	stdoutCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConf),
		stdoutSyncer,
		level,
	)

	appInsightsCore := NewAppInsightsCore()
	allCores := []zapcore.Core{stdoutCore, appInsightsCore}
	teeCore := zapcore.NewTee(allCores...)

	return zap.New(teeCore, zap.AddCaller()).Sugar()
}
