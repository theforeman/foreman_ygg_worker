module github.com/theforeman/ygg_worker

go 1.26

require (
	github.com/google/uuid v1.6.0
	github.com/redhatinsights/yggdrasil v0.4.9
	github.com/redhatinsights/yggdrasil_v0 v0.0.0-20210811162724-41397343c25b
	github.com/subpop/go-log v0.1.2
	google.golang.org/grpc v1.82.2
)

require (
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/redhatinsights/yggdrasil_v0 v0.0.0-20210811162724-41397343c25b => ./yggdrasil
