package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/firewatcher/system"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type loginResult struct {
	Token string
}
type appResult struct {
	App *App `json:"app"`
}
type App struct {
	Id        string    `json:"id"`
	Instances int       `json:"instances"`
	Container Container `json:"container"`
	Tasks     []Tasks
}
type Container struct {
	Docker       Docker         `json:"docker"`
	Portmappings []PortMappings `json:"portmappings"`
}
type Tasks struct {
	Ipaddresses []Ipaddresses
	Ports       []int
	Id          string
	AppID       string
	Host        string
	State       string
}
type Ipaddresses struct {
	IpAddress string
}
type Docker struct {
	Portmappings []PortMappings `json:"portmappings"`
}
type PortMappings struct {
	ContainerPort int `json:"containerport"`
}

// ReadConfig would handle io operation
type ReadConfig interface {
	ReadFile(string) ([]byte, error)
}

// Config structure is to store data and configs
type Config struct {
	sync.RWMutex
	TW         sync.WaitGroup
	cache      system.ReturnStruct
	jsonConfig system.JSON
	WorkingDir *string
	dataDir    string
}

// ReadFile method to read config file
func (*Config) ReadFile(fileName string) ([]byte, error) {
	//source, err := ioutil.ReadFile(fileName)
	return ioutil.ReadFile(fileName)
}

// SetupStorage Would setup data dir for data storage
func (*Config) SetupStorage(dirName string, dbFile string) {
	err := os.Mkdir(*ConfigObject.WorkingDir+"/data", 0744)
	ConfigObject.dataDir = *ConfigObject.WorkingDir + "/data"
	if os.IsExist(err) {
		fmt.Println("data dir Already exist")
	} else {
		checkError(err)
	}

	err = os.Mkdir(*ConfigObject.WorkingDir+"/data/"+dirName, 0744)
	if os.IsExist(err) {
		fmt.Println(dirName + " dir Already exist")
	} else {
		checkError(err)
	}
	newFile, err := os.Create(*ConfigObject.WorkingDir + "/data/" + dbFile)
	checkError(err)
	fmt.Println("DB file created: ", newFile.Name())
}

// DCOSLogin method returns a login token.
func (*Config) DCOSLogin(jsonConfig system.JSON) string {
	skip := false
	timeout := time.Duration(5 * time.Second)
	dcos := jsonConfig.ServiceDiscovery.Login.Url
	username := jsonConfig.ServiceDiscovery.Login.Username
	password := jsonConfig.ServiceDiscovery.Login.Password
	client := http.Client{
		Timeout: timeout,
	}
	var jsonData = []byte(`{"uid": "` + username + `", "password": "` + password + `"}`)
	req, err := http.NewRequest("POST", dcos, bytes.NewBuffer(jsonData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Error: DCOSLogin Http request timeout...")
		fmt.Println("------------ Try again in next cycle ------------")
		skip = true
	} else {
		checkError(err)
	}
	if skip != true {
		defer resp.Body.Close()
		var result loginResult
		err = json.NewDecoder(resp.Body).Decode(&result)
		checkError(err)
		return result.Token
	} else {
		return ""
	}
}

// GetServiceList method to get list of Ips
func (*Config) GetServiceList(token string, url string, appPort string) ([]string, error) {
	var urlList []string
	exposedPort := appPort
	portLocation := -1
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token="+token)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Error: Http request timeout...")
		fmt.Println("------------ Try again in next cycle ------------")
		return urlList, nil
	} else if err != nil {
		return nil, err
	}
	//fmt.Println(resp.StatusCode)
	defer resp.Body.Close()
	var result appResult

	err = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode == 401 {
		return nil, errors.New("401")
	} else if resp.StatusCode == 404 {
		fmt.Println("Error: Application [" + url + "]does not exsist..")
		fmt.Println("------------ Try again in next cycle ------------")
		return urlList, nil
	} else {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			fmt.Println("Error: Http request timeout...")
			fmt.Println("------------ Try again in next cycle ------------")
			return urlList, nil
		}
	}
	var containerPortMapping []PortMappings
	if len(result.App.Container.Docker.Portmappings) > 0 {
		fmt.Println("OLD version of DCOS")
		containerPortMapping = result.App.Container.Docker.Portmappings
	} else if len(result.App.Container.Portmappings) > 0 {
		fmt.Println("NEW version of DCOS")
		containerPortMapping = result.App.Container.Portmappings
	} else {
		if portLocation == -1 {
			fmt.Println("Error: config file has wrong port OR port is not exposed in the application container OR container is not up yet" + url)
			return urlList, nil
		}
	}
	if len(containerPortMapping) > 0 {
		for i := 0; i < len(containerPortMapping); i++ {
			if exposedPort == strconv.Itoa(containerPortMapping[i].ContainerPort) {
				portLocation = i
				break
			}
		}
	}
	// Get list of runnnig tasks
	// fmt.Println(result.App.Tasks)
	// fmt.Println(portLocation)
	for i := 0; i < len(result.App.Tasks); i++ {
		fmt.Println(result.App.Tasks[i].State)
		if result.App.Tasks[i].State == "TASK_RUNNING" {
			if portLocation > -1 {
				urlList = append(urlList, result.App.Tasks[i].Host+":"+strconv.Itoa(result.App.Tasks[i].Ports[portLocation]))
			} else {
				println("Error: Port [" + exposedPort + "] in configration seems to be wrong! Please make sure port is exposed from the container " + url)
			}
		}
	}
	return urlList, nil
}

