FROM golang:1.15

RUN apt update && apt install -y protobuf-compiler

RUN go get google.golang.org/protobuf/cmd/protoc-gen-go \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc

RUN export PATH="$PATH:$(go env GOPATH)/bin"

COPY . .

RUN protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        proto/adviser.proto