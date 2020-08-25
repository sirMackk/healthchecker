package healthchecker

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSimpleYamlConfigRead(t *testing.T) {
	exampleYaml, _ := ioutil.ReadFile("fixtures/exampleConfig.yaml")
	config := ConfigFromYaml(exampleYaml)
	if config.Core["HTTPTimeout"] != "10" {
		t.Errorf("Fail: %s\n%v", exampleYaml, config)
	}
	// TODO add finer check
}
