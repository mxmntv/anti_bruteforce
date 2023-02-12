package main

import (
	"flag"
	"fmt"
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
		log.Fatalf("config error: %s", err)
	}
	fmt.Println(cfg)
	if err := app.Run(cfg); err != nil {
		os.Exit(1)
	}
}
