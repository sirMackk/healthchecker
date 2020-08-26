package main

import (
	"flag"
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
	"time"

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
	for hcName, hcDetails := range c.HealthChecks {
		sinks, err := createSinks(hcDetails.Sinks, registry)
		if err != nil {
			fmt.Printf("Could not register %s due to: %s\n", hcName, err)
			continue
		}
		_, err = registry.NewCheck(hcDetails.Type, hcDetails.Args, hcDetails.Interval, sinks)
		if err != nil {
			fmt.Printf("Could not register %s due to %s\n", hcName, err)
		}
	}
}

func createSinks(sinkConfig []map[string]map[string]string, registry *hchecker.CheckRegistry) ([]hchecker.Sink, error) {
	sinks := make([]hchecker.Sink, 0)
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

	flag.Parse()

	if *printVersion {
		fmt.Printf("version: %s\n", version)
		os.Exit(0)
	}

	config, err := setupConfig(*cfgFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	registry := hchecker.NewCheckRegistry()

	// register available health check modules
	populateRegistry(config, registry)

	// register health checks
	registerHealthChecks(config, registry)

	registry.StartRunning()
}
