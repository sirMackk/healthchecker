package healthchecker

import (
	"reflect"
	"testing"
	"io/ioutil"
)

func TestSimpleJsonConfigRead(t *testing.T) {
	exampleJson, _ := ioutil.ReadFile("fixtures/basicConfig.json")
	config := ConfigFromJson(exampleJson)
	if config.Core["timeout"] != "10" {
		t.FailNow()
	}

	coreMap := map[string]string{"timeout": "10"}
	if !reflect.DeepEqual(config.Core, coreMap) {
		t.FailNow()
	}

	healthcheckMap := map[string]map[string]string{
		"basic": {"timeout": "5", "target": "http://example.com"},
		"advanced": {"timeout": "7", "target": "https://example.com"},
	}
	if !reflect.DeepEqual(config.Healthchecks, healthcheckMap) {
		t.FailNow()
	}
}
