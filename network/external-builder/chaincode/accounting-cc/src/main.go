package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	pb "example.com/simple-chaincode/accounting_service"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var (
	// hardcode for test
	accountServiceURL = "172.16.79.8:50052"
	client            pb.AccountingClient
	conn              *grpc.ClientConn
)

// SmartContract example simple Chaincode implementation
type SmartContract struct {
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("hellooo")
	// init code
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {
	// invoke code
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "create-account" {
		return s.createAccount(APIstub, args)
	} else if function == "update" {
		return s.update(APIstub, args)
	} else if function == "get" {
		return s.get(APIstub, args)
	} else if function == "prune" {
		return s.prune(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) createAccount(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Check we have a valid number of args
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1, got " + strconv.Itoa(len(args)))
	}

	// Extract the args
	account := args[0]
	// Retrieve info needed for the update procedure
	txid := APIstub.GetTxID()

	err := createAccount(txid, account)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("createAccount success!")))
}

func (s *SmartContract) update(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	// Check we have a valid number of args
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, expecting 3, got " + strconv.Itoa(len(args)))
	}

	// Extract the args
	account := args[0]
	op := args[2]
	valueFloat, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Provided value was not a number")
	}

	// Make sure a valid operator is provided
	if op != "+" && op != "-" {
		return shim.Error(fmt.Sprintf("Operator %s is unrecognized", op))
	}

	// Retrieve info needed for the update procedure
	txid := APIstub.GetTxID()
	compositeIndexKey := "account~op~value~txID"

	// Create the composite key that will allow us to query for all deltas on a particular variable
	compositeKey, compositeErr := APIstub.CreateCompositeKey(compositeIndexKey, []string{account, op, args[1], txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", account, compositeErr.Error()))
	}
	// Save the composite key index
	compositePutErr := APIstub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", account, compositePutErr.Error()))
	}

	switch op {
	case "+":
		err = deposit(txid, account, valueFloat)
	case "-":
		err = withdraw(txid, account, valueFloat)
	default:
		err = errors.Errorf("Operator %s is unrecognized", op)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("Invoke success!")))
}

/**
 * Retrieves the aggregate value of a variable in the ledger. Gets all delta rows for the variable
 * and computes the final value from all deltas. The args array for the invocation must contain the
 * following argument:
 *	- args[0] -> The name of the variable to get the value of
 *
 * @param APIstub The chaincode shim
 * @param args The arguments array for the get invocation
 *
 * @return A response structure indicating success or failure with a message
 */
func (s *SmartContract) get(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Check we have a valid number of args
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	account := args[0]
	// Get all deltas for the variable
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("account~op~value~txID", []string{account})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", account, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the account %s exists", account))
	}

	// Iterate through result set and compute final value
	var finalVal float64
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		// Get the next row
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		// Split the composite key into its component parts
		_, keyParts, splitKeyErr := APIstub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		// Retrieve the delta value and operation
		operation := keyParts[1]
		valueStr := keyParts[2]

		// Convert the value string and perform the operation
		value, convErr := strconv.ParseFloat(valueStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		switch operation {
		case "+":
			finalVal += value
		case "-":
			finalVal -= value
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", operation))
		}
	}

	return shim.Success([]byte(strconv.FormatFloat(finalVal, 'f', -1, 64)))
}

/**
 * Prunes a variable by deleting all of its delta rows while computing the final value. Once all rows
 * have been processed and deleted, a single new row is added which defines a delta containing the final
 * computed value of the variable. The args array contains the following argument:
 *	- args[0] -> The name of the variable to prune
 *
 * @param APIstub The chaincode shim
 * @param args The args array for the prune invocation
 *
 * @return A response structure indicating success or failure with a message
 */
