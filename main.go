package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/firewatcher/system"
)

var data system.Data
var jsonConfig system.JSON

func metrics(w http.ResponseWriter, r *http.Request) {
	var d, s string
	for _, commands := range jsonConfig.Commands {
		name := commands.Name + "_" + commands.Type
		s = "# HELP " + name + " check status of the service 0 = OK | 1 = WARNING | 2 = CRITICAL | 3 = UNKNOWN\n"
		s = s + "# TYPE " + name + " Untyped"
		d += s + "\n"
		//name := commands.Name

		switch strings.TrimSpace(string(data.Result[commands.Name])) {
		case "0":
			fmt.Println("OK")
			d += name + `{type="` + commands.Lables.Type + `",status="OK"} ` + string(data.Result[commands.Name]) + ``
		case "1":
			fmt.Println("WARNING")
			d += name + `{type="` + commands.Lables.Type + `",status="WARNING"} ` + string(data.Result[commands.Name]) + ``
		case "2":
			fmt.Println("CRITICAL")
			d += name + `{type="` + commands.Lables.Type + `",status="CRITICAL"} ` + string(data.Result[commands.Name]) + ``
		default:
			fmt.Println("UNKNOWN")
			d += name + `{type="` + commands.Lables.Type + `",status="UNKNOWN"} ` + string(data.Result[commands.Name]) + ``
		}
	}
	fmt.Fprintf(w, d)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<a href=/metrics>View Matrics</a>")
}

func parseConfig() {
	pwd, _ := os.Getwd()
	var filename = "/config.json"
	source, err := ioutil.ReadFile(pwd + filename)
	//fmt.Printf("Value: %s", source)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(source, &jsonConfig)
	//
	// fmt.Println(jsonConfig.Duration)
	// for _, commands := range jsonConfig.Commands {
	// 	fmt.Println(commands.Name)
	// }
}

func main() {
	parseConfig()
	go system.Poll(&data, jsonConfig)
	http.HandleFunc("/", home)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(":"+jsonConfig.Port, nil)
}
