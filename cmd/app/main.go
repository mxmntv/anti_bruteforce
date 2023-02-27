package main

import (
	"flag"
	"log"
	"os"

	"github.com/mxmntv/anti_bruteforce/config"
	"github.com/mxmntv/anti_bruteforce/internal/app"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./config/config.yml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg, err := config.NewConfig(configFile)
	if err != nil {
		log.Fatalf("[main] parse config error: %s", err.Error())
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("[main] app start failed error: %s", err.Error())
		os.Exit(1)
	}
}