func (s *SmartContract) prune(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Check we have a valid number of ars
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	// Retrieve the name of the variable to prune
	account := args[0]

	// Get all delta rows for the variable
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("account~op~value~txID", []string{account})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", account, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the account %s exists", account))
	}

	// Iterate through result set computing final value while iterating and deleting each key
	var finalVal float64
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		// Get the next row
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		// Split the key into its composite parts
		_, keyParts, splitKeyErr := APIstub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		// Retrieve the operation and value
		operation := keyParts[1]
		valueStr := keyParts[2]

		// Convert the value to a float
		value, convErr := strconv.ParseFloat(valueStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		// Delete the row from the ledger
		deltaRowDelErr := APIstub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}

		// Add the value of the deleted row to the final aggregate
		switch operation {
		case "+":
			finalVal += value
		case "-":
			finalVal -= value
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", operation))
		}
	}

	// Update the ledger with the final value
	updateResp := s.update(APIstub, []string{account, strconv.FormatFloat(finalVal, 'f', -1, 64), "+"})
	if updateResp.Status == 500 {
		return shim.Error(fmt.Sprintf("Could not update the final value of the variable after pruning: %s", updateResp.Message))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], finalVal, i)))
}

func createAccount(txID string, accountID string) error {
	ctx := context.Background()

	accountResponse, err := client.CreateAccount(ctx, &pb.CreateAccountRequest{TxID: txID, AccountID: accountID})
	if err != nil {
		fmt.Println("could not Deposit:", err)
		return err
	}
	if !accountResponse.Status {
		return errors.Errorf("Accounting service error: %s", accountResponse.Message)
	}

	fmt.Println(accountResponse.Message)
	return nil
}

func deposit(txID string, accountID string, value float64) error {
	ctx := context.Background()

	depositResponse, err := client.Deposit(ctx, &pb.DepositRequest{TxID: txID, AccountID: accountID, Value: float32(value)})
	if err != nil {
		fmt.Println("could not Deposit:", err)
		return err
	}

	if !depositResponse.Status {
		return errors.Errorf("Accounting service error: %s", depositResponse.Message)
	}

	fmt.Println(depositResponse.Message)
	return nil
}

func withdraw(txID string, accountID string, value float64) error {
	ctx := context.Background()

	withdrawResponse, err := client.Withdraw(ctx, &pb.WithdrawRequest{TxID: txID, AccountID: accountID, Value: float32(value)})
	if err != nil {
		fmt.Println("could not Deposit:", err)
		return err
	}

	if !withdrawResponse.Status {
		return errors.Errorf("Accounting service error: %s", withdrawResponse.Message)
	}

	fmt.Println(withdrawResponse.Message)
	return nil
}

func checkConnection() {
	if conn.GetState().String() != "CONNECTING" {
		connectToAccountingService()
	}
}

func connectToAccountingService() {
	var err error
	// Set up a connection to the server.
	conn, err = grpc.Dial(accountServiceURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// defer conn.Close()
	client = pb.NewAccountingClient(conn)
	fmt.Println("Connected to grpc server:", accountServiceURL)
}

//NOTE - parameters such as ccid and endpoint information are hard coded here for illustration. This can be passed in in a variety of standard ways
func main() {
	// The ccid is assigned to the chaincode on install (using the “peer lifecycle chaincode install <package>” command) for instance

	// if len(os.Args) < 3 {
	// 	fmt.Println("Please supply:\n- installed chaincodeID  (using the “peer lifecycle chaincode install <package>” command)\n- chaincode address (host:port)")
	// 	return
	// }

	// ccid := os.Args[1]
	// address := os.Args[2]

	// server := &shim.ChaincodeServer{
	// 	CCID:    ccid,
	// 	Address: address,
	// 	CC:      new(SmartContract),
	// 	TLSProps: shim.TLSProperties{
	// 		Disabled: true,
	// 	},
	// }

	go connectToAccountingService()

	// fmt.Println("Start Chaincode server on " + address)
	// err := server.Start()
	// if err != nil {
	// 	fmt.Printf("Error starting Simple chaincode: %s", err)
	// 	return
	// }

	// // Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
