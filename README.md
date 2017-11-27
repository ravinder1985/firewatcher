# firewatcher
 Prometheus Monitoring plugin to alert

# build_linux_centos
 env GOOS=linux GOARCH=amd64 go build -o firewatcher_linux_amd64 github.com/firewatcher

# Test
go test -covermode=count ./...
ok  	github.com/firewatcher	0.008s	coverage: 18.2% of statements
ok  	github.com/firewatcher/system	0.028s	coverage: 85.7% of statements
