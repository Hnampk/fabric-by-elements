rm accounting_service/accounting_service_grpc.pb.go
rm accounting_service/accounting_service.pb.go

export PATH="$PATH:$(go env GOPATH)/bin"

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative accounting_service/accounting_service.proto