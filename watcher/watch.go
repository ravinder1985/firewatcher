package watcher

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/firewatcher/system"

	"github.com/firewatcher/common"
)

func performAction(index int, jsonConfig system.JSON, watcher *common.Watcher) {
	appName := *jsonConfig.Watcher.Monitor[index].Name
	var result map[string]interface{}
	if jsonConfig.Watcher.Monitor[index].URL != nil {

		url := *jsonConfig.Watcher.Monitor[index].URL
		result = watcher.HTTPGet(url)
		// aquire the lock before write anything
		watcher.Lock()
		watcher.Response[appName] = result
		watcher.Unlock()
	}

	if jsonConfig.Watcher.Monitor[index].URL != nil &&
		jsonConfig.Watcher.Monitor[index].Key != nil &&
		jsonConfig.Watcher.Monitor[index].Rules != nil {

		key := *jsonConfig.Watcher.Monitor[index].Key
		conditionType := strings.ToUpper(*jsonConfig.Watcher.Monitor[index].Rules.Type)
		condition := strings.ToUpper(*jsonConfig.Watcher.Monitor[index].Rules.Condition)
		conditionsList := strings.Split(condition, "|")
		if conditionType == "BOOL" {
			liveStatus := common.GetBoolValue(result, key)
			//println(strconv.FormatBool(true))
			//fmt.Println("################", liveStatus)
			for _, value := range conditionsList {
				// value == "FALSE" to make sure its configured to perform
				if value == "FALSE" && liveStatus == false {
					if status, ok := watcher.Cache[appName]; ok {
						fmt.Println(status)
						if status.LastState != "false" {
							if jsonConfig.Watcher.Monitor[index].Rules.FALSE != nil &&
								status.Count >= jsonConfig.Watcher.Monitor[index].Rules.FALSE.Times {
								if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.FALSE) {
									// Update the status after the action so we wont keep
									watcher.Cache[appName] = common.Status{"false", 0}
								}
							} else {
								watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
							}
						}
					} else {
						watcher.Cache[appName] = common.Status{"true", 1}
					}
				}
				// value == "TRUE" to make sure its configured to perform
				if value == "TRUE" && liveStatus == true {
					if status, ok := watcher.Cache[appName]; ok {
						fmt.Println(status)
						if status.LastState != "true" {
							if jsonConfig.Watcher.Monitor[index].Rules != nil &&
								jsonConfig.Watcher.Monitor[index].Rules.TRUE != nil &&
								status.Count >= jsonConfig.Watcher.Monitor[index].Rules.TRUE.Times {
								if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.TRUE) {
									watcher.Cache[appName] = common.Status{"true", 0}
								}
							} else {
								watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
							}
						}
					} else {
						watcher.Cache[appName] = common.Status{"false", 1}
					}
				}
			}
		} else if conditionType == "NUMERIC" {
			threshold := *jsonConfig.Watcher.Monitor[index].Rules.Threshold
			liveStatus := int(common.GetValue(result, key))
			//fmt.Println("################", liveStatus)
			for _, value := range conditionsList {
				switch strings.ToUpper(value) {
				case "GTE":
					fmt.Println("in GTE")
					fmt.Println(threshold)
					if liveStatus >= threshold {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "GTE" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.GTE != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.GTE.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.GTE) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"GTE", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"UNKNOWN", 1}
						}
					}
					fmt.Println(*jsonConfig.Watcher.Monitor[index].Rules.GTE)
				case "LTE":
					fmt.Println("in LTE")
					fmt.Println(threshold)
					if liveStatus <= threshold {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "LTE" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.LTE != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.LTE.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.LTE) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"LTE", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"UNKNOWN", 1}
						}
					}
				case "LT":
					fmt.Println(threshold)
					if liveStatus < threshold {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "LT" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.LT != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.LT.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.LT) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"LT", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"UNKNOWN", 1}
						}
					}
				case "GT":
					fmt.Println(threshold)
					if liveStatus > threshold {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "GT" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.GT != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.GT.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.GT) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"GT", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"UNKNOWN", 1}
						}
					}
				case "EQ":
					fmt.Println(threshold)
					if liveStatus == threshold {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "EQ" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.EQ != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.EQ.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.EQ) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"EQ", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"UNKNOWN", 1}
						}
					}
				case "NEQ":
					fmt.Println(threshold)
					if liveStatus != threshold {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "NEQ" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.NEQ != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.NEQ.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.NEQ) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"NEQ", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"UNKNOWN", 1}
						}
					}
				default:
					fmt.Println(value, "operation is not supported")
				}
			}
		} else if conditionType == "STRING" {
			compare := *jsonConfig.Watcher.Monitor[index].Rules.Compare
			liveStatus := common.GetStringValue(result, key)
			fmt.Println("################", liveStatus)
			for _, value := range conditionsList {
				switch strings.ToUpper(value) {
				case "EQ":
					if liveStatus == compare {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "EQ" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.EQ != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.EQ.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.EQ) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"EQ", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.EQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"NEQ", 1}
						}
					}
				case "NEQ":
					if liveStatus != compare {
						if status, ok := watcher.Cache[appName]; ok {
							if status.LastState != "NEQ" {
								if jsonConfig.Watcher.Monitor[index].Rules != nil &&
									jsonConfig.Watcher.Monitor[index].Rules.NEQ != nil &&
									status.Count >= jsonConfig.Watcher.Monitor[index].Rules.NEQ.Times {
									if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.NEQ) {
										// Update the status after the action so we wont keep
										watcher.Cache[appName] = common.Status{"NEQ", 0}
									}
									//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
								} else {
									watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
								}
							}
						} else {
							watcher.Cache[appName] = common.Status{"EQ", 1}
						}
					}
				default:
					fmt.Println(value, "operation is not supported")
				}
			}
		} else {
			fmt.Println(conditionType, "type is not supported")
		}
	}
	if jsonConfig.Watcher.Monitor[index].Command != nil &&
		jsonConfig.Watcher.Monitor[index].Rules != nil {
		command := *jsonConfig.Watcher.Monitor[index].Command
		option := jsonConfig.Watcher.Monitor[index].Options
		returnCode := system.ExecuteCommandForWatcher(command, option)

		//conditionType := strings.ToUpper(*jsonConfig.Watcher.Monitor[index].Rules.Type)
		condition := strings.ToUpper(*jsonConfig.Watcher.Monitor[index].Rules.Condition)
		conditionsList := strings.Split(condition, "|")
		threshold := *jsonConfig.Watcher.Monitor[index].Rules.Threshold
		for _, value := range conditionsList {
			switch strings.ToUpper(value) {
			case "EQ":
				if jsonConfig.Watcher.Monitor[index].Rules.EQ != nil &&
					returnCode == threshold {
					if status, ok := watcher.Cache[appName]; ok {
						fmt.Println(status)
						fmt.Println(returnCode, threshold)
						if status.LastState != "EQ" {
							if status.Count >= jsonConfig.Watcher.Monitor[index].Rules.EQ.Times {
								fmt.Println("###### take action #######")
								if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.EQ) {
									// Update the status after the action so we wont keep
									watcher.Cache[appName] = common.Status{"EQ", 0}
								}
								//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
							} else {
								watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
							}
						}
					} else {
						watcher.Cache[appName] = common.Status{"NEQ", 1}
					}
				}
			case "NEQ":
				if returnCode != threshold {
					if status, ok := watcher.Cache[appName]; ok {
						fmt.Println(status)
						if jsonConfig.Watcher.Monitor[index].Rules != nil &&
							status.LastState != "NEQ" {
							if jsonConfig.Watcher.Monitor[index].Rules.NEQ != nil &&
								status.Count >= jsonConfig.Watcher.Monitor[index].Rules.NEQ.Times {
								if watcher.Action(jsonConfig.Watcher.Monitor[index].Rules.NEQ) {
									// Update the status after the action so we wont keep
									watcher.Cache[appName] = common.Status{"NEQ", 0}
								}
								//watcher.Events <- *jsonConfig.Watcher.Monitor[index].Rules.NEQ
							} else {
								watcher.Cache[appName] = common.Status{status.LastState, status.Count + 1}
							}
						}
					} else {
						watcher.Cache[appName] = common.Status{"EQ", 1}
					}
				}
			default:
				fmt.Println(value, "operation is not supported")
			}
		}
	}

}

