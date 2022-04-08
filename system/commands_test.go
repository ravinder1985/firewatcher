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
				Lables: map[string]string{
					"Type":   "test",
					"Metric": "Untyped",
				},
			},
		},
	}
	var jsonConfig ConfigTest
	jsonConfig.data.InitializeMemory()
	Poll(&jsonConfig.data, jsonConfigTest, false)
	result := jsonConfig.data.Get()
	if result["TestPollSuccess"] != "0" && result["TestPollSuccessResult"] != "success" {
		t.Errorf("should be :%d, instead of :%s", 0, result["TestPollSuccess"])
	}
	value := jsonConfig.data.SafeRead("TestPollSuccess")
	if value != "0" && jsonConfig.data.SafeRead("TestPollSuccessResult") != "success" {
		t.Errorf("should be :%d, instead of :%s", 0, value)
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
				Lables: map[string]string{
					"Type":   "test",
					"Metric": "Untyped",
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
				Lables: map[string]string{
					"Type":   "test",
					"Metric": "Untyped",
				},
			},
		},
	}
	var data ReturnStruct
	data.InitializeMemory()
	//data.SetThreadWait(len(jsonConfigTest.Commands))
	ExternalCommandConcurrency(jsonConfigTest, &data)
	result := data.Get()
	if result["test"] != "0" {
		t.Errorf("should be :%d, instead of :%s", 0, result["test"])
	}
	value := data.SafeRead("test")
	if value != "0" {
		t.Errorf("should be :%d, instead of :%s", 0, value)
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
				Lables: map[string]string{
					"Type":   "test",
					"Metric": "Untyped",
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
				Lables: map[string]string{
					"Type":   "test",
					"Metric": "Untyped",
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
				Lables: map[string]string{
					"Type":   "test",
					"Metric": "Untyped",
				},
			},
		},
	}
	var data ReturnStruct
	// if data.Get() == nil {
	// 	log.Println("Set memory for cache.Result map")
	// 	data.Result = make(map[string]string, 1)
	// }
	data.InitializeMemory()
	//data.SetThreadWait(len(jsonConfigTest.Commands))
	ExternalCommandConcurrency(jsonConfigTest, &data)
	result := data.Get()
	if result["test"] != "7" {
		t.Errorf("should be :%d, instead of :%s", 0, result["test"])
	}
	value := data.SafeRead("test")
	if value != "7" {
		t.Errorf("should be :%d, instead of :%s", 0, value)
	}
}
