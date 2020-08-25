package healthchecker

import (
	"io/ioutil"
	//"reflect"
	"testing"
)

//func TestSimpleJsonConfigRead(t *testing.T) {
	//exampleJson, _ := ioutil.ReadFile("fixtures/basicConfig.json")
	//config := ConfigFromJson(exampleJson)
	//if config.Core["timeout"] != "10" {
		//t.FailNow()
	//}

	//coreMap := map[string]string{"timeout": "10"}
	//if !reflect.DeepEqual(config.Core, coreMap) {
		//t.FailNow()
	//}

	//healthcheckMap := map[string]map[string]string{
		//"basic":    {"timeout": "5", "target": "http://example.com"},
		//"advanced": {"timeout": "7", "target": "https://example.com"},
	//}
	//if !reflect.DeepEqual(config.HealthChecks, healthcheckMap) {
		//t.FailNow()
	//}
//}

func TestSimpleYamlConfigRead(t *testing.T) {
	exampleYaml, _ := ioutil.ReadFile("fixtures/exampleConfig.yaml")
	config := ConfigFromYaml(exampleYaml)
	if config.Core["HTTPTimeout"] != "10" {
		t.Errorf("Fail: %s\n%v", exampleYaml, config)
	}
}
