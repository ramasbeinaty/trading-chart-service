//go:build ignore

// Contains instructions for required generators, invoke by entering
// go generate gen.go

package main

//go:generate protoc --proto_path=proto --go_out=contracts --go_opt=paths=source_relative binance/contracts/models.proto
