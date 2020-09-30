module example.com/high-throughput

go 1.14

require (
	github.com/go-redis/redis/v8 v8.1.3
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200728190242-9b3ae92d8664
	github.com/hyperledger/fabric-protos-go v0.0.0-20200728190333-526bfc137380
	github.com/pkg/errors v0.9.1
	github.com/snowzach/protosmart v0.0.0-20200822032325-cae29a40844e // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/hyperledger/fabric-chaincode-go => ../../../libs/fabric-chaincode-go

replace github.com/hyperledger/fabric-protos-go => ../../../libs/fabric-protos-go
