build:
	go fmt ./...

protoc-gen-go:
	protoc --proto_path=${GOPATH}/src:. --go_out=. examples/*.proto

protoc-gen-gapi: protoc-gen-go
	cd cmd/protoc-gen-gapi/ && go install && cd ../../examples
	protoc --proto_path=${GOPATH}/src:. --gapi_out=. examples/*.proto