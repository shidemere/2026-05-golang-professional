// Package proto contains directive for generating proto files.
package proto

//go:generate protoc --proto_path=../.. --go_out=../../internal/transport/grpc --go-grpc_out=../../internal/transport/grpc api/grpc/EventService.proto
