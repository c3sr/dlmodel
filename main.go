package main

import (
	"github.com/c3sr/dlmodel/cmd"
	"github.com/c3sr/config"
)

func main() {
	cmd.Execute()
}

func init() {
	config.Init(
		config.AppName("carml"),
		config.DebugMode(true),
		config.VerboseMode(true),
	)
}
