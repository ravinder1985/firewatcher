package main

import (
	"errors"
	"testing"
)

type testConfig struct {
}

func (testConfig) ReadFile(fileName string) ([]byte, error) {
	if fileName == "" {
		return []byte(`{}`), errors.New("File Not Found")
	}
	return []byte(`{"duration": 12, "port": "8080"}`), nil
}

var ConfigObjectTest testConfig

func TestParseConfigSuccess(t *testing.T) {
	data, err := parseConfig("config.json", testConfig{})
	if data.Duration != 12 && err == nil {
		t.Errorf("should be :%d, instead of :%d", 12, data.Duration)
	}
}
func TestParseConfigError(t *testing.T) {
	data, err := parseConfig("", testConfig{})
	if err.Error() != "File Not Found" && data.Port != "" {
		t.Errorf("should be :%s, instead of :%s", "File Not Found", err.Error())
	}
}
