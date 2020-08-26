package healthchecker

import (
	"encoding/json"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Core         map[string]string
	HealthChecks []struct {
		Type  string
		Args  map[string]string
		Sinks []map[string]map[string]string
	} `health-checks`
}

func ConfigFromJson(fileContents []byte) *Config {
	var config Config
	err := json.Unmarshal(fileContents, &config)
	if err != nil {
		panic("Cannot create Config from json")
	}
	return &config
}

func ConfigFromYaml(fileContents []byte) *Config {
	c := Config{}
	err := yaml.Unmarshal(fileContents, &c)
	if err != nil {
		panic("Cannot create config from yaml")
	}
	return &c
}
