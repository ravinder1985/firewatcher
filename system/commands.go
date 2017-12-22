package system

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
)

// JSON has json object
type JSON struct {
	Duration int
	Port     string
	Commands []Commands
}

// Commands has commands onject
type Commands struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Options []string `json:"options"`
	Lables  Lables   `json:"labels"`
	Help    string   `json:"help"`
}

// Lables has leveles for commands
type Lables struct {
	Type   string `json:"type"`
	Metric string `json:"metric"`
}

// ReturnStruct would hold the data
type ReturnStruct struct {
	Result map[string]string
}

var jsonConfig JSON

// Sum to add numbers
func Sum(x int, y int) int {
	return x + y
}

// Poll would parse configs and run them every interval.
func Poll(data *ReturnStruct, jsonConfig JSON, forever bool) {
	duration := jsonConfig.Duration
	ExternalCommand(jsonConfig, data)
	if forever == true {
		for {
			<-time.After(time.Duration(duration) * time.Second)
			go ExternalCommand(jsonConfig, data)
		}
	}
}

// ExternalCommand Run command from config file
func ExternalCommand(jsonConfig JSON, data *ReturnStruct) {
	for _, commands := range jsonConfig.Commands {
		name := commands.Name
		command := commands.Command
		options := commands.Options
		if data.Result == nil {
			data.Result = make(map[string]string, 1)
		}
		cmd := exec.Command(command, options...)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			exitCode := string(err.Error())
			data.Result[name] = string(exitCode[len(exitCode)-1])
		} else {
			data.Result[name] = "0"
		}

		data.Result[name+"Result"] = strings.TrimSpace(out.String())
	}
}
