package common

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

type dcosCalls struct {
}

var dcos_rc dcosCalls

func (dcosCalls) Login(url string, jsonData []byte, client http.Client) (*http.Response, error) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"Token": "123456789"}`)
	}
	req := httptest.NewRequest("POST", "http://example.com/foo", nil)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	fmt.Println(resp.StatusCode)
	// fmt.Println(resp.Header.Get("Content-Type"))
	// fmt.Println(resp)
	return resp, nil
}
func (dcosCalls) PrometheusMetricsScrape(url string, client http.Client) (*http.Response, error) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `nginx_service{type="monitor",app="nginx",unique="nginx"} 7`)
	}
	req := httptest.NewRequest("POST", "http://example.com/foo", nil)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	fmt.Println(resp.StatusCode)
	// fmt.Println(resp.Header.Get("Content-Type"))
	// fmt.Println(resp)
	return resp, nil
}

func (dcosCalls) ServiceDiscovery(token string, url string, client http.Client) (*http.Response, error) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if token != "1234" {
			w.WriteHeader(401)
		} else if url != "dcos" && url != "old_dcos" {
			w.WriteHeader(404)
		} else if url == "old_dcos" {
			io.WriteString(w, `{"app":{"id":"/dev/test","acceptedResourceRoles":["*"],"backoffFactor":1.15,"backoffSeconds":1,"container":{"type":"DOCKER","docker":{"forcePullImage":true,"image":"","parameters":[{"key":"volume-driver","value":"pxd"},{"key":"volume","value":"name=dev:/opt/docker/logs"}],"portMappings":[{"containerPort":2540,"hostPort":0,"labels":{},"name":"default","protocol":"tcp","servicePort":10353},{"containerPort":9096,"hostPort":0,"labels":{},"name":"prometheus","protocol":"tcp","servicePort":10355}],"privileged":false},"volumes":[]},"cpus":1,"disk":0,"env":{"JAVA_OPTS":"-Xmx3G","APP-NAME":"abc","ENV":"dev","HOST_TCP":"","env":"dev"},"executor":"","fetch":[{"uri":"/etc/docker.tar.gz","extract":true,"executable":false,"cache":false}],"healthChecks":[{"gracePeriodSeconds":300,"intervalSeconds":60,"maxConsecutiveFailures":3,"path":"/metrics","portIndex":1,"protocol":"MESOS_HTTP","timeoutSeconds":20,"delaySeconds":15}],"instances":1,"labels":{"HAPROXY_0_VHOST":"","HAPROXY_GROUP":"devl"},"maxLaunchDelaySeconds":3600,"mem":4096,"gpus":0,"networks":[{"name":"dcos","mode":"container"}],"requirePorts":false,"upgradeStrategy":{"maximumOverCapacity":0,"minimumHealthCapacity":0},"version":"2018-09-20T19:09:12.036Z","versionInfo":{"lastScalingAt":"2018-09-20T19:09:12.036Z","lastConfigChangeAt":"2018-09-20T19:09:12.036Z"},"killSelection":"YOUNGEST_FIRST","unreachableStrategy":{"inactiveAfterSeconds":300,"expungeAfterSeconds":600},"tasksStaged":0,"tasksRunning":1,"tasksHealthy":1,"tasksUnhealthy":0,"deployments":[],"tasks":[{"ipAddresses":[{"ipAddress":"9.0.36.134","protocol":"IPv4"}],"stagedAt":"2018-11-26T21:34:54.650Z","state":"TASK_RUNNING","ports":[7786,7787],"startedAt":"2018-11-26T21:36:16.040Z","version":"2018-09-20T19:09:12.036Z","id":"dev","appId":"/dev/test","slaveId":"5e2605c3-cf49-4479-af14-73cd00e70ea2-S5","host":"1.4.2.8","healthCheckResults":[{"alive":true,"consecutiveFailures":0,"firstSuccess":"2018-11-27T18:30:47.532Z","lastFailure":null,"lastSuccess":"2018-11-27T18:30:47.532Z","lastFailureCause":null,"instanceId":"dev"}]}],"lastTaskFailure":{"appId":"/dev/test","host":"10.45.22.79","message":"Container exited with status 137","state":"TASK_FAILED","taskId":"dev","timestamp":"2018-11-26T21:34:53.050Z","version":"2018-09-20T19:09:12.036Z","slaveId":"5e2605c3-cf49-4479-af14-73cd00e70ea2-S3"}}}`)

		} else {

			io.WriteString(w, `{"app":{"id":"/dev/test","acceptedResourceRoles":["*"],"backoffFactor":1.15,"backoffSeconds":1,"container":{"type":"DOCKER","docker":{"forcePullImage":true,"image":"","parameters":[{"key":"volume-driver","value":"pxd"},{"key":"volume","value":"name=dev:/opt/docker/logs"}],"privileged":false},"volumes":[],"portMappings":[{"containerPort":2540,"hostPort":0,"labels":{},"name":"default","protocol":"tcp","servicePort":10353},{"containerPort":9096,"hostPort":0,"labels":{},"name":"prometheus","protocol":"tcp","servicePort":10355}]},"cpus":1,"disk":0,"env":{"JAVA_OPTS":"-Xmx3G","APP-NAME":"abc","ENV":"dev","HOST_TCP":"","env":"dev"},"executor":"","fetch":[{"uri":"/etc/docker.tar.gz","extract":true,"executable":false,"cache":false}],"healthChecks":[{"gracePeriodSeconds":300,"intervalSeconds":60,"maxConsecutiveFailures":3,"path":"/metrics","portIndex":1,"protocol":"MESOS_HTTP","timeoutSeconds":20,"delaySeconds":15}],"instances":1,"labels":{"HAPROXY_0_VHOST":"","HAPROXY_GROUP":"devl"},"maxLaunchDelaySeconds":3600,"mem":4096,"gpus":0,"networks":[{"name":"dcos","mode":"container"}],"requirePorts":false,"upgradeStrategy":{"maximumOverCapacity":0,"minimumHealthCapacity":0},"version":"2018-09-20T19:09:12.036Z","versionInfo":{"lastScalingAt":"2018-09-20T19:09:12.036Z","lastConfigChangeAt":"2018-09-20T19:09:12.036Z"},"killSelection":"YOUNGEST_FIRST","unreachableStrategy":{"inactiveAfterSeconds":300,"expungeAfterSeconds":600},"tasksStaged":0,"tasksRunning":1,"tasksHealthy":1,"tasksUnhealthy":0,"deployments":[],"tasks":[{"ipAddresses":[{"ipAddress":"9.0.36.134","protocol":"IPv4"}],"stagedAt":"2018-11-26T21:34:54.650Z","state":"TASK_RUNNING","ports":[7786,7787],"startedAt":"2018-11-26T21:36:16.040Z","version":"2018-09-20T19:09:12.036Z","id":"dev","appId":"/dev/test","slaveId":"5e2605c3-cf49-4479-af14-73cd00e70ea2-S5","host":"1.4.2.8","healthCheckResults":[{"alive":true,"consecutiveFailures":0,"firstSuccess":"2018-11-27T18:30:47.532Z","lastFailure":null,"lastSuccess":"2018-11-27T18:30:47.532Z","lastFailureCause":null,"instanceId":"dev"}]}],"lastTaskFailure":{"appId":"/dev/test","host":"10.45.22.79","message":"Container exited with status 137","state":"TASK_FAILED","taskId":"dev","timestamp":"2018-11-26T21:34:53.050Z","version":"2018-09-20T19:09:12.036Z","slaveId":"5e2605c3-cf49-4479-af14-73cd00e70ea2-S3"}}}`)
		}
	}
	req := httptest.NewRequest("POST", "http://example.com/foo", nil)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	fmt.Println(resp.StatusCode)
	// fmt.Println(resp.Header.Get("Content-Type"))
	// fmt.Println(resp)

	return resp, nil
}
