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
	Duration         int
	Port             string
	ServiceDiscovery *ServiceDiscovery
	LocalMetrics     *LocalMetrics
	Commands         []Commands
}

// ServiceDiscovery to get data from DCOS
type ServiceDiscovery struct {
	Enable          bool
	Scrape_interval int
	Type            string
	Login           Login
	Marathon        Marathon
	Apps            []Apps
}

// LocalMetrics to get data from DCOS
type LocalMetrics struct {
	Enable          bool
	Scrape_interval int
	Type            string
	Urls            []Urls
}

// Urls to get data from DCOS
type Urls struct {
	Name string
	Url  string
}

// Login to get data from DCOS
type Login struct {
	Url      string
	Username string
	Password string
}

// Marathon to get data from DCOS
type Marathon struct {
	Url string
}

// App to get data from DCOS
type Apps struct {
	Name   string
	Id     string
	Port   string
	Path   string
	Unique string
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
	sync.RWMutex
	TW     sync.WaitGroup
	Read   uint64
	Write  uint64
	Result map[string]string
}

// Sum to add numbers
func Sum(x int, y int) int {
	return x + y
}

// InitializeMemory for the cache
func (cache *ReturnStruct) InitializeMemory() {
	cache.Lock()
	defer cache.unlock()
	if cache.Result == nil {
		log.Println("Initialize memory for the cache")
		cache.Result = make(map[string]string, 1)
	}
}

// SetThreadWait values in cache
func (cache *ReturnStruct) SetThreadWait(number int) {
	cache.TW.Add(number)
}

// Set values in cache
func (cache *ReturnStruct) Set(key string, value string) {
	cache.Lock()
	defer cache.unlock()
	//time.Sleep(500 * time.Millisecond)
	cache.Result[key] = value
}

func (cache *ReturnStruct) unlock() {
	cache.Unlock()
}

// Get data from cache
func (cache *ReturnStruct) Get() map[string]string {
	atomic.AddUint64(&cache.Read, 1)
	cache.RLock()
	defer cache.RUnlock()
	// Deep copy in order to return different memory location for map read operation.
	var readCache = make(map[string]string)
	for key, value := range cache.Result {
		readCache[key] = value
	}
	return readCache
}

// SafeRead data from cache
func (cache *ReturnStruct) SafeRead(key string) string {
	cache.RLock()
	defer cache.RUnlock()
	return cache.Result[key]
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
	counter := 0
	// Set maximum number of threads to wait
	cache.SetThreadWait(len(jsonConfig.Commands))
	for _, commands := range jsonConfig.Commands {
		counter++
		go executeCommand(commands, cache, counter)
	}
	log.Println("Running Threads: ", counter)
	cache.TW.Wait()
	log.Println("All Running Threads Completed: ", counter)
}

func executeCommand(commands Commands, cache *ReturnStruct, id int) {
	defer cache.TW.Done()
	name := commands.Name
	command := commands.Command
	options := commands.Options
	unique := commands.Unique
	saveas := name + unique
	log.Println("------ Thread id: ", id, " Started: ", options, " ------")
	cmd := exec.Command(command, options...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		exitCode := string(err.Error())
		cache.Set(saveas, string(exitCode[len(exitCode)-1]))
	} else {
		cache.Set(saveas, "0")
	}
	cache.Set(saveas+"Result", strings.TrimSpace(out.String()))
	atomic.AddUint64(&cache.Write, 1)
	log.Println("------ Thread id: ", id, " Completed: ", options, " ------")
}
