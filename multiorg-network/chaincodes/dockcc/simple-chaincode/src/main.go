package main

import (
	"fmt"

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
	// // Create a new Smart Contract
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
