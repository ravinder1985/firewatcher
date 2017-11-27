package system

import (
	"testing"
)

type ConfigTest struct {
	data       ReturnStruct
	jsonConfig JSON
}

var jsonConfigTest JSON

func TestSum(t *testing.T) {
	total := Sum(5, 5)
	if total != 10 {
		t.Errorf("should be :%d, instead of :%d", 10, total)
	}
}

func TestPollSuccess(t *testing.T) {
	//Mock object for testing
	jsonConfigTest = JSON{
		Duration: 10,
		Port:     "8080",
		Commands: []Commands{
			{
				Name:    "test",
				Type:    "test",
				Command: "echo",
				Options: []string{"success"},
				Lables: Lables{
					Type:   "test",
					Metric: "Untyped",
				},
			},
		},
	}
	var jsonConfig ConfigTest
	Poll(&jsonConfig.data, jsonConfigTest, false)
	if jsonConfig.data.Result["test"] != "0" && jsonConfig.data.Result["testResult"] != "success" {
		t.Errorf("should be :%d, instead of :%s", 0, jsonConfig.data.Result["test"])
	}
}
func TestExternalCommandSuccess(t *testing.T) {
	//Mock object for testing
	jsonConfigTest = JSON{
		Duration: 10,
		Port:     "8080",
		Commands: []Commands{
			{
				Name:    "test",
				Type:    "test",
				Command: "echo",
				Options: []string{"success"},
				Lables: Lables{
					Type:   "test",
					Metric: "Untyped",
				},
			},
		},
	}
	var data ReturnStruct
	ExternalCommand(jsonConfigTest, &data)
	if data.Result["test"] != "0" {
		t.Errorf("should be :%d, instead of :%s", 0, data.Result["test"])
	}
}

func TestExternalCommandFailed(t *testing.T) {
	//Mock object for testing
	jsonConfigTest = JSON{
		Duration: 10,
		Port:     "8080",
		Commands: []Commands{
			{
				Name:    "test",
				Type:    "test",
				Command: "sh",
				Options: []string{"exit", "7"},
				Lables: Lables{
					Type:   "test",
					Metric: "Untyped",
				},
			},
		},
	}
	var data ReturnStruct
	ExternalCommand(jsonConfigTest, &data)
	if data.Result["test"] != "7" {
		t.Errorf("should be :%d, instead of :%s", 0, data.Result["test"])
	}
}
