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