func (*Config) aggregateData(aggregateFile string, serviceType string) {
	ConfigObject.Lock()
	serviceTypeUrl := ConfigObject.dataDir + "/" + serviceType
	aggregateResultFile := ConfigObject.dataDir + "/" + aggregateFile
	fmt.Println(aggregateResultFile)
	// Open file to add specific server
	aggregateResultFileDescriptor, err := os.OpenFile(aggregateResultFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	checkError(err)
	WBuffer := bufio.NewWriter(aggregateResultFileDescriptor)
	defer aggregateResultFileDescriptor.Close()
	dataDirInfo, err := ioutil.ReadDir(serviceTypeUrl)
	checkError(err)
	for _, f := range dataDirInfo {
		if f.IsDir() {
			//fmt.Println(f.Name())
			tmpDirinfo, err := ioutil.ReadDir(serviceTypeUrl + "/" + f.Name())
			path := serviceTypeUrl + "/" + f.Name()
			checkError(err)
			for _, tf := range tmpDirinfo {
				if !tf.IsDir() {
					file := path + "/" + tf.Name()
					fmt.Println(path)
					fmt.Println(file)
					tmpFileDescriptor, err := os.OpenFile(file, os.O_RDWR, 0755)
					checkError(err)
					defer tmpFileDescriptor.Close()
					scanner := bufio.NewScanner(tmpFileDescriptor)
					for scanner.Scan() {
						fileLine := scanner.Text()
						_, err := WBuffer.WriteString(fileLine + "\n")
						checkError(err)
						WBuffer.Flush()
					}
					// Delete the resources we created
					err = os.Remove(file)
					checkError(err)
				}
			}
			// Delete the resources we created
			err = os.Remove(path)
			checkError(err)
		}
	}
	ConfigObject.Unlock()
}

func (*Config) processData(filePath string, url string, tempDirPath string, appID string) {
	// Create a file in new temp directory
	tempFile, err := ioutil.TempFile(tempDirPath, "TMP_PROCESSED")
	filename := tempFile.Name()
	if err != nil {
		log.Fatal(err)
	}
	// Open file to add specific server
	fd, err := os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0777)
	checkError(err)
	WBuffer := bufio.NewWriter(fd)

	f, err := os.Open(filePath)
	checkError(err)
	defer f.Close()
	defer fd.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fileLine := scanner.Text()
		var line string
		if !strings.Contains(fileLine, "#") {
			if strings.Contains(fileLine, "{") {
				line = strings.Replace(fileLine, "{", `{endpoint="`+url+`",appid="`+appID+`",`, -1)
			} else {
				line = strings.Replace(fileLine, " ", `{endpoint="`+url+`",appid="`+appID+`"} `, -1)
			}
			_, err := WBuffer.WriteString(line + "\n")
			checkError(err)
		} else {
			_, err := WBuffer.WriteString(fileLine + "\n")
			checkError(err)
		}
		WBuffer.Flush()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// Delete the resources we created
	err = os.Remove(filePath)
	checkError(err)
	fmt.Println("Process file completed..")
}

