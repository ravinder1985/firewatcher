package dcos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/firewatcher/common"
	"github.com/firewatcher/system"
)

type httpcalls struct {
}

var hrc httpcalls
var sdMap = make(map[string]map[int][]string, 1)

// Login to get token
func (httpcalls) Login(url string, jsonData []byte, client http.Client) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)
	return resp, err
}

// PrometheusMetricsScrape to get token
func (httpcalls) PrometheusMetricsScrape(url string, client http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	return resp, err
}

// DCOSLoginTest method returns a login token.
func DCOSLogin(jsonConfig *system.JSON, remote common.DCOSCalls) string {
	skip := false
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	url := jsonConfig.ServiceDiscovery.Login.Url
	username := jsonConfig.ServiceDiscovery.Login.Username
	password := jsonConfig.ServiceDiscovery.Login.Password

	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if password == "" {
		password = os.Getenv("PASSWORD")
	}
	var jsonData = []byte(`{"uid": "` + username + `", "password": "` + password + `"}`)
	resp, err := remote.Login(url, jsonData, client)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Error: DCOSLogin Http request timeout...")
		fmt.Println("------------ Try again in next cycle ------------")
		skip = true
	} else {
		common.CheckError(err)
	}
	if skip != true {
		defer resp.Body.Close()
		var result common.LoginResult
		err = json.NewDecoder(resp.Body).Decode(&result)
		common.CheckError(err)
		return result.Token
	} else {
		return ""
	}
	// var jsonData = []byte(`{"uid": "` + username + `", "password": "` + password + `"}`)
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	// req.Header.Add("Accept", "application/json")
	// req.Header.Add("Content-type", "application/json")
	// resp, err := client.Do(req)
	// if e, ok := err.(net.Error); ok && e.Timeout() {
	// 	fmt.Println("Error: DCOSLogin Http request timeout...")
	// 	fmt.Println("------------ Try again in next cycle ------------")
	// 	skip = true
	// } else {
	// 	common.CheckError(err)
	// }
	// if skip != true {
	// 	defer resp.Body.Close()
	// 	var result common.LoginResult
	// 	err = json.NewDecoder(resp.Body).Decode(&result)
	// 	common.CheckError(err)
	// 	return result.Token
	// } else {
	// 	return ""
	// }
}

// // Login to get token
// func (httpcalls) Login(url string, username string, password string) string {
// 	skip := false
// 	timeout := time.Duration(5 * time.Second)
// 	client := http.Client{
// 		Timeout: timeout,
// 	}
// 	var jsonData = []byte(`{"uid": "` + username + `", "password": "` + password + `"}`)
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	req.Header.Add("Accept", "application/json")
// 	req.Header.Add("Content-type", "application/json")
// 	resp, err := client.Do(req)
// 	if e, ok := err.(net.Error); ok && e.Timeout() {
// 		fmt.Println("Error: DCOSLogin Http request timeout...")
// 		fmt.Println("------------ Try again in next cycle ------------")
// 		skip = true
// 	} else {
// 		common.CheckError(err)
// 	}
// 	if skip != true {
// 		defer resp.Body.Close()
// 		var result common.LoginResult
// 		err = json.NewDecoder(resp.Body).Decode(&result)
// 		common.CheckError(err)
// 		return result.Token
// 	} else {
// 		return ""
// 	}
// }
//
// // DCOSLogin method returns a login token.
// func DCOSLogin(jsonConfig *system.JSON, remote common.DCOSCalls) string {
// 	url := jsonConfig.ServiceDiscovery.Login.Url
// 	username := jsonConfig.ServiceDiscovery.Login.Username
// 	password := jsonConfig.ServiceDiscovery.Login.Password
//
// 	if username == "" {
// 		username = os.Getenv("USERNAME")
// 	}
// 	if password == "" {
// 		password = os.Getenv("PASSWORD")
// 	}
// 	return remote.Login(url, username, password)
// 	// var jsonData = []byte(`{"uid": "` + username + `", "password": "` + password + `"}`)
// 	// req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	// req.Header.Add("Accept", "application/json")
// 	// req.Header.Add("Content-type", "application/json")
// 	// resp, err := client.Do(req)
// 	// if e, ok := err.(net.Error); ok && e.Timeout() {
// 	// 	fmt.Println("Error: DCOSLogin Http request timeout...")
// 	// 	fmt.Println("------------ Try again in next cycle ------------")
// 	// 	skip = true
// 	// } else {
// 	// 	common.CheckError(err)
// 	// }
// 	// if skip != true {
// 	// 	defer resp.Body.Close()
// 	// 	var result common.LoginResult
// 	// 	err = json.NewDecoder(resp.Body).Decode(&result)
// 	// 	common.CheckError(err)
// 	// 	return result.Token
// 	// } else {
// 	// 	return ""
// 	// }
// }

// Login to get token
func (httpcalls) ServiceDiscovery(token string, url string, client http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token="+token)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	return resp, err
}

