package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (s *SimpleChaincode) Init(APIstub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("hellooo")
	// init code
	return shim.Success(nil)
}

func (s *SimpleChaincode) Invoke(APIstub shim.ChaincodeStubInterface) pb.Response {
	// invoke code
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "update" {
		return s.update(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SimpleChaincode) update(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	// trigger event for listener
	eventErr := APIstub.SetEvent("updateEvent", []byte("event-hello"))
	if eventErr != nil {
		return shim.Error(fmt.Sprintf("Failed to emit event"))
	}
	fmt.Println("Someone is calling me")

	return shim.Success([]byte(fmt.Sprintf("Invoke success!")))
}

//NOTE - parameters such as ccid and endpoint information are hard coded here for illustration. This can be passed in in a variety of standard ways
func main() {
	// The ccid is assigned to the chaincode on install (using the “peer lifecycle chaincode install <package>” command) for instance

	if len(os.Args) < 3 {
		fmt.Println("Please supply:\n- installed chaincodeID  (using the “peer lifecycle chaincode install <package>” command)\n- chaincode address (host:port)")
		return
	}

	ccid := os.Args[1]
	address := os.Args[2]

	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: address,
		CC:      new(SimpleChaincode),
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	fmt.Println("Start Chaincode server on " + address)
	err := server.Start()
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
		return
	}
}
