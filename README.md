# Foreman worker for yggdrasil

This is a worker service for yggdrasil that acts as pull client for Foreman.


## Developement

Refreshing message and services definitions from proto:

```shell
gem install grpc-tools

grpc_tools_ruby_protoc -I ../yggdrasil/protocol/ --ruby_out=. --grpc_out=. yggdrasil.proto

```
