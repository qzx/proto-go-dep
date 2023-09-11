package example

//go:generate protoc --go-dep_out=. --go-dep_opt=paths=source_relative --go_out=. --go_opt=paths=source_relative -I . example.proto
