package main

import (
	"flag"

	"github.com/kayac/nopaste"
)

func main() {
	var config string
	flag.StringVar(&config, "c", "config.yaml", "path to config.yaml")
	flag.StringVar(&config, "config", "config.yaml", "path to config.yaml")
	flag.Parse()
	err := nopaste.RunMsgr(config)
	if err != nil {
		panic(err)
	}
}
