# Foreman worker for yggdrasil

This is a POC worker service for yggdrasil that can act as pull client for Foreman.

This is not meant for production use, only as a simple example how to hook into the ygg.

## Hacking

Client is in `foreman_worker`, that includes `yggdrassil_pb.rb` and `yggdrasil_services_pb.rb` to load the gRPC definitions.

For the server side we have the simle `publish_notify.rb` to play around with the server/client MQTT around and get the idea what is going on under the hood.
IRL it would be hooked into the smart_proxy, that will be notifying about the real jobs.


### Preparing gRPC

This is needed only if the ygg protocol will ever change.

Refreshing message and services definitions from proto:

```shell
gem install grpc-tools

grpc_tools_ruby_protoc -I ../yggdrasil/protocol/ --ruby_out=. --grpc_out=. yggdrasil.proto

```
