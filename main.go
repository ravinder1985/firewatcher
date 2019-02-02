package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/firewatcher/common"
	"github.com/firewatcher/dcos"
	"github.com/firewatcher/system"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type configHandler struct {
	// ConfigObject keep the configs and data in it
	ConfigObject common.Config
}

func (CTX *configHandler) metrics(w http.ResponseWriter, r *http.Request) {
	jsonConfig := CTX.ConfigObject.JsonConfig
	//monitoringResult := ConfigObject.cache.Get()

	//Lock before read
	CTX.ConfigObject.Cache.RLock()
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
		s = "\n# HELP " + name + " " + commands.Help + "\n"
		s = s + "# TYPE " + name + " " + metric
		d += s + "\n"
		uniqueTag := strings.Split(unique, ",")

		commandresult := CTX.ConfigObject.Cache.Result[saveas+"Result"]
		results := strings.Split(commandresult, "\n")
		tag := ""

		//fmt.Println(data.Result)
		// status could be 0, 1 ,2 or 3 in case of monitoring
		serviceStatus := CTX.ConfigObject.Cache.Result[saveas]
		outputPerTag := ""
		skip := false

		if unique != "" {
			//tag = `,unique="` + unique + `"`
			numberOfTags := len(uniqueTag)
			for _, commandTags := range uniqueTag {
				tag = ""
				commandTagTrimmed := strings.TrimSpace(commandTags)
				commandTagUpper := strings.ToUpper(commandTagTrimmed)
				for _, result := range results {
					commandOutput := strings.Split(result, ":")
					if len(commandOutput) == 2 {
						commandOutputUpper0 := strings.ToUpper(commandOutput[0])
						commandOutputTrimmed1 := strings.TrimSpace(commandOutput[1])
						if commandOutputUpper0 != "RESULT" {
							if !strings.Contains(commandOutputUpper0, "RESULT") {
								tagsFromScript := commandOutput[0]
								// if strings.Contains(commandOutputUpper0, commandTagUpper) {
								// 	tagSetByScript = true
								// 	tagValues := strings.Replace(commandOutputUpper0, "TAG_"+commandTagUpper+"_", "", 1)
								// 	tagsFromScript = tagValues
								// 	tag += ","
								// 	tag += tagsFromScript + `="` + commandOutputTrimmed1 + `"`
								// } else if strings.Contains(commandOutputUpper0, "Tag_") {
								//
								// }

								if strings.Contains(commandOutputUpper0, "TAG_") && strings.Contains(commandOutputUpper0, commandTagUpper) {
									tagValues := strings.Replace(commandOutputUpper0, "TAG_"+commandTagUpper+"_", "", 1)
									tagsFromScript = tagValues
									tag += ","
									tag += tagsFromScript + `="` + commandOutputTrimmed1 + `"`
								} else {
									if !strings.Contains(commandOutputUpper0, "TAG_") {
										tag += ","
										tag += tagsFromScript + `="` + commandOutputTrimmed1 + `"`
									}
								}
							}
							if commandOutputUpper0 == commandTagUpper+"_RESULT" {
								outputPerTag = commandOutputTrimmed1
							}
						} else {
							serviceStatus = commandOutputTrimmed1
						}
					}
				}
				if numberOfTags > 1 && outputPerTag == "" {
					skip = true
				}
				if outputPerTag != "" {
					serviceStatus = outputPerTag
					outputPerTag = ""
				}
				if !skip {
					d += name + `{type="` + commands.Lables.Type + `",app="` + commands.Name + `",unique="` + commandTagTrimmed + `"` + tag + `} ` + serviceStatus + ``
					d += "\n"
				}
			}
		} else {
			for _, result := range results {
				commandOutput := strings.Split(result, ":")
				if len(commandOutput) == 2 {
					commandOutputUpper0 := strings.ToUpper(commandOutput[0])
					commandOutputTrimmed1 := strings.TrimSpace(commandOutput[1])
					if commandOutputUpper0 != "RESULT" {
						if !strings.Contains(commandOutputUpper0, "RESULT") {
							tag += ","
							tag += commandOutput[0] + `="` + commandOutputTrimmed1 + `"`
						}
					} else {
						serviceStatus = commandOutputTrimmed1
					}
				}
			}
			d += name + `{type="` + commands.Lables.Type + `",app="` + commands.Name + `"` + tag + `} ` + serviceStatus + ``
			d += "\n"
		}

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
	file, err := ioutil.ReadFile(CTX.ConfigObject.DataDir + "/aggregateLocal.db")
	checkError(err)
	w.Write([]byte(d))
	w.Write(file)
	// Release the lock
	CTX.ConfigObject.Cache.RUnlock()
}

func (CTX *configHandler) home(w http.ResponseWriter, r *http.Request) {
	read := atomic.LoadUint64(&CTX.ConfigObject.Cache.Read)
	fmt.Fprintf(w, "<h1>read operations: %d</h1>", read)
	write := atomic.LoadUint64(&CTX.ConfigObject.Cache.Write)
	fmt.Fprintf(w, "<h1>write operations: %d</h1>", write)
	fmt.Fprintf(w, "<a href=/metrics>View Metrics</a>")
	fmt.Fprintf(w, "<br><a href=/dcosmetrics>View DCOS Metrics</a>")
	fmt.Fprintf(w, "<br><a href=/servicediscovery>Service Discovery</a>")
}

