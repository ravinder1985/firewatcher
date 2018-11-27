package common

import (
	"bufio"
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

// ServiceDiscovery to get data from DCOS
type ServiceDiscovery struct {
	Enable          bool
	Scrape_interval int
	Type            string
	Login           Login
	Marathon        Marathon
	Apps            []Apps
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
	Id   string
	Port string
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

// RemoteCalls would handle network calls
type DCOSCalls interface {
	Login(string, []byte, http.Client) (*http.Response, error)
	ServiceDiscovery(string, string, http.Client) (*http.Response, error)
	PrometheusMetricsScrape(string, http.Client) (*http.Response, error)
}

// Config structure is to store data and configs
type Config struct {
	sync.RWMutex
	TW         sync.WaitGroup
	Cache      system.ReturnStruct
	JsonConfig system.JSON
	WorkingDir *string
	DataDir    string
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
func SetupStorage(ConfigObject *Config, dbFile string) {
	dirName := ConfigObject.JsonConfig.LocalMetrics.Type
	err := os.Mkdir(*ConfigObject.WorkingDir+"/data", 0744)
	ConfigObject.DataDir = *ConfigObject.WorkingDir + "/data"
	if os.IsExist(err) {
		fmt.Println("data dir Already exist")
	} else {
		CheckError(err)
	}

	err = os.Mkdir(*ConfigObject.WorkingDir+"/data/"+dirName, 0744)
	if os.IsExist(err) {
		fmt.Println(dirName + " dir Already exist")
	} else {
		CheckError(err)
	}
	newFile, err := os.Create(*ConfigObject.WorkingDir + "/data/" + dbFile)
	CheckError(err)
	fmt.Println("DB file created: ", newFile.Name())
}

func AggregateData(ConfigObject *Config, aggregateFile string, serviceType string) {
	ConfigObject.Lock()
	serviceTypeUrl := ConfigObject.DataDir + "/" + serviceType
	aggregateResultFile := ConfigObject.DataDir + "/" + aggregateFile
	fmt.Println(aggregateResultFile)
	// Open file to add specific server
	aggregateResultFileDescriptor, err := os.OpenFile(aggregateResultFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
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
			//fmt.Println(resp.StatusCode)
			CheckError(err)
			// fmt.Println(string(body))
			WriteToFile(ConfigObject, serviceType, resultTmpDir, body, url, appID)

			defer resp.Body.Close()
		}
	}
	defer CO.TW.Done()
}
