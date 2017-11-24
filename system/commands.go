package system

import (
	"log"
	"os/exec"
	"time"
)

//ConfigJSON object
type ConfigJSON struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Options string `json:"options"`
}

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
}

// Lables has leveles for commands
type Lables struct {
	Type   string `json:"type"`
	Metric string `json:"metric"`
}

// Data would hole the data
type Data struct {
	Result map[string][]byte
}

var jsonConfig JSON

// Poll would parse configs and run them every interval.
func Poll(data *Data, jsonConfig JSON) {
	duration := jsonConfig.Duration
	go ConfigCommand(jsonConfig, data)
	for {
		<-time.After(time.Duration(duration) * time.Second)
		go ConfigCommand(jsonConfig, data)
	}
}

// ConfigCommand Run command from config file
func ConfigCommand(jsonConfig JSON, data *Data) {
	for _, commands := range jsonConfig.Commands {
		name := commands.Name
		command := commands.Command
		options := commands.Options
		if data.Result == nil {
			//fmt.Println("Allocate memory")
			data.Result = make(map[string][]byte, 1)
		}
		if options == nil {
			out, err := exec.Command(command).Output()
			if (err) != nil {
				log.Fatal(err)
			}
			data.Result[name] = out
			//fmt.Printf("%s: %s", name, out)
		} else {
			//fmt.Println(options)
			out, err := exec.Command(command, options...).Output()
			if (err) != nil {
				log.Fatal(err)
			}
			//fmt.Println(strings.TrimSpace(string(out)))
			data.Result[name] = out
			//fmt.Printf("%s: %s", name, out)
		}
	}
}