func (CTX *configHandler) dcosmetrics(w http.ResponseWriter, r *http.Request) {
	CTX.ConfigObject.Lock()
	defer CTX.ConfigObject.Unlock()
	// fd, err := os.Open(ConfigObject.dataDir + "/aggregateResult.db")
	// checkError(err)
	// defer fd.Close()
	file, err := ioutil.ReadFile(CTX.ConfigObject.DataDir + "/aggregateResult.db")
	checkError(err)
	// data := make([]byte, 100000)
	// _, err = fd.Read(data)
	// checkError(err)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//fmt.Fprintf(w, d)
	w.Write(file)
}

func (CTX *configHandler) serviceDiscovery(w http.ResponseWriter, r *http.Request) {
	appID := strings.Replace(r.RequestURI, "/servicediscovery", "", 1)
	var js []byte
	var err error
	if appID == "/" {
		js, err = json.Marshal(CTX.ConfigObject.ServicesEndpoints.Get())
	} else {
		js, err = json.Marshal(CTX.ConfigObject.ServicesEndpoints.GetApp(appID))
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func parseConfig(ConfigFile string, config common.ReadConfig) (system.JSON, error) {
	// pwd, _ := os.Getwd()
	// var filename = "/config.json"
	// source, err := ioutil.ReadFile(pwd + filename)
	var JSONdata system.JSON
	source, err := config.ReadFile(ConfigFile)
	if err != nil {
		return JSONdata, err
	}
	json.Unmarshal(source, &JSONdata)
	return JSONdata, nil
}

// // PollLocal would Get ips from DCOS
// func (CTX *configHandler) PollLocal(ConfigObject *common.Config, aggregateFile string, forever bool) {
// 	duration := CTX.ConfigObject.JsonConfig.LocalMetrics.Scrape_interval
// 	serviceType := CTX.ConfigObject.JsonConfig.LocalMetrics.Type
// 	// Setup storage dir
// 	common.SetupStorage(&CTX.ConfigObject, aggregateFile)
// 	for {
// 		if len(CTX.ConfigObject.JsonConfig.LocalMetrics.Urls) <= 0 {
// 			fmt.Println("-------------- Nothing to monitor for Local Url --------------")
// 			fmt.Println("-------------- Stopping the thread --------------")
// 			break
// 		}
// 		for i := 0; i < len(CTX.ConfigObject.JsonConfig.LocalMetrics.Urls); i++ {
// 			appID := CTX.ConfigObject.JsonConfig.LocalMetrics.Urls[i].Name
// 			url := CTX.ConfigObject.JsonConfig.LocalMetrics.Urls[i].Url
// 			resultTmpDir := strings.Replace(appID, "/", "_", -1)
// 			CTX.ConfigObject.TW.Add(1)
// 			go common.ScrapeMetrics(&CTX.ConfigObject, url, serviceType, resultTmpDir, &CTX.ConfigObject, appID)
// 		}
// 		CTX.ConfigObject.TW.Wait()
// 		common.AggregateData(&CTX.ConfigObject, aggregateFile, serviceType)
// 		<-time.After(time.Duration(duration) * time.Second)
// 	}
// }

func main() {
	// ConfigObject keep the configs and data in it
	// var ConfigObject common.Config
	// Default config file name is config.json
	var CTX configHandler
	var err error
	filename := flag.String("config", "config.json", "a string")
	// Set default Working dir to current location
	currentWD, err := os.Getwd()
	checkError(err)
	CTX.ConfigObject.WorkingDir = flag.String("storage.path", currentWD, "a string")
	flag.Parse()

	err = os.Remove(*CTX.ConfigObject.WorkingDir + "/data/aggregateLocal.db")
	if os.IsExist(err) {
		checkError(err)
	}
	err = os.Remove(*CTX.ConfigObject.WorkingDir + "/data/aggregateResult.db")
	if os.IsExist(err) {
		checkError(err)
	}

	CTX.ConfigObject.JsonConfig, err = parseConfig(*filename, &CTX.ConfigObject)
	checkError(err)
	if CTX.ConfigObject.JsonConfig.ServiceDiscovery != nil && CTX.ConfigObject.JsonConfig.ServiceDiscovery.Enable {
		CTX.ConfigObject.ServicesEndpoints.InitializeMemory()
		go dcos.PollDCOS(&CTX.ConfigObject, "aggregateResult.db", true)
	}
	if CTX.ConfigObject.JsonConfig.LocalMetrics != nil && CTX.ConfigObject.JsonConfig.LocalMetrics.Enable {
		go dcos.PollLocal(&CTX.ConfigObject, "aggregateLocal.db", true)
	}
	forever := true
	CTX.ConfigObject.Cache.InitializeMemory()
	dirType := "local"
	if CTX.ConfigObject.JsonConfig.LocalMetrics != nil {
		dirType = CTX.ConfigObject.JsonConfig.LocalMetrics.Type
	}
	common.SetupStorage(&CTX.ConfigObject, dirType, "aggregateLocal.db")
	go system.Poll(&CTX.ConfigObject.Cache, CTX.ConfigObject.JsonConfig, forever)
	http.HandleFunc("/", CTX.home)
	http.HandleFunc("/metrics", CTX.metrics)
	http.HandleFunc("/dcosmetrics", CTX.dcosmetrics)
	http.HandleFunc("/servicediscovery/", CTX.serviceDiscovery)
	http.ListenAndServe(":"+CTX.ConfigObject.JsonConfig.Port, nil)
}
