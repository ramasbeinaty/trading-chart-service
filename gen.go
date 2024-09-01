//go:build ignore

// Contains instructions for required generators, invoke by entering
// go generate gen.go

package main

//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/candlestick/contracts/models.proto
//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/candlestick/contracts/service.proto
