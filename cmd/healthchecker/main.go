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
	return hchecker.ConfigFromJson(contents)
}

func main() {
	cfgFilePath := "exampleConfig.json"
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
	for hcName, hcDetails := range config.HealthChecks {
		checkType := hcDetails.Type
		checkArgs := hcDetails.Args
		sinks := make([]hchecker.Sink, 0)
		for _, s := range hcDetails.Sinks {
			for sinkName, sinkArgs := range s {
	// ../src/github.com/sirmackk/healthchecker/cmd/healthchecker/main.go:40:51: cannot use sinkArgs (type []string) as type bool in argument to registry.Sinks[sinkName]
	//../src/github.com/sirmackk/healthchecker/cmd/healthchecker/main.go:40:51: invalid use of ... in call to registry.Sinks[sinkName]
				sinks = append(sinks, registry.Sinks[sinkName](sinkArgs...))
			}
		}
		registry.NewCheck(checkType, checkArgs, sinks)
	}
	// TODO create checkRunner
	registry.Checks[0].Run()
}
