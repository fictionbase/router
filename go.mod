module github.com/fictionbase/router

go 1.12

require (
	github.com/aws/aws-sdk-go v1.19.27
	github.com/fictionbase/fictionbase v0.0.0-20190510051858-251968169696
	github.com/stretchr/testify v1.3.0
	go.uber.org/zap v1.10.0
	golang.org/x/sys v0.0.0-20190509141414-a5b02f93d862 // indirect
	golang.org/x/text v0.3.2 // indirect
)

replace (
	github.com/fictionbase/agent => ../agent
	github.com/fictionbase/fictionbase => ../fictionbase
	github.com/fictionbase/monitor => ../monitor
	github.com/fictionbase/router => ../router
)
