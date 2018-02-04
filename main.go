package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/firewatcher/system"
)

//var data system.Data
var data system.ReturnStruct
var jsonConfig system.JSON

// ReadConfig would handle io operation
type ReadConfig interface {
	ReadFile(string) ([]byte, error)
}

// Config structure is to store data and configs
type Config struct {
	data       system.ReturnStruct
	jsonConfig system.JSON
}

// ReadFile method to read config file
func (*Config) ReadFile(fileName string) ([]byte, error) {
	//source, err := ioutil.ReadFile(fileName)
	return ioutil.ReadFile(fileName)
}

// // Set function testing
// func (ds *Config) Set(key string, value string) {
// 	ds.data.Lock()
// 	defer ds.data.Unlock()
// 	ds.data.Result[key] = value
// }
//
// func (ds *Config) unlock() {
// 	ds.data.Unlock()
// 	log.Println("Unlock the cache data")
// }
//
// // Get function testing
// func (ds *Config) Get() map[string]string {
// 	ds.data.Lock()
// 	log.Println("Lock the cache data")
// 	defer ds.unlock()
// 	return ds.data.Result
// }

// ConfigObject keep the configs and data in it
var ConfigObject Config

func metrics(w http.ResponseWriter, r *http.Request) {
	jsonConfig := ConfigObject.jsonConfig
	monitoringResult := ConfigObject.data.Get()
	log.Println("Data in memory: ", monitoringResult)
	var d, s string
	//var serviceStatus string
	for _, commands := range jsonConfig.Commands {
		unique := commands.Unique
		saveas := commands.Name + unique
		metric := "Untyped"
		if commands.Lables.Metric != "" {
			metric = commands.Lables.Metric
		}

		name := commands.Name + "_" + commands.Type
		s = "# HELP " + name + " " + commands.Help + "\n"
		s = s + "# TYPE " + name + " " + metric
		d += s + "\n"
		commandresult := monitoringResult[saveas+"Result"]
		results := strings.Split(commandresult, "\n")
		tag := ""
		if unique != "" {
			tag = `,unique="` + unique + `"`
		}

		//fmt.Println(data.Result)
		// status could be 0, 1 ,2 or 3 in case of monitoring
		serviceStatus := monitoringResult[saveas]
		for _, result := range results {
			commandOutput := strings.Split(result, ":")
			if len(commandOutput) == 2 {
				if strings.ToUpper(commandOutput[0]) != "RESULT" {
					tag += ","
					tag += commandOutput[0] + `="` + strings.TrimSpace(commandOutput[1]) + `"`
				} else {
					serviceStatus = strings.TrimSpace(commandOutput[1])
				}
			}
		}

		d += name + `{type="` + commands.Lables.Type + `",app="` + commands.Name + `"` + tag + `} ` + serviceStatus + ``
		d += "\n"
		//name := commands.Name

		// switch strings.TrimSpace(string(data.Result[commands.Name])) {
		// case "0":
		// 	d += name + `{type="` + commands.Lables.Type + `",status="OK"} ` + string(data.Result[commands.Name]) + ``
		// case "1":
		// 	d += name + `{type="` + commands.Lables.Type + `",status="WARNING"} ` + string(data.Result[commands.Name]) + ``
		// case "2":
		// 	d += name + `{type="` + commands.Lables.Type + `",status="CRITICAL"} ` + string(data.Result[commands.Name]) + ``
		// default:
		// 	d += name + `{type="` + commands.Lables.Type + `",status="UNKNOWN"} ` + string(data.Result[commands.Name]) + ``
		// }
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//fmt.Fprintf(w, d)
	w.Write([]byte(d))
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<a href=/metrics>View Matrics</a>")
}

func parseConfig(ConfigFile string, config ReadConfig) (system.JSON, error) {
	// pwd, _ := os.Getwd()
	// var filename = "/config.json"
	// source, err := ioutil.ReadFile(pwd + filename)
	var data system.JSON
	source, err := config.ReadFile(ConfigFile)
	if err != nil {
		return data, err
	}
	json.Unmarshal(source, &data)
	return data, nil
}

func main() {
	// Default config file name is config.json
	filename := flag.String("config", "config.json", "a string")
	flag.Parse()
	var err error
	ConfigObject.jsonConfig, err = parseConfig(*filename, &ConfigObject)
	if err != nil {
		log.Fatal(err)
	}
	forever := true
	go system.Poll(&ConfigObject.data, ConfigObject.jsonConfig, forever)
	http.HandleFunc("/", home)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(":"+ConfigObject.jsonConfig.Port, nil)
}
