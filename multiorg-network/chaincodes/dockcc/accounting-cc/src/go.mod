module example.com/simple-chaincode

go 1.14

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.1
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200728190242-9b3ae92d8664
	github.com/hyperledger/fabric-protos-go v0.0.0-20200728190333-526bfc137380
	github.com/pkg/errors v0.9.1
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)

replace github.com/go-sql-driver/mysql => ../../../libs/mysql
replace github.com/hyperledger/fabric-chaincode-go => ../../../libs/fabric-chaincode-go
replace github.com/hyperledger/fabric-protos-go => ../../../libs/fabric-protos-go