// WriteToFile method to read config file
func (*Config) WriteToFile(serviceType string, resultTmpDir string, data []byte, url string, appID string) {
	//Create a temp dir in the system default temp folder
	tempDirPath, err := ioutil.TempDir(ConfigObject.dataDir+"/"+serviceType, "TMP_"+resultTmpDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Temp dir created:", tempDirPath)
	//fmt.Println(ConfigObject.dataDir)
	// Create a file in new temp directory
	tempFile, err := ioutil.TempFile(tempDirPath, "TMP_"+resultTmpDir)
	if err != nil {
		log.Fatal(err)
	}
	filename := tempFile.Name()
	fmt.Println("Temp file created:", tempFile.Name())
	// Write to the file
	err = ioutil.WriteFile(filename, data, 0744)
	checkError(err)
	ConfigObject.processData(filename, url, tempDirPath, appID)
}

// ScrapeMetrics method to read config file
func (*Config) ScrapeMetrics(url string, serviceType string, resultTmpDir string, CO *Config, appID string) {
	timeout := time.Duration(10 * time.Second)
	skip := false
	client := http.Client{
		Timeout: timeout,
	}
	//fmt.Println(url)
	//url = "http://cdtsckfk04p:8080/metrics"
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Error: Http request timeout...")
		fmt.Println("------------ Try again in next cycle ------------")
		skip = true
	} else {
		checkError(err)
	}
	if skip != true {
		body, err := ioutil.ReadAll(resp.Body)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			fmt.Println("Error: Http request timeout...")
			fmt.Println("------------ Try again in next cycle ------------")
		} else {
			//fmt.Println(resp.StatusCode)
			checkError(err)
			// fmt.Println(string(body))
			ConfigObject.WriteToFile(serviceType, resultTmpDir, body, url, appID)

			defer resp.Body.Close()
		}
	}
	defer CO.TW.Done()
}

// ConfigObject keep the configs and data in it
var ConfigObject Config

func metrics(w http.ResponseWriter, r *http.Request) {
	jsonConfig := ConfigObject.jsonConfig
	//monitoringResult := ConfigObject.cache.Get()

	//Lock before read
	ConfigObject.cache.RLock()
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
		commandresult := ConfigObject.cache.Result[saveas+"Result"]
		results := strings.Split(commandresult, "\n")
		tag := ""
		if unique != "" {
			tag = `,unique="` + unique + `"`
		}

		//fmt.Println(data.Result)
		// status could be 0, 1 ,2 or 3 in case of monitoring
		serviceStatus := ConfigObject.cache.Result[saveas]
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
	file, err := ioutil.ReadFile(ConfigObject.dataDir + "/aggregateLocal.db")
	checkError(err)
	w.Write([]byte(d))
	w.Write(file)
	// Release the lock
	ConfigObject.cache.RUnlock()
}

func home(w http.ResponseWriter, r *http.Request) {
	read := atomic.LoadUint64(&ConfigObject.cache.Read)
	fmt.Fprintf(w, "<h1>read operations: %d</h1>", read)
	write := atomic.LoadUint64(&ConfigObject.cache.Write)
	fmt.Fprintf(w, "<h1>write operations: %d</h1>", write)
	fmt.Fprintf(w, "<a href=/metrics>View Matrics</a>")
	fmt.Fprintf(w, "<br><a href=/dcosmetrics>View DCOS Matrics</a>")
}

func dcosmetrics(w http.ResponseWriter, r *http.Request) {
	ConfigObject.Lock()
	// fd, err := os.Open(ConfigObject.dataDir + "/aggregateResult.db")
	// checkError(err)
	// defer fd.Close()
	file, err := ioutil.ReadFile(ConfigObject.dataDir + "/aggregateResult.db")
	checkError(err)
	// data := make([]byte, 100000)
	// _, err = fd.Read(data)
	// checkError(err)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//fmt.Fprintf(w, d)
	w.Write(file)
	ConfigObject.Unlock()
}

