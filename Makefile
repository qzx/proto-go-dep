compile:
	protoc --go_out=. --go_opt=paths=source_relative --go-dep_out=. --go-dep_opt=paths=source_relative example/example.proto
