package main

import (
	"flag"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/kayac/nopaste"
)

func main() {
	var config, logLevel string
	flag.StringVar(&config, "c", "config.yaml", "path to config.yaml")
	flag.StringVar(&config, "config", "config.yaml", "path to config.yaml")
	flag.StringVar(&logLevel, "log-level", "info", "log level")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	err := nopaste.Run(config)
	if err != nil {
		panic(err)
	}
}
