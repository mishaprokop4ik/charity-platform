package zlog

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

var Log logr.Logger

func Init() {
	conf := zap.NewDevelopmentConfig()
	conf.OutputPaths = append(conf.OutputPaths, "stderr")
	zapLog, err := conf.Build()
	if err != nil {
		panic(fmt.Sprintf("init zap log err: %s;", err))
	}

	Log = zapr.NewLogger(zapLog)
}
