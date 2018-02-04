package system

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"sync"
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
	Unique  string   `json:"unique"`
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
	sync.Mutex
	Result map[string]string
}

var jsonConfig JSON

// channel to control execution
var done = make(chan bool)

// Sum to add numbers
func Sum(x int, y int) int {
	return x + y
}

// Set values in cache
func (cache *ReturnStruct) Set(key string, value string) {
	cache.Lock()
	defer cache.unlock()
	cache.Result[key] = value
}

func (cache *ReturnStruct) unlock() {
	cache.Unlock()
}

// Get data from cache
func (cache *ReturnStruct) Get() map[string]string {
	cache.Lock()
	defer cache.unlock()
	return cache.Result
}

// Poll would parse configs and run them every interval.
func Poll(data *ReturnStruct, jsonConfig JSON, forever bool) {
	duration := jsonConfig.Duration
	ExternalCommandConcurrency(jsonConfig, data)
	if forever == true {
		for {
			<-time.After(time.Duration(duration) * time.Second)
			log.Println("Starting check after")
			ExternalCommandConcurrency(jsonConfig, data)
		}
	}
}

// ExternalCommand Run command from config file
func ExternalCommand(jsonConfig JSON, data *ReturnStruct) {
	if data.Result == nil {
		data.Result = make(map[string]string, 1)
	}
	for _, commands := range jsonConfig.Commands {
		name := commands.Name
		command := commands.Command
		options := commands.Options
		unique := commands.Unique
		saveas := name + unique
		cmd := exec.Command(command, options...)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			exitCode := string(err.Error())
			data.Result[saveas] = string(exitCode[len(exitCode)-1])
		} else {
			data.Result[saveas] = "0"
		}

		data.Result[saveas+"Result"] = strings.TrimSpace(out.String())
	}
}

// ExternalCommandConcurrency Run command from config file
func ExternalCommandConcurrency(jsonConfig JSON, data *ReturnStruct) {
	if data.Result == nil {
		data.Result = make(map[string]string, 1)
	}
	for _, commands := range jsonConfig.Commands {
		log.Println("Check command: " + commands.Name + "_" + commands.Type)
		go executeCommand(commands, data)
	}
	// need to block for at least for one process since testing was fauling otherwise
	<-done
}

func executeCommand(commands Commands, data *ReturnStruct) {
	name := commands.Name
	command := commands.Command
	options := commands.Options
	unique := commands.Unique
	saveas := name + unique
	cmd := exec.Command(command, options...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		exitCode := string(err.Error())
		//data.Result[saveas] = string(exitCode[len(exitCode)-1])
		data.Set(saveas, string(exitCode[len(exitCode)-1]))
	} else {
		data.Set(saveas, "0")
		//data.Result[saveas] = "0"
	}
	data.Set(saveas+"Result", strings.TrimSpace(out.String()))
	//data.Result[saveas+"Result"] = strings.TrimSpace(out.String())
	done <- true
}