//MonitorChanges ...
func MonitorChanges(jsonConfig system.JSON, watcher *common.Watcher) {
	//fmt.Println("Status per applications", watcher.Cache)
	//fmt.Println("Status per applications", watcher.Response)
	for index := range jsonConfig.Watcher.Monitor {
		go performAction(index, jsonConfig, watcher)
	}
}

// Poll would parse configs and run them every interval.
func Poll(config *common.Config, forever bool) {
	jsonConfig := config.JsonConfig
	if jsonConfig.Watcher != nil && jsonConfig.Watcher.Enable {
		duration := jsonConfig.Watcher.Scrape_interval

		// // Start routine to perform actions
		// go func(watcher *common.Watcher) {
		// 	dac := watcher.Events
		// 	for {
		// 		select {
		// 		case data := <-dac:
		// 			println("------------------")
		// 			println(data.Times)
		// 			println(data.Action)
		// 			println(data.URL)
		// 			println(data.Payload)
		// 			println("------------------")
		// 		case <-time.After(5 * time.Second):
		// 			println("timeout after 5 sec")
		// 		}
		// 	}
		// }(&config.Watcher)
		MonitorChanges(jsonConfig, &config.Watcher)
		if forever == true {
			for {
				<-time.After(time.Duration(duration) * time.Second)
				log.Println("------ Starting Watching check after: ", duration, "second ------")
				MonitorChanges(jsonConfig, &config.Watcher)
			}
		}
	} else {
		fmt.Println("Watcher is disabled")
	}
}
