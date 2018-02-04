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
				Name:    "TestPollSuccess",
				Type:    "service",
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
	println(jsonConfig.data.Result)
	if jsonConfig.data.Result["TestPollSuccess"] != "0" && jsonConfig.data.Result["TestPollSuccessResult"] != "success" {
		t.Errorf("should be :%d, instead of :%s", 0, jsonConfig.data.Result["TestPollSuccess"])
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
				Type:    "service",
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

func TestExternalCommandConcurrencySuccess(t *testing.T) {
	//Mock object for testing
	jsonConfigTest = JSON{
		Duration: 10,
		Port:     "8080",
		Commands: []Commands{
			{
				Name:    "test",
				Type:    "service",
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
	ExternalCommandConcurrency(jsonConfigTest, &data)
	if data.Result["test"] != "0" {
		t.Errorf("should be :%d, instead of :%s", 0, data.Result["test"])
	}
}

func TestExternalCommandSuccessWithUnique(t *testing.T) {
	//Mock object for testing
	jsonConfigTest = JSON{
		Duration: 10,
		Port:     "8080",
		Commands: []Commands{
			{
				Name:    "test",
				Type:    "service",
				Command: "echo",
				Unique:  "uniqueValue",
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
	if data.Result["testuniqueValue"] != "0" {
		t.Errorf("should be :%d, instead of :%s", 0, data.Result["testuniqueValue"])
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
				Type:    "service",
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

func TestExternalCommandConcurrencyFailed(t *testing.T) {
	//Mock object for testing
	jsonConfigTest = JSON{
		Duration: 10,
		Port:     "8080",
		Commands: []Commands{
			{
				Name:    "test",
				Type:    "service",
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
	ExternalCommandConcurrency(jsonConfigTest, &data)
	if data.Result["test"] != "7" {
		t.Errorf("should be :%d, instead of :%s", 0, data.Result["test"])
	}
}
