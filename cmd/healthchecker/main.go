package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	hchecker "github.com/sirmackk/healthchecker"
)

func setupConfig(cfgFilePath string) *hchecker.Config {
	contents, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		panic(fmt.Sprintf("Cannot read config file %s", cfgFilePath))
	}
	return hchecker.ConfigFromYaml(contents)
}

func populateRegistry(c *hchecker.Config, registry *hchecker.CheckRegistry) {
	httpTimeout, _ := strconv.Atoi(c.Core["HTTPTimeout"])
	httpChecker := hchecker.NewHTTPChecker(time.Duration(httpTimeout) * time.Second)

	registry.CheckConstructors["SimpleHTTPCheck"] = httpChecker.NewSimpleHTTPCheck
	registry.CheckConstructors["RegexpHTTPCheck"] = httpChecker.NewRegexpHTTPCheck
	registry.SinkConstructors["ConsoleSink"] = hchecker.NewConsoleSink
}

func registerHealthChecks(c *hchecker.Config, registry *hchecker.CheckRegistry) {
	for _, hcDetails := range c.HealthChecks {
		sinks := createSinks(hcDetails.Sinks, registry)
		registry.NewCheck(hcDetails.Type, hcDetails.Args, sinks)
	}
}

func createSinks(sinkConfig []map[string]map[string]string, registry *hchecker.CheckRegistry) []hchecker.Sink {
	sinks := make([]hchecker.Sink, 0)
	for _, sink := range sinkConfig {
		for sinkName, sinkArgs := range sink {
			sinks = append(sinks, registry.SinkConstructors[sinkName](sinkArgs))
		}
	}
	return sinks
}


func main() {
	cfgFilePath := "exampleConfig.yaml"
	config := setupConfig(cfgFilePath)
	registry := hchecker.NewCheckRegistry()

	// register available health check modules
	populateRegistry(config, registry)

	// register health checks
	registerHealthChecks(config, registry)

	// TODO create checkRunner
	registry.Checks[0].Run()
}