// GetServiceList method to get list of Ips
func GetServiceList(token string, url string, appPort string, appId string, remote common.DCOSCalls) ([]string, error) {
	var urlList []string

	exposedPort := appPort
	portLocation := -1
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := remote.ServiceDiscovery(token, url, client)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Error: Http request timeout...")
		fmt.Println("------------ Try again in next cycle ------------")
		return urlList, nil
	} else if err != nil {
		return nil, err
	}
	//fmt.Println(resp.StatusCode)
	defer resp.Body.Close()
	var result common.AppResult

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
	var containerPortMapping []common.PortMappings

	if result.App != nil {
		if result.App.Container.Docker != nil && len(result.App.Container.Docker.Portmappings) > 0 {
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
				fmt.Println(containerPortMapping[i].ContainerPort)
				if exposedPort == strconv.Itoa(containerPortMapping[i].ContainerPort) {
					portLocation = i
					break
				}
			}
		}
		var appMap = make(map[int][]string, 1)
		for i := 0; i < len(result.App.Tasks); i++ {
			fmt.Println(result.App.Tasks[i].State)
			if result.App.Tasks[i].State == "TASK_RUNNING" {
				if portLocation > -1 {
					urlList = append(urlList, result.App.Tasks[i].Host+":"+strconv.Itoa(result.App.Tasks[i].Ports[portLocation]))
				} else {
					println("Error: Port [" + exposedPort + "] in configration seems to be wrong! Please make sure port is exposed from the container " + url)
				}
				if len(containerPortMapping) > 0 {
					for j := 0; j < len(containerPortMapping); j++ {
						appMap[containerPortMapping[j].ContainerPort] = append(appMap[containerPortMapping[j].ContainerPort], result.App.Tasks[i].Host+":"+strconv.Itoa(result.App.Tasks[i].Ports[j]))
					}
				}
			}
		}
		sdMap[appId] = appMap
	}

	return urlList, nil
}

// PollDCOS would Get ips from DCOS
func PollDCOS(ConfigObject *common.Config, aggregateFile string, forever bool) {
	duration := ConfigObject.JsonConfig.ServiceDiscovery.Scrape_interval
	token := DCOSLogin(&ConfigObject.JsonConfig, hrc)
	serviceType := ConfigObject.JsonConfig.ServiceDiscovery.Type
	// Setup storage dir
	common.SetupStorage(ConfigObject, serviceType, aggregateFile)
	for {
		if len(ConfigObject.JsonConfig.ServiceDiscovery.Apps) <= 0 {
			fmt.Println("-------------- Nothing to monitor for DCOS --------------")
			fmt.Println("-------------- Stopping the thread --------------")
			break
		}
		for i := 0; i < len(ConfigObject.JsonConfig.ServiceDiscovery.Apps); i++ {
			skip := false
			appID := ConfigObject.JsonConfig.ServiceDiscovery.Apps[i].Id
			url := ConfigObject.JsonConfig.ServiceDiscovery.Marathon.Url + "/" + appID
			appPort := ConfigObject.JsonConfig.ServiceDiscovery.Apps[i].Port
			appPath := ConfigObject.JsonConfig.ServiceDiscovery.Apps[i].Path
			appUnique := ConfigObject.JsonConfig.ServiceDiscovery.Apps[i].Unique
			list, err := GetServiceList(token, url, appPort, appID, hrc)
			if err != nil {
				//fmt.Println(err.Error())
				if err.Error() == "401" {
					fmt.Println("Login Agian........")
					token = DCOSLogin(&ConfigObject.JsonConfig, hrc)
					list, err = GetServiceList(token, url, appPort, appID, hrc)
					common.CheckError(err)
				} else {
					if e, ok := err.(net.Error); ok && e.Timeout() {
						fmt.Println("Error: Http request timeout...")
						fmt.Println("------------ Try again in next cycle ------------")
						skip = true
					} else {
						common.CheckError(err)
					}
					//panic(err)
				}
			} else {
				fmt.Println("No need to login again........")
			}

			if skip != true {
				ConfigObject.TW.Add(1)
				fmt.Println(appUnique)
				if appUnique != "" && appPath != "" {
					go common.UrlScrape(ConfigObject, i, list)
				}
				fmt.Println("List of ips for ", appID, list)
				resultTmpDir := strings.Replace(appID, "/", "_", -1)

				//Set the values in SDMap
				ConfigObject.ServicesEndpoints.Set(sdMap)
				for index := range list {
					//fmt.Println(list[index])
					ConfigObject.TW.Add(1)
					go common.ScrapeMetrics(ConfigObject, "http://"+list[index]+"/metrics", serviceType, resultTmpDir, ConfigObject, appID, hrc)
				}
			}
		}
		ConfigObject.TW.Wait()
		common.AggregateData(ConfigObject, "aggregateDcosApps.db", "dcos_apps")
		common.AggregateData(ConfigObject, aggregateFile, serviceType)
		<-time.After(time.Duration(duration) * time.Second)
	}
}

// PollLocal would Get ips from local endpoints
func PollLocal(ConfigObject *common.Config, aggregateFile string, forever bool) {
	duration := ConfigObject.JsonConfig.LocalMetrics.Scrape_interval
	serviceType := ConfigObject.JsonConfig.LocalMetrics.Type
	// Setup storage dir
	common.SetupStorage(ConfigObject, serviceType, aggregateFile)
	for {
		if len(ConfigObject.JsonConfig.LocalMetrics.Urls) <= 0 {
			fmt.Println("-------------- Nothing to monitor for Local Url --------------")
			fmt.Println("-------------- Stopping the thread --------------")
			break
		}
		for i := 0; i < len(ConfigObject.JsonConfig.LocalMetrics.Urls); i++ {
			appID := ConfigObject.JsonConfig.LocalMetrics.Urls[i].Name
			url := ConfigObject.JsonConfig.LocalMetrics.Urls[i].Url
			resultTmpDir := strings.Replace(appID, "/", "_", -1)
			ConfigObject.TW.Add(1)
			go common.ScrapeMetrics(ConfigObject, url, serviceType, resultTmpDir, ConfigObject, appID, hrc)
		}
		ConfigObject.TW.Wait()
		common.AggregateData(ConfigObject, aggregateFile, serviceType)
		<-time.After(time.Duration(duration) * time.Second)
	}
}
