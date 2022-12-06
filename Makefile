build:
	go fmt ./...

protoc-gen-go:
	protoc --proto_path=${GOPATH}/src:. --go_out=. examples/*.proto

protoc-gen-gapi: protoc-gen-go
	protoc --proto_path=${GOPATH}/src:. --gapi_out=. examples/*.proto