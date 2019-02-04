package common

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/firewatcher/system"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

//ServicesEndpoints to expose
type ServicesEndpoints struct {
	sync.RWMutex
	TW          sync.WaitGroup
	ServicesMap map[string]map[int][]string
}

// InitializeMemory for the cache
func (sEndpointMap *ServicesEndpoints) InitializeMemory() {
	sEndpointMap.Lock()
	defer sEndpointMap.unlock()
	if sEndpointMap.ServicesMap == nil {
		log.Println("Initialize memory for the services endpoint")
		sEndpointMap.ServicesMap = make(map[string]map[int][]string, 1)
	}
}

// Set values in SDMap
func (sEndpointMap *ServicesEndpoints) Set(value map[string]map[int][]string) {
	sEndpointMap.Lock()
	defer sEndpointMap.unlock()
	sEndpointMap.ServicesMap = value
}

// Get data from SDMap
func (sEndpointMap *ServicesEndpoints) Get() map[string]map[int][]string {
	sEndpointMap.RLock()
	defer sEndpointMap.RUnlock()
	return sEndpointMap.ServicesMap
}

// GetApp data from SDMap
func (sEndpointMap *ServicesEndpoints) GetApp(appId string) map[int][]string {
	sEndpointMap.RLock()
	defer sEndpointMap.RUnlock()
	return sEndpointMap.ServicesMap[appId]
}

func (sEndpointMap *ServicesEndpoints) unlock() {
	sEndpointMap.Unlock()
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

type LoginResult struct {
	Token string
}
type AppResult struct {
	App *App `json:"app"`
}
type App struct {
	Id        string    `json:"id"`
	Instances int       `json:"instances"`
	Container Container `json:"container"`
	Tasks     []Tasks
}
type Container struct {
	Docker       *Docker        `json:"docker"`
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

// RemoteCalls would handle network calls
type DCOSCalls interface {
	Login(string, []byte, http.Client) (*http.Response, error)
	ServiceDiscovery(string, string, http.Client) (*http.Response, error)
	PrometheusMetricsScrape(string, http.Client) (*http.Response, error)
}

// Config structure is to store data and configs
type Config struct {
	sync.RWMutex
	TW                sync.WaitGroup
	Cache             system.ReturnStruct
	JsonConfig        system.JSON
	ServicesEndpoints ServicesEndpoints
	WorkingDir        *string
	DataDir           string
}

type httpcalls struct {
}

var hrc httpcalls

// ReadFile method to read config file
func (*Config) ReadFile(fileName string) ([]byte, error) {
	//source, err := ioutil.ReadFile(fileName)
	return ioutil.ReadFile(fileName)
}

// SetupStorage Would setup data dir for data storage
func SetupStorage(ConfigObject *Config, serviceType string, dbFile string) {
	err := os.Mkdir(*ConfigObject.WorkingDir+"/data", 0744)
	ConfigObject.DataDir = *ConfigObject.WorkingDir + "/data"
	if os.IsExist(err) {
		fmt.Println("data dir Already exist")
	} else {
		CheckError(err)
	}

	err = os.Mkdir(*ConfigObject.WorkingDir+"/data/"+serviceType, 0744)
	if os.IsExist(err) {
		fmt.Println(serviceType + " dir Already exist")
	} else {
		CheckError(err)
	}
	newFile, err := os.Create(*ConfigObject.WorkingDir + "/data/" + dbFile)
	CheckError(err)
	fmt.Println("DB file created: ", newFile.Name())
}

func AggregateData(ConfigObject *Config, aggregateFile string, serviceType string) {
	serviceTypeUrl := ConfigObject.DataDir + "/" + serviceType
	aggregateResultFile := ConfigObject.DataDir + "/" + aggregateFile
	aggregateResultTmpFile := ConfigObject.DataDir + "/" + aggregateFile + "_tmp"

	fmt.Println(aggregateResultFile)
	// Open file to add specific server
	aggregateResultFileDescriptor, err := os.OpenFile(aggregateResultTmpFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	CheckError(err)
	WBuffer := bufio.NewWriter(aggregateResultFileDescriptor)
	defer aggregateResultFileDescriptor.Close()
	dataDirInfo, err := ioutil.ReadDir(serviceTypeUrl)
	CheckError(err)
	for _, f := range dataDirInfo {
		if f.IsDir() {
			//fmt.Println(f.Name())
			tmpDirinfo, err := ioutil.ReadDir(serviceTypeUrl + "/" + f.Name())
			path := serviceTypeUrl + "/" + f.Name()
			CheckError(err)
			for _, tf := range tmpDirinfo {
				if !tf.IsDir() {
					file := path + "/" + tf.Name()
					fmt.Println(path)
					fmt.Println(file)
					tmpFileDescriptor, err := os.OpenFile(file, os.O_RDWR, 0755)
					CheckError(err)
					defer tmpFileDescriptor.Close()
					scanner := bufio.NewScanner(tmpFileDescriptor)
					for scanner.Scan() {
						fileLine := scanner.Text()
						_, err := WBuffer.WriteString(fileLine + "\n")
						CheckError(err)
						WBuffer.Flush()
					}
					// Delete the resources we created
					err = os.Remove(file)
					CheckError(err)
				}
			}
			// Delete the resources we created
			err = os.Remove(path)
			CheckError(err)
		}
	}
	ConfigObject.Lock()
	input, err := ioutil.ReadFile(aggregateResultTmpFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	os.Remove(aggregateResultFile)
	err = ioutil.WriteFile(aggregateResultFile, input, 0755)
	if err != nil {
		fmt.Println("Error creating", aggregateResultFile)
		fmt.Println(err)
		return
	}
	os.Remove(aggregateResultTmpFile)
	ConfigObject.Unlock()
}

func ProcessData(filePath string, url string, tempDirPath string, appID string) {
	// Create a file in new temp directory
	tempFile, err := ioutil.TempFile(tempDirPath, "TMP_PROCESSED")
	filename := tempFile.Name()
	if err != nil {
		log.Fatal(err)
	}
	// Open file to add specific server
	fd, err := os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0777)
	CheckError(err)
	WBuffer := bufio.NewWriter(fd)

	f, err := os.Open(filePath)
	CheckError(err)
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
			CheckError(err)
		} else {
			_, err := WBuffer.WriteString(fileLine + "\n")
			CheckError(err)
		}
		WBuffer.Flush()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// Delete the resources we created
	err = os.Remove(filePath)
	CheckError(err)
	fmt.Println("Process file completed..")
}

// WriteToFile method to read config file
func WriteToFile(ConfigObject *Config, serviceType string, resultTmpDir string, data []byte, url string, appID string) {
	//Create a temp dir in the system default temp folder
	tempDirPath, err := ioutil.TempDir(ConfigObject.DataDir+"/"+serviceType, "TMP_"+resultTmpDir)
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
	CheckError(err)
	ProcessData(filename, url, tempDirPath, appID)
}

// ScrapeMetrics method to read config file
func ScrapeMetrics(ConfigObject *Config, url string, serviceType string, resultTmpDir string, CO *Config, appID string, remote DCOSCalls) {
	timeout := time.Duration(10 * time.Second)
	skip := false
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := remote.PrometheusMetricsScrape(url, client)
	if resp != nil && resp.StatusCode == 404 {
		fmt.Println("Error: endpoint " + url + " Does not exist...")
		fmt.Println("------------ Fix your Configrations ------------")
		skip = true
	}
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Error: Http request timeout...")
		fmt.Println("------------ Try again in next cycle ------------")
		skip = true
	} else {
		CheckError(err)
	}
	if skip != true {
		body, err := ioutil.ReadAll(resp.Body)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			fmt.Println("Error: Http request timeout...")
			fmt.Println("------------ Try again in next cycle ------------")
		} else {
			CheckError(err)
			// fmt.Println(string(body))
			WriteToFile(ConfigObject, serviceType, resultTmpDir, body, url, appID)

			defer resp.Body.Close()
		}
	}
	defer CO.TW.Done()
}

