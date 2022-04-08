# firewatcher
 Prometheus Monitoring plugin to alert

# build_linux_centos
 env GOOS=linux GOARCH=amd64 go build -o firewatcher_linux_amd64 github.com/firewatcher

# Test
go test -covermode=count ./...
ok  	github.com/firewatcher	0.008s	coverage: 18.2% of statements
ok  	github.com/firewatcher/system	0.028s	coverage: 85.7% of statements

# package
tar -cvf firewatcher_linux_amd64_1.0.7.tar firewatcher_linux_amd64 config.json

# Options
-config="config.json" or you can give -config="/var/anyname.json"
-storage.path="/something/something/" if -storage.path is missing it would pick up current location


Watcher Configs
Watcher Could monitor URLs and based on the key configured would take actions.
Actions: HTTP|SCRIPT is supported
Types: BOOL|NUMERIC|STRING is supported
Conditions: EQ|NEQ|GTE|LTE|GT|LT|TRUE|FALSE are supported

```

"watcher": {
    "enable": true,
    "scrape_interval": 10,
    "monitor": [
      {
        "name": "SOMEAPP1",
        "url": "someEndpoint"
      },
      {
        "name": "SOMEAPP2",
        "url": "URL to monitor",
        "key": "Key from json",
        "rules": {
          "type" : "BOOL",
          "condition": "true|false",
          "false": {
            "type":"http",
            "times": 1,
            "action": "DELETE",
            "url": "URL",
            "payload": "{\"something\":\"1\"}"
          },
          "true": {
            "type":"http",
            "times": 5,
            "action": "PUT",
            "url": "URL",
            "payload": "{\"something\":\"2\"}"
          }
        }
      },
      {
        "name": "SOMEAPP3",
        "url": "URL",
        "key": "something.something",
        "rules": {
          "type" : "NUMERIC",
          "threshold": 80,
          "condition": "GTE",
          "GTE": {
            "type":"http",
            "times": 5,
            "action": "PUT",
            "url": "URL",
            "payload": "{\"something\":\"1\"}"
          }
        }
      },
      {
        "name": "SOMEAPP4",
        "url": "URL",
        "key": "scmBranch",
        "rules": {
          "type" : "STRING",
          "compare": "master",
          "condition": "NEQ",
          "NEQ": {
            "type":"http",
            "times": 10,
            "action": "PUT",
            "url": "URL",
            "payload": "{\"something\":\"1\"}"
          }
        }
      },
      {
        "name": "SOMEAPP4",
        "command": "sh",
        "options": ["./scripts/someScript.sh"],
        "rules": {
          "type" : "NUMERIC",
          "threshold": 0,
          "condition": "EQ|NEQ",
          "EQ": {
            "type":"script",
            "times": 5,
            "command": "python",
            "options": ["./scripts/cleanUp.py"]
          },
          "NEQ": {
            "type":"script",
            "times": 10,
            "command": "sh",
            "options": ["./scripts/daemon_check.sh"]
          }
        }
      }
    ]
  }

```
