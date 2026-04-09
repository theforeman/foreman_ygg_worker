module github.com/theforeman/ygg_worker

go 1.25

require (
	github.com/google/uuid v1.6.0
	github.com/redhatinsights/yggdrasil v0.4.9
	github.com/redhatinsights/yggdrasil_v0 v0.0.0-20210811162724-41397343c25b
	github.com/subpop/go-log v0.1.2
	google.golang.org/grpc v1.79.3
)

require (
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/redhatinsights/yggdrasil_v0 v0.0.0-20210811162724-41397343c25b => ./yggdrasil