// getValue would be used to get values from recursive json file.
func getValue(result map[string]interface{}, tag string) float64 {
	deepvalues := strings.Split(tag, ".")
	value := 0.0
	if len(deepvalues) > 1 {
		t := strings.TrimSpace(deepvalues[0])
		newTag := strings.Replace(tag, t+".", "", 1)
		value = getValue(result[t].(map[string]interface{}), newTag)
	}
	trimmedTag := strings.TrimSpace(tag)
	if result[trimmedTag] != nil {
		return result[trimmedTag].(float64)
	}
	return value
}

// WriteToTmpFile method to read config file
func WriteToTmpFile(ConfigObject *Config, serviceType string, resultTmpDir string, data []byte) {
	//Create a temp dir in the system default temp folder
	tempDirPath, err := ioutil.TempDir(ConfigObject.DataDir+"/"+serviceType, "TMP_"+resultTmpDir)
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
	CheckError(err)
}

func UrlScrape(ConfigObject *Config, location int, list []string) {
	defer ConfigObject.TW.Done()
	data := ""
	name := ConfigObject.JsonConfig.ServiceDiscovery.Apps[location].Name
	appId := ConfigObject.JsonConfig.ServiceDiscovery.Apps[location].Id
	appPath := ConfigObject.JsonConfig.ServiceDiscovery.Apps[location].Path
	appUnique := ConfigObject.JsonConfig.ServiceDiscovery.Apps[location].Unique
	//saveas := appId + appUnique
	client := http.Client{}
	for index := range list {

		request, err := http.NewRequest("GET", "http://"+list[index]+""+appPath+"", nil)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("--------------------- " + appUnique + " -----------------------")
		resp, err := client.Do(request)
		if err != nil {
			log.Fatalln(err)
		}

		if resp != nil {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			uniqueTag := strings.Split(appUnique, ",")
			d := ""
			tag := ""
			for _, commandTags := range uniqueTag {
				commandTagTrimmed := strings.TrimSpace(commandTags)
				tagValue := getValue(result, commandTags)
				// Open file to add specific server
				d += name + `{id="` + appId + `",url="` + list[index] + appPath + `",unique="` + commandTagTrimmed + `"` + tag + `} ` + fmt.Sprintf("%f", tagValue) + ``
				d += "\n"
			}
			data += d + "\n"
		}
	}
	WriteToTmpFile(ConfigObject, "dcos_apps", name, []byte(data))
	data = ""
}
