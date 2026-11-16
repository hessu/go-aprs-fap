// Package aprsis2 has the Go types for the APRS-IS2 protocol, generated from
// aprsis2.proto, which is a copy of src/aprsis2.proto of aprsc
// (https://github.com/hessu/aprsc).
//
// The generated file is committed, so that the package builds without
// protoc. After updating the schema from aprsc, regenerate it with
//
//	go generate ./aprsis2
//
// which needs protoc and protoc-gen-go (go install
// google.golang.org/protobuf/cmd/protoc-gen-go@latest) in PATH.
package aprsis2

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go_opt=Maprsis2.proto=github.com/hessu/go-aprs-fap/aprsis2 aprsis2.proto
