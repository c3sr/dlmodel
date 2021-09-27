package cmd

import (
	"github.com/c3sr/config"
	"github.com/c3sr/logger"
	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Entry
)

func init() {
	config.AfterInit(func() {
		log = logger.New().WithField("pkg", "dlmodel")
	})
	config.Init(
		config.AppName("carml"),
		config.DebugMode(true),
		config.VerboseMode(true),
	)
}
