package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	hc "github.com/sirmackk/healthchecker"
)

func setupConfig(cfgFilePath) *Config {
	contents, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		panic(fmt.Sprintf("Cannot read config file %s", cfgFilePath))
	}
	return ConfigFromJson(contents)
}

func main() {
	cfgFilePath := "exampleConfig.json"
	config := SetupConfig(cfgFilePath)
	registry := NewCheckRegistry()

	// register available health check modules
	httpTimeout, _ := strconv.Atoi(config.Core["HTTPTimeout"]0
	httpChecker := NewHTTPChecker(httpTimeout * time.Second)
	registry.CheckFuncs["SimpleHTTPCheck"] = httpChecker.NewSimpleHTTPCheck
	registry.CheckFuncs["RegexpHTTPCheck"] = httpChecker.NewRegexpHTTPCheck
	registry.Sinks["ConsoleSink"] = NewConsoleSink

	// register health checks
	// TODO unfuck this mess into neat code
	for _, hc := range config.HealthChecks:
		for _, checkDetails := range hc {
			checkType := checkDetails["type"]
			checkArgs := checkDetails["args"]
			sinks := make([]*Sink, 0)
			for _, s := range checkDetails["sinks"] {
				for sinkName, sinkArgs := range s {
					sinks = append(sinks, registry.Sinks[sinkName](sinksArgs))
				}
			}
			registry.NewCheck(checkType, CheckArgs, sinks)
		}
	}
	// TODO create checkRunner

}
