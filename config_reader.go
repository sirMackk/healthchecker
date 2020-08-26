package healthchecker

import (
	"fmt"
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

func ConfigFromYaml(fileContents []byte) (*Config, error) {
	c := Config{}
	err := yaml.Unmarshal(fileContents, &c)
	if err != nil {
		return &Config{}, fmt.Errorf("Cannot create config from yaml: %s", string(fileContents))
	}
	return &c, nil
}
