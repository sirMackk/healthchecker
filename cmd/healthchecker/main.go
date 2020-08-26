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

func main() {
	cfgFilePath := "exampleConfig.yaml"
	config := setupConfig(cfgFilePath)
	registry := hchecker.NewCheckRegistry()

	// register available health check modules
	httpTimeout, _ := strconv.Atoi(config.Core["HTTPTimeout"])
	httpChecker := hchecker.NewHTTPChecker(time.Duration(httpTimeout) * time.Second)
	registry.CheckFuncs["SimpleHTTPCheck"] = httpChecker.NewSimpleHTTPCheck
	registry.CheckFuncs["RegexpHTTPCheck"] = httpChecker.NewRegexpHTTPCheck
	registry.Sinks["ConsoleSink"] = hchecker.NewConsoleSink

	// register health checks
	// TODO unfuck this mess into neat code
	for _, hcDetails := range config.HealthChecks {
		checkType := hcDetails.Type
		checkArgs := hcDetails.Args
		sinks := make([]hchecker.Sink, 0)
		for _, s := range hcDetails.Sinks {
			for sinkName, sinkArgs := range s {
				sinks = append(sinks, registry.Sinks[sinkName](sinkArgs[0]))
			}
		}
		registry.NewCheck(checkType, checkArgs, sinks)
	}
	// TODO create checkRunner
	registry.Checks[0].Run()
}
