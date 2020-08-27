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

func registerHealthChecks(c *hchecker.Config, registry *hchecker.CheckRegistry) {
	for _, hc := range c.HealthChecks {
		log.Debugf("Creating sinks for %s", hc.Name)
		sinks, err := createSinks(hc.Sinks, registry)
		if err != nil {
			log.Errorf("Could not register %s due to: %s", hc.Name, err)
			continue
		}
		_, err = registry.NewCheck(hc.Name, hc.Type, hc.Args, hc.Interval, sinks)
		if err != nil {
			log.Errorf("Could not register %s due to: %s", hc.Name, err)
		}
	}
}

func createSinks(sinkConfig []map[string]map[string]string, registry *hchecker.CheckRegistry) ([]hchecker.Emitter, error) {
	sinks := make([]hchecker.Emitter, 0)
	for _, sink := range sinkConfig {
		for sinkName, sinkArgs := range sink {
			newSink, err := registry.SinkConstructors[sinkName](sinkArgs)
			if err != nil {
				return sinks, fmt.Errorf("Unable to create sink '%s' with args: %v", sinkName, sinkArgs)
			}
			sinks = append(sinks, newSink)
		}
	}
	return sinks, nil
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

	// register available health check modules
	populateRegistry(config, registry)

	// register health checks
	registerHealthChecks(config, registry)

	registry.StartRunning()
}
