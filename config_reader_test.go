package healthchecker

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestSimpleYamlConfigRead(t *testing.T) {
	exampleYaml, _ := ioutil.ReadFile("fixtures/exampleConfig.yaml")
	config, _ := ConfigFromYaml(exampleYaml)
	if config.Core["HTTPTimeout"] != "10" {
		t.Errorf("Fail: %s\n%v", exampleYaml, config)
	}
	blogCheck := config.HealthChecks[0]
	if blogCheck.Type != "SimpleHTTPCheck" ||
		!reflect.DeepEqual(blogCheck.Args, map[string]string{"url": "http://mattscodecave.com"}) ||
		!reflect.DeepEqual(blogCheck.Sinks[0],
			map[string]map[string]string{"ConsoleSink": map[string]string{"useStdout": "true"}}) {
		t.Errorf("Fail: %s\n%v", exampleYaml, config)
	}
}

func TestSimpleYamlConfigReadFailure(t *testing.T) {
	badYaml := `
	---
	something: 10
	bla bla bla`
	config, err := ConfigFromYaml([]byte(badYaml))
	if err == nil {
		t.Errorf("Was supposed to fail on bad yaml, got: %v", config)
	}
}
