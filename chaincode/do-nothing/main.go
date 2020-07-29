package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//SmartContract is the data structure which represents this contract and on which  various contract lifecycle functions are attached
type SmartContract struct {
}

// Init is called when the smart contract is instantiated
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke routes invocations to the appropriate function in chaincode
// Current supported invocations are:
//	- update, adds a delta to an aggregate variable in the ledger, all variables are assumed to start at 0
//	- get, retrieves the aggregate value of a variable in the ledger
//	- prune, deletes all rows associated with the variable and replaces them with a single row containing the aggregate value
//	- delete, removes all rows associated with the variable
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) pb.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "update" {
		return s.update(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) update(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success([]byte(fmt.Sprintf("Invoke success!")))
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
