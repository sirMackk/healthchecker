package main

import (
	"flag"
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	hchecker "github.com/sirmackk/healthchecker"
)

var version string = "0.0.1"

func setupConfig(cfgFilePath string) (*hchecker.Config, error) {
	contents, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		return &hchecker.Config{}, fmt.Errorf("Cannot read config file '%s'", cfgFilePath)
	}
	config, err := hchecker.ConfigFromYaml(contents)
	if err != nil {
		return config, fmt.Errorf("Could not parse config file '%s'", cfgFilePath)
	}
	return config, nil
}

func populateRegistry(c *hchecker.Config, registry *hchecker.CheckRegistry) {
	httpTimeout, _ := strconv.Atoi(c.Core["HTTPTimeout"])
	httpChecker := hchecker.NewHTTPChecker(time.Duration(httpTimeout) * time.Second)

	registry.CheckConstructors["SimpleHTTPCheck"] = httpChecker.NewSimpleHTTPCheck
	registry.CheckConstructors["RegexpHTTPCheck"] = httpChecker.NewRegexpHTTPCheck
	registry.SinkConstructors["ConsoleSink"] = hchecker.NewConsoleSink
}



func main() {
	var cfgFilePath = flag.String("cfgFilePath", "config.yaml", "Absolute path to yaml config file")
	var printVersion = flag.Bool("version", false, "Print version")
	var debug = flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *printVersion {
		log.Infof("Version: %s", version)
		os.Exit(0)
	}

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	if *debug {
		log.Info("Enabling debug-level logging")
		log.SetLevel(log.DebugLevel)
	}

	config, err := setupConfig(*cfgFilePath)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	registry := hchecker.NewCheckRegistry()
	populateRegistry(config, registry)
	registry.RegisterHealthChecks(config)
	registry.StartRunning()
}