func parseConfig(ConfigFile string, config ReadConfig) (system.JSON, error) {
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

// PollDCOS would Get ips from DCOS
func PollDCOS(ConfigObject Config, aggregateFile string, forever bool) {
	duration := ConfigObject.jsonConfig.ServiceDiscovery.Scrape_interval
	token := ConfigObject.DCOSLogin(ConfigObject.jsonConfig)
	serviceType := ConfigObject.jsonConfig.ServiceDiscovery.Type
	// Setup storage dir
	ConfigObject.SetupStorage(ConfigObject.jsonConfig.ServiceDiscovery.Type, aggregateFile)
	for {
		if len(ConfigObject.jsonConfig.ServiceDiscovery.Apps) <= 0 {
			fmt.Println("-------------- Nothing to monitor for DCOS --------------")
			fmt.Println("-------------- Stopping the thread --------------")
			break
		}
		for i := 0; i < len(ConfigObject.jsonConfig.ServiceDiscovery.Apps); i++ {
			skip := false
			appID := ConfigObject.jsonConfig.ServiceDiscovery.Apps[i].Id
			url := ConfigObject.jsonConfig.ServiceDiscovery.Marathon.Url + "/" + appID
			appPort := ConfigObject.jsonConfig.ServiceDiscovery.Apps[i].Port
			list, err := ConfigObject.GetServiceList(token, url, appPort)
			if err != nil {
				//fmt.Println(err.Error())
				if err.Error() == "401" {
					fmt.Println("Login Agian........")
					token = ConfigObject.DCOSLogin(ConfigObject.jsonConfig)
					list, err = ConfigObject.GetServiceList(token, url, appPort)
					checkError(err)
				} else {
					if e, ok := err.(net.Error); ok && e.Timeout() {
						fmt.Println("Error: Http request timeout...")
						fmt.Println("------------ Try again in next cycle ------------")
						skip = true
					} else {
						checkError(err)
					}
					//panic(err)
				}
			} else {
				fmt.Println("No need to login again........")
			}
			if skip != true {
				fmt.Println("List of ips for ", appID, list)
				resultTmpDir := strings.Replace(appID, "/", "_", -1)
				for index := range list {
					//fmt.Println(list[index])
					ConfigObject.TW.Add(1)
					go ConfigObject.ScrapeMetrics("http://"+list[index]+"/metrics", serviceType, resultTmpDir, &ConfigObject, appID)
				}
			}
		}
		ConfigObject.TW.Wait()
		ConfigObject.aggregateData(aggregateFile, serviceType)
		<-time.After(time.Duration(duration) * time.Second)
	}
}

// PollLocal would Get ips from DCOS
func PollLocal(ConfigObject Config, aggregateFile string, forever bool) {
	duration := ConfigObject.jsonConfig.LocalMetrics.Scrape_interval
	serviceType := ConfigObject.jsonConfig.LocalMetrics.Type
	// Setup storage dir
	ConfigObject.SetupStorage(ConfigObject.jsonConfig.LocalMetrics.Type, aggregateFile)
	for {
		if len(ConfigObject.jsonConfig.LocalMetrics.Urls) <= 0 {
			fmt.Println("-------------- Nothing to monitor for Local Url --------------")
			fmt.Println("-------------- Stopping the thread --------------")
			break
		}
		for i := 0; i < len(ConfigObject.jsonConfig.LocalMetrics.Urls); i++ {
			appID := ConfigObject.jsonConfig.LocalMetrics.Urls[i].Name
			url := ConfigObject.jsonConfig.LocalMetrics.Urls[i].Url
			resultTmpDir := strings.Replace(appID, "/", "_", -1)
			ConfigObject.TW.Add(1)
			go ConfigObject.ScrapeMetrics(url, serviceType, resultTmpDir, &ConfigObject, appID)
		}
		ConfigObject.TW.Wait()
		ConfigObject.aggregateData(aggregateFile, serviceType)
		<-time.After(time.Duration(duration) * time.Second)
	}
}

func main() {
	// Default config file name is config.json
	var err error
	filename := flag.String("config", "config.json", "a string")
	// Set default Working dir to current location
	currentWD, err := os.Getwd()
	checkError(err)
	ConfigObject.WorkingDir = flag.String("storage.path", currentWD, "a string")
	flag.Parse()

	err = os.Remove(*ConfigObject.WorkingDir + "/data/aggregateLocal.db")
	if os.IsExist(err) {
		checkError(err)
	}
	err = os.Remove(*ConfigObject.WorkingDir + "/data/aggregateResult.db")
	if os.IsExist(err) {
		checkError(err)
	}

	ConfigObject.jsonConfig, err = parseConfig(*filename, &ConfigObject)
	checkError(err)
	if ConfigObject.jsonConfig.ServiceDiscovery != nil && ConfigObject.jsonConfig.ServiceDiscovery.Enable {
		go PollDCOS(ConfigObject, "aggregateResult.db", true)
	}
	if ConfigObject.jsonConfig.LocalMetrics != nil && ConfigObject.jsonConfig.LocalMetrics.Enable {
		go PollLocal(ConfigObject, "aggregateLocal.db", true)
	}
	forever := true
	ConfigObject.cache.InitializeMemory()
	go system.Poll(&ConfigObject.cache, ConfigObject.jsonConfig, forever)
	http.HandleFunc("/", home)
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/dcosmetrics", dcosmetrics)
	http.ListenAndServe(":"+ConfigObject.jsonConfig.Port, nil)
}
