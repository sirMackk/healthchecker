package healthchecker

import (
	"encoding/json"
)

type Config struct {
	Core         map[string]string
	Healthchecks map[string]map[string]string
}

func ConfigFromJson(fileContents []byte) *Config {
	var config Config
	err := json.Unmarshal(fileContents, &config)
	if err != nil {
		panic("Cannot create Config from json")
	}
	return &config
}

// TODO: read from ini/yaml
