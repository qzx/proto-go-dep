compile:
	protoc --go_opt=paths=source_relative --go_out=./dep/ dep.proto
