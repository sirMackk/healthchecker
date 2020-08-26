package healthchecker

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestSimpleYamlConfigRead(t *testing.T) {
	exampleYaml, _ := ioutil.ReadFile("fixtures/exampleConfig.yaml")
	config := ConfigFromYaml(exampleYaml)
	if config.Core["HTTPTimeout"] != "10" {
		t.Errorf("Fail: %s\n%v", exampleYaml, config)
	}
	blogCheck := config.HealthChecks[0]
	if blogCheck.Type != "SimpleHTTPCheck" ||
		!reflect.DeepEqual(blogCheck.Args, map[string]string{"url": "http://mattscodecave.com"}) ||
		!reflect.DeepEqual(blogCheck.Sinks[0], map[string][]string{"ConsoleSink": []string{"true"}}) {
			t.Errorf("Fail: %s\n%v", exampleYaml, config)
	}
}
