package system

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
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
	Read   uint64
	Write  uint64
	Result map[string]string
}

// channel to control execution
var done = make(chan bool)

// Sum to add numbers
func Sum(x int, y int) int {
	return x + y
}

// Set values in cache
func (cache *ReturnStruct) Set(key string, value string) {
	atomic.AddUint64(&cache.Write, 1)
	cache.Lock()
	defer cache.unlock()
	cache.Result[key] = value
}

func (cache *ReturnStruct) unlock() {
	cache.Unlock()
}

// Get data from cache
func (cache *ReturnStruct) Get() map[string]string {
	atomic.AddUint64(&cache.Read, 1)
	cache.Lock()
	defer cache.unlock()
	return cache.Result
}

// Poll would parse configs and run them every interval.
func Poll(cache *ReturnStruct, jsonConfig JSON, forever bool) {
	duration := jsonConfig.Duration
	ExternalCommandConcurrency(jsonConfig, cache)
	if forever == true {
		for {
			<-time.After(time.Duration(duration) * time.Second)
			log.Println("------ Starting check after: ", duration, "second ------")
			ExternalCommandConcurrency(jsonConfig, cache)
		}
	}
}

// ExternalCommand Run command from config file
func ExternalCommand(jsonConfig JSON, cache *ReturnStruct) {
	if cache.Result == nil {
		cache.Result = make(map[string]string, 1)
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
			cache.Result[saveas] = string(exitCode[len(exitCode)-1])
		} else {
			cache.Result[saveas] = "0"
		}

		cache.Result[saveas+"Result"] = strings.TrimSpace(out.String())
	}
}

// ExternalCommandConcurrency Run command from config file
func ExternalCommandConcurrency(jsonConfig JSON, cache *ReturnStruct) {
	if cache.Result == nil {
		cache.Result = make(map[string]string, 1)
	}
	counter := 0
	for _, commands := range jsonConfig.Commands {
		counter++
		log.Println("Check command: " + commands.Name + "_" + commands.Type)
		go executeCommand(commands, cache, counter)
	}
	log.Println("Running Threads: ", counter)
	// need to block for at least for one process since testing was fauling otherwise
	<-done
}

func executeCommand(commands Commands, cache *ReturnStruct, id int) {
	log.Println("------ Thread id: ", id, " Started ------")
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
		cache.Set(saveas, string(exitCode[len(exitCode)-1]))
	} else {
		cache.Set(saveas, "0")
		//data.Result[saveas] = "0"
	}
	cache.Set(saveas+"Result", strings.TrimSpace(out.String()))
	//data.Result[saveas+"Result"] = strings.TrimSpace(out.String())
	log.Println("------ Thread id: ", id, " Completed ------")
	done <- true
}
