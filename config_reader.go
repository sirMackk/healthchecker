package healthchecker

import (
	"fmt"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Core         map[string]string
	HealthChecks []struct {
		Name     string
		Type     string
		Args     map[string]string
		Sinks    []map[string]map[string]string
		Interval int
	} `health-checks`
}

func ConfigFromYaml(fileContents []byte) (*Config, error) {
	c := Config{}
	err := yaml.Unmarshal(fileContents, &c)
	if err != nil {
		return nil, fmt.Errorf("Cannot create config from yaml: %s", string(fileContents))
	}
	return &c, nil
}
