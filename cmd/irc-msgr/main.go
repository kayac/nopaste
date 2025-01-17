package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/kayac/nopaste"
)

func main() {
	var config string
	var logLevel string
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

	flag.Parse()
	err := nopaste.RunMsgr(context.Background(), config)
	if err != nil {
		panic(err)
	}
}
