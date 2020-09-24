/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Demonstrates how to handle data in an application with a high transaction volume where the transactions
 * all attempt to change the same key-value pair in the ledger. Such an application will have trouble
 * as multiple transactions may read a value at a certain version, which will then be invalid when the first
 * transaction updates the value to a new version, thus rejecting all other transactions until they're
 * re-executed.
 * Rather than relying on serialization of the transactions, which is slow, this application initializes
 * a value and then accepts deltas of that value which are added as rows to the ledger. The actual value
 * is then an aggregate of the initial value combined with all of the deltas. Additionally, a pruning
 * function is provided which aggregates and deletes the deltas to update the initial value. This should
 * be done during a maintenance window or when there is a lowered transaction volume, to avoid the proliferation
 * of millions of rows of data.
 *
 * @author	Alexandre Pauwels for IBM
 * @created	17 Aug 2017
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	// shim "../../../fabric-chaincode-go/shim"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"

	redis "github.com/go-redis/redis/v8"
	// "google.golang.org/protobuf/internal/errors"
)

//SmartContract is the data structure which represents this contract and on which  various contract lifecycle functions are attached
type SmartContract struct {
}

type ShimMessage struct {
	Message string
	Nonces  []string
}

type InternalWorldState struct {
	value map[string]float64
	mux   sync.Mutex
}

func (w *InternalWorldState) Lock() {
	w.mux.Lock()
}

func (w *InternalWorldState) Unlock() {
	w.mux.Unlock()
}

func (w *InternalWorldState) Deposit(key string, value float64) error {
	w.value[key] += value

	return nil
}

func (w *InternalWorldState) Withdraw(key string, value float64) error {
	w.value[key] -= value

	return nil
}

func (w *InternalWorldState) GetAccountBalance(key string) (float64, error) {
	if balance, ok := w.value[key]; ok {
		return balance, nil
	}

	return -1, errors.Errorf("account not exist internally")
}

func (w *InternalWorldState) updateAccountBalance(key string, value float64) error {
	w.value[key] = value

	return nil
}

// Define Status codes for the response
const (
	OK                    = 200
	ERROR                 = 500
	REDIS_PORT            = "6379"
	REDIS_LOCK_SUFFIX     = "-lock"
	REDIS_NONCE_SUFFIX    = "nonce"
	REDIS_API_PUBSUB_CHAN = "api-chaincode-channel"
)

const (
	RDB_TXSTATUS_WAITING = 0
	RDB_TXSTATUS_SUCCESS = 1
	RDB_TXSTATUS_FAILED  = 2
)

var (
	REDIS_HOST = "redis.example.com"
	isUseRedis = true
	rdb        *redis.Client
)

// Init is called when the smart contract is instantiated
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) pb.Response {

	_, args := APIstub.GetFunctionAndParameters()

	if len(args) < 1 {
		return shim.Error("Please supply redis host!")
	}

	REDIS_HOST = args[0]

	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return shim.Success(nil)
}
func (s *SmartContract) changeredis(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Please supply redis host!")
	}

	REDIS_HOST = args[0]

	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

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
	} else if function == "transfer" {
		return s.transfer(APIstub, args)
	} else if function == "get" {
		return s.get(APIstub, args)
	} else if function == "prune" {
		return s.prune(APIstub, args)
	} else if function == "delete" {
		return s.delete(APIstub, args)
	} else if function == "putstandard" {
		return s.putStandard(APIstub, args)
	} else if function == "putstandardwithget" {
		return s.putStandardWithGet(APIstub, args)
	} else if function == "getstandard" {
		return s.getStandard(APIstub, args)
	} else if function == "delstandard" {
		return s.delStandard(APIstub, args)
	} else if function == "changeredis" {
		return s.changeredis(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func rdbTryToLockKey(ctx context.Context, key string) (bool, error) {
	isSuccess, err := rdb.SetNX(ctx, key+REDIS_LOCK_SUFFIX, "1", time.Second*30).Result()
	if err != nil {
		return false, err
	}

	if !isSuccess {
		time.Sleep(time.Second)
		return rdbTryToLockKey(ctx, key)
	} else {
		return true, nil
	}
}

func rdbTryToUnlockKey(ctx context.Context, key string) error {
	err := rdb.Del(ctx, key+REDIS_LOCK_SUFFIX).Err()
	return err
}

func rdbDeposit(account string, balance float64, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	rdb.Set(ctx, account, balance+value, 0)

	return nil
}

func rdbWithdraw(account string, balance float64, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	rdb.Set(ctx, account, balance-value, 0)

	return nil
}

func rdbGetNonce(account string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	nonce, err := rdb.Incr(ctx, account+"-"+REDIS_NONCE_SUFFIX).Result()

	if err != nil {
		return -1
	}
	return nonce
}

type rdbTxByNonce struct {
	CompositeIndexKey string
	Status            int
}

func rdbSaveTxByNonce(account string, nonce int64, compositeIndexKey string) (bool, error) {
	ctx := context.Background()
	key := account + "-" + strconv.FormatInt(nonce, 10)
	data, err := json.Marshal(rdbTxByNonce{CompositeIndexKey: compositeIndexKey, Status: RDB_TXSTATUS_WAITING})

	if err != nil {
		return false, err
	}

	isSuccess, err := rdb.SetNX(ctx, key, data, 0).Result()

	if err != nil {
		return false, err
	}

	if !isSuccess {
		fmt.Println("nonce existed")
		return false, nil
	}

	return true, nil
}

func rdbUpdateTxStatus(key string, status int) (bool, error) {
	ctx := context.Background()
	redisDataB, err := rdb.Get(ctx, key).Result()

	if err != nil {
		if err.Error() != "redis: nil" {
			return false, err
		}
		// could do something here
		// update of an inexistent key
		return false, nil
	}

	var redisData rdbTxByNonce
	json.Unmarshal([]byte(redisDataB), &redisData)
	redisData.Status = status

	data, err := json.Marshal(redisData)

	if err != nil {
		return false, err
	}

	_, err = rdb.Set(ctx, key, data, 0).Result()

	if err != nil {
		return false, err
	}

	return true, nil
}

/*
	rdbFetchAccountBalance: for Redis
*/
func (s *SmartContract) rdbFetchAccountBalance(APIstub shim.ChaincodeStubInterface, account string, needValidate bool, valueFloat float64) (float64, error) {
	getCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	getResult, err := rdb.Get(getCtx, account).Result()
	var balance float64

	if err != nil {
		if err.Error() != "redis: nil" {
			return -1, err
		}

		// account balance is not existed in Redis
		fmt.Println("[rdbFetchAccountBalance] missing account Balance")
		var getResponse pb.Response
		getResponse = s.get(APIstub, []string{account})

		if getResponse.Status >= shim.ERRORTHRESHOLD {
			// new account
			balance = 0
		} else {
			balance, err = strconv.ParseFloat(string(getResponse.Payload), 64)

			if err != nil {
				return -1, errors.Errorf("getResponse.Payload is not float number??? %s", err.Error())
			}

			setCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			setError := rdb.Set(setCtx, account, balance, 0).Err()

			if setError != nil {
				return -1, errors.Errorf("Set Redis account balance error", setError)
			}
		}
	} else {
		balance, err = strconv.ParseFloat(getResult, 64)
		if err != nil {
			return -1, errors.Errorf("balance in Redis is not float number??? %s", err)
		}
	}

	fmt.Println("balance", balance)

	// only need to validate when operation is -
	if needValidate && balance-valueFloat < 0 {
		return -1, errors.Errorf("Invalid account %s's balance: %f", account, balance)
	}

	return balance, nil
}

/**
 * Updates the ledger to include a new delta for a particular variable. If this is the first time
 * this variable is being added to the ledger, then its initial value is assumed to be 0. The arguments
 * to give in the args array are as follows:
 *	- args[0] -> name of the variable
 *	- args[1] -> new delta (float)
 *	- args[2] -> operation (currently supported are addition "+" and subtraction "-")
 *
 * @param APIstub The chaincode shim
 * @param args The arguments array for the update invocation
 * @param helpUpdateInternalWS give something into this slice will stop update Internal world state
 *
 * @return A response structure indicating success or failure with a message
 */
func (s *SmartContract) update(APIstub shim.ChaincodeStubInterface, args []string, helpUpdateInternalWS ...bool) pb.Response {
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

	var balance float64
	// ===== START VALIDATION =====
	if !isUseRedis {
		/*
			USE VARIABLES FOR TEST
		*/
		internalWorldState.Lock()
		defer internalWorldState.Unlock()

		if op == "-" {
			err = s.fetchAccountBalance(APIstub, account, true, valueFloat)
		} else {
			err = s.fetchAccountBalance(APIstub, account, false, -1)
		}

		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		/*
			USE REDIS
			todo:
				(1) Lock account in Redis
				(2) Defer unlock account in Redis
				(3) Get account balance is Redis & Validate account balance on "-" operation
		*/
		ctx := context.Background()
		// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		// defer cancel()

		isSucceed, err := rdbTryToLockKey(ctx, account)
		defer rdbTryToUnlockKey(context.Background(), account)

		if err != nil {
			return shim.Error(err.Error())
		}

		if !isSucceed {
			return shim.Error("Cannot lock key in Redis, somehow")
		}

		if op == "-" {
			balance, err = s.rdbFetchAccountBalance(APIstub, account, true, valueFloat)
		} else {
			balance, err = s.rdbFetchAccountBalance(APIstub, account, false, -1) // passing -1 for nothing
		}

		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// ===== END VALIDATION =====

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

	if len(helpUpdateInternalWS) == 0 {
		if !isUseRedis {
			/*
				USE VARIABLES FOR TEST
			*/
			switch {
			case op == "+":
				internalWorldState.Deposit(account, valueFloat)
			case op == "-":
				internalWorldState.Withdraw(account, valueFloat)
			}
		} else {
			/*
				USE REDIS
				todo:
					(1) Update account balance in Redis
			*/
			switch {
			case op == "+":
				err = rdbDeposit(account, balance, valueFloat)
			case op == "-":
				err = rdbWithdraw(account, balance, valueFloat)
			}

			if err != nil {
				// NAMPKH: should we continue without return an error here?
				return shim.Error("Cannot update account balance in Redis!")
			}
		}
	}

	isSuccess := false
	var nonce int64 = -1

	for !isSuccess {
		nonce = rdbGetNonce(account)
		if nonce < 0 {
			return shim.Error("Fail to get nonce for account: " + account)
		}

		isSuccess, err = rdbSaveTxByNonce(account, nonce, compositeKey)
		if err != nil {
			fmt.Println("Cannot rdbSaveTxByNonce to Redis!", err)
			return shim.Error("Cannot rdbSaveTxByNonce to Redis!")
		}
	}

	key := account + "-" + strconv.FormatInt(nonce, 10)
	// eventErr := APIstub.SetEvent("updateEvent", []byte("event-hello"))
	// if eventErr != nil {
	// 	return shim.Error(fmt.Sprintf("Failed to emit event"))
	// }
	payload, err := json.Marshal(ShimMessage{Message: fmt.Sprintf("Successfully added %s%s to %s, nonce: %d", op, args[1], account, nonce), Nonces: []string{key}})

	if err != nil {
		fmt.Println("Fail to marshal payload", err)
		return shim.Error("Fail to marshal payload")
	}

	return shim.Success(payload)
}

type updateResponse struct {
	Payload []byte
	Err     error
}

/**
 * Updates the ledger to include a new delta for a particular variable. If this is the first time
 * this variable is being added to the ledger, then its initial value is assumed to be 0. The arguments
 * to give in the args array are as follows:
 *	- args[0] -> name of the variable
 *	- args[1] -> new delta (float)
 *	- args[2] -> operation (currently supported are addition "+" and subtraction "-")
 *
 * @param APIstub The chaincode shim
 * @param args The arguments array for the update invocation
 * @param helpUpdateInternalWS give something into this slice will stop update Internal world state
 *
 * @return A response structure indicating success or failure with a message
 */
func (s *SmartContract) updateWithRoutine(APIstub shim.ChaincodeStubInterface, args []string, compositeKey string, respChan chan updateResponse, helpUpdateInternalWS ...bool) {
	var err error

	// Check we have a valid number of args
	if len(args) != 3 {
		respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Incorrect number of arguments, expecting 3, got " + strconv.Itoa(len(args)))}
		return
	}

	// Extract the args
	account := args[0]
	op := args[2]
	valueFloat, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Provided value was not a number")}
		return
	}

	// Make sure a valid operator is provided
	if op != "+" && op != "-" {
		respChan <- updateResponse{Payload: nil, Err: errors.Errorf(fmt.Sprintf("Operator %s is unrecognized", op))}
		return
	}

	var balance float64
	// ===== START VALIDATION =====
	if !isUseRedis {
		/*
			USE VARIABLES FOR TEST
		*/
		internalWorldState.Lock()
		defer internalWorldState.Unlock()

		if op == "-" {
			err = s.fetchAccountBalance(APIstub, account, true, valueFloat)
		} else {
			err = s.fetchAccountBalance(APIstub, account, false, -1)
		}

		if err != nil {
			respChan <- updateResponse{Payload: nil, Err: err}
			return
		}
	} else {
		/*
			USE REDIS
			todo:
				(1) Lock account in Redis
				(2) Defer unlock account in Redis
				(3) Get account balance is Redis & Validate account balance on "-" operation
		*/
		ctx := context.Background()
		// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		// defer cancel()

		isSucceed, err := rdbTryToLockKey(ctx, account)
		defer rdbTryToUnlockKey(context.Background(), account)

		if err != nil {
			respChan <- updateResponse{Payload: nil, Err: err}
			return
		}

		if !isSucceed {
			respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Cannot lock key in Redis, somehow")}
			return
		}

		if op == "-" {
			balance, err = s.rdbFetchAccountBalance(APIstub, account, true, valueFloat)
		} else {
			balance, err = s.rdbFetchAccountBalance(APIstub, account, false, -1) // passing -1 for nothing
		}

		if err != nil {
			respChan <- updateResponse{Payload: nil, Err: err}
			return
		}
	}

	// ===== END VALIDATION =====

	if len(helpUpdateInternalWS) == 0 {
		if !isUseRedis {
			/*
				USE VARIABLES FOR TEST
			*/
			switch {
			case op == "+":
				internalWorldState.Deposit(account, valueFloat)
			case op == "-":
				internalWorldState.Withdraw(account, valueFloat)
			}
		} else {
			/*
				USE REDIS
				todo:
					(1) Update account balance in Redis
			*/
			switch {
			case op == "+":
				err = rdbDeposit(account, balance, valueFloat)
			case op == "-":
				err = rdbWithdraw(account, balance, valueFloat)
			}

			if err != nil {
				// NAMPKH: should we continue without return an error here?
				respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Cannot update account balance in Redis!")}
				return
			}
		}
	}

	isSuccess := false
	var nonce int64 = -1

	for !isSuccess {
		nonce = rdbGetNonce(account)
		if nonce < 0 {
			respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Fail to get nonce for account: " + account)}
			return
		}

		isSuccess, err = rdbSaveTxByNonce(account, nonce, compositeKey)
		if err != nil {
			respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Cannot rdbSaveTxByNonce to Redis!", err)}
			return
		}
	}

	key := account + "-" + strconv.FormatInt(nonce, 10)
	// eventErr := APIstub.SetEvent("updateEvent", []byte("event-hello"))
	// if eventErr != nil {
	// 	return shim.Error(fmt.Sprintf("Failed to emit event"))
	// }
	payload, err := json.Marshal(ShimMessage{Message: fmt.Sprintf("Successfully added %s%s to %s, nonce: %d", op, args[1], account, nonce), Nonces: []string{key}})

	if err != nil {
		respChan <- updateResponse{Payload: nil, Err: errors.Errorf("Fail to marshal payload", err)}
		return
	}

	respChan <- updateResponse{Payload: payload, Err: nil}
	return
}

/*
	fetchAccountBalance: for variable using
*/
func (s *SmartContract) fetchAccountBalance(APIstub shim.ChaincodeStubInterface, account string, needValidate bool, valueFloat float64) error {
	balance, err := internalWorldState.GetAccountBalance(account)

	if err != nil {
		fmt.Println("[fetchAccountBalance] missing internal World state")
		//
		var getResponse pb.Response
		getResponse = s.get(APIstub, []string{account})

		if getResponse.Status >= shim.ERRORTHRESHOLD {
			// new account
			balance = 0
		} else {
			balance, err = strconv.ParseFloat(string(getResponse.Payload), 64)
			internalWorldState.updateAccountBalance(account, balance)

			if err != nil {
				return errors.Errorf("getResponse.Payload is not float number??? %s", err.Error())
			}
		}
	}

	fmt.Println("balance", balance)

	// only need to validate when operation is -
	if needValidate && balance-valueFloat < 0 {
		return errors.Errorf("Invalid account %s's balance: %f", account, balance)
	}

	return nil
}

/**
 * transfer
 * The arguments
 * to give in the args array are as follows:
 *	- args[0] -> source account
 *	- args[1] -> destination account
 *	- args[2] -> value
 *
 * @param APIstub The chaincode shim
 * @param args The arguments array for the update invocation
 *
 * @return A response structure indicating success or failure with a message
 */
func (s *SmartContract) transfer(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check we have a valid number of args
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, expecting 3")
	}

	// Extract the args
	sourceAccount := args[0]
	destinationAccount := args[1]
	value := args[2]
	valueFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return shim.Error("Provided value was not a number")
	}

	if valueFloat < 0 {
		return shim.Error(fmt.Sprintf("Negative value: %s", value))
	}

	// ===== START VALIDATION & TRANSFER =====

	// Retrieve info needed for the update procedure
	txid := APIstub.GetTxID()
	compositeIndexKey := "account~op~value~txID"

	// Create the composite key that will allow us to query for all deltas on a particular variable
	compositeKeySource, compositeErr := APIstub.CreateCompositeKey(compositeIndexKey, []string{sourceAccount, "-", value, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", sourceAccount, compositeErr.Error()))
	}

	// Save the composite key index
	compositePutErr := APIstub.PutState(compositeKeySource, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", sourceAccount, compositePutErr.Error()))
	}

	// Create the composite key that will allow us to query for all deltas on a particular variable
	compositeKeyDest, compositeErr := APIstub.CreateCompositeKey(compositeIndexKey, []string{destinationAccount, "+", value, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", destinationAccount, compositeErr.Error()))
	}

	// Save the composite key index
	compositePutErr = APIstub.PutState(compositeKeyDest, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", destinationAccount, compositePutErr.Error()))
	}

	tempChan := make(chan updateResponse)
	var wg sync.WaitGroup
	wg.Add(2)
	messages := []ShimMessage{}

	func() {
		for msg := range tempChan {
			go func(msg updateResponse) {
				var tempMsg ShimMessage

				json.Unmarshal(msg.Payload, &tempMsg)
				messages = append(messages, tempMsg)

				wg.Done()
			}(msg)
		}
	}()

	go s.updateWithRoutine(APIstub, []string{sourceAccount, value, "-"}, compositeKeySource, tempChan)
	go s.updateWithRoutine(APIstub, []string{destinationAccount, value, "+"}, compositeKeyDest, tempChan)

	wg.Wait()

	// var firstMessage ShimMessage
	// var secondMessage ShimMessage

	// Nampkh: should tuning here
	// firstMsg := <-tempChan
	// secondMsg := <-tempChan

	// if firstMsg.Err != nil {
	// 	if secondMsg.Err != nil {
	// 		return shim.Error(firstMsg.Err.Error())
	// 	}

	// 	json.Unmarshal(secondMsg.Payload, &secondMessage)
	// 	rdbUpdateTxStatus(secondMessage.Nonces[0], RDB_TXSTATUS_FAILED)

	// 	return shim.Error(firstMsg.Err.Error())
	// }

	// if secondMsg.Err != nil {
	// 	json.Unmarshal(firstMsg.Payload, &firstMessage)
	// 	rdbUpdateTxStatus(firstMessage.Nonces[0], RDB_TXSTATUS_FAILED)

	// 	return shim.Error(secondMsg.Err.Error())
	// }
	// json.Unmarshal(firstMsg.Payload, &firstMessage)
	// json.Unmarshal(secondMsg.Payload, &secondMessage)

	// senderUpdate := s.update(APIstub, []string{sourceAccount, value, "-"})
	// var senderMessage ShimMessage
	// json.Unmarshal(senderUpdate.Payload, &senderMessage)

	// if senderUpdate.Status > shim.ERRORTHRESHOLD {
	// 	rdbUpdateTxStatus(senderMessage.Nonces[0], RDB_TXSTATUS_FAILED)
	// 	return senderUpdate
	// }

	// receiverUpdate := s.update(APIstub, []string{destinationAccount, value, "+"})
	// var receiveMessage ShimMessage
	// json.Unmarshal(receiverUpdate.Payload, &receiveMessage)

	// if receiverUpdate.Status > shim.ERRORTHRESHOLD {
	// 	if !isUseRedis {
	// 		internalWorldState.Lock()
	// 		defer internalWorldState.Unlock()

	// 		// refund for source account
	// 		internalWorldState.Deposit(sourceAccount, valueFloat)
	// 	} else {
	// 		senderUpdate := s.update(APIstub, []string{sourceAccount, value, "+"})

	// 		// set tx status in redis to failed
	// 		rdbUpdateTxStatus(senderMessage.Nonces[0], RDB_TXSTATUS_FAILED)

	// 		if senderUpdate.Status > shim.ERRORTHRESHOLD {
	// 			// return senderUpdate
	// 		}
	// 	}

	// 	return receiverUpdate
	// }

	// ===== END VALIDATION & TRANSFER =====

	payload, err := json.Marshal(ShimMessage{Message: fmt.Sprintf("Successfully transfer %s from %s to %s", value, sourceAccount, destinationAccount), Nonces: []string{}})

	if err != nil {
		fmt.Println("Fail to marshal payload", err)
		return shim.Error("Fail to marshal payload")
	}

	return shim.Success(payload)
	// return shim.Success([]byte(fmt.Sprintf("Successfully transfer %s from %s to %s", value, sourceAccount, destinationAccount)))
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
func (s *SmartContract) get(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
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
func (s *SmartContract) prune(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	if updateResp.Status == ERROR {
		return shim.Error(fmt.Sprintf("Could not update the final value of the variable after pruning: %s", updateResp.Message))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], finalVal, i)))
}

/**
 * Deletes all rows associated with an aggregate variable from the ledger. The args array
 * contains the following argument:
 *	- args[0] -> The name of the variable to delete
 *
 * @param APIstub The chaincode shim
 * @param args The arguments array for the delete invocation
 *
 * @return A response structure indicating success or failure with a message
 */
func (s *SmartContract) delete(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check there are a correct number of arguments
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	// Retrieve the variable name
	account := args[0]

	// Delete all delta rows
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("account~op~value~txID", []string{account})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve delta rows for %s: %s", account, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Ensure the variable exists
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the account %s exists", account))
	}

	// Iterate through result set and delete all indices
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(fmt.Sprintf("Could not retrieve next delta row: %s", nextErr.Error()))
		}

		deltaRowDelErr := APIstub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
	}

	return shim.Success([]byte(fmt.Sprintf("Deleted %s, %d rows removed", account, i)))
}

/**
 * Converts a float64 to a byte array
 *
 * @param f The float64 to convert
 *
 * @return The byte array representation
 */
func f2barr(f float64) []byte {
	str := strconv.FormatFloat(f, 'f', -1, 64)

	return []byte(str)
}

/**
 * All functions below this are for testing traditional editing of a single row
 */
func (s *SmartContract) putStandard(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	name := args[0]
	valStr := args[1]

	// _, getErr := APIstub.GetState(name)
	// if getErr != nil {
	// 	return shim.Error(fmt.Sprintf("Failed to retrieve the state of %s: %s", name, getErr.Error()))
	// }

	putErr := APIstub.PutState(name, []byte(valStr))
	if putErr != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", putErr.Error()))
	}

	return shim.Success(nil)
}

/**
 * All functions below this are for testing traditional editing of a single row
 */
func (s *SmartContract) putStandardWithGet(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	name := args[0]
	valStr := args[1]

	_, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to retrieve the state of %s: %s", name, getErr.Error()))
	}

	putErr := APIstub.PutState(name, []byte(valStr))
	if putErr != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", putErr.Error()))
	}

	return shim.Success(nil)
}

func (s *SmartContract) getStandard(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	name := args[0]

	val, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to get state: %s", getErr.Error()))
	}

	return shim.Success(val)
	// return shim.Success([]byte("val"))
}

func (s *SmartContract) delStandard(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	name := args[0]

	getErr := APIstub.DelState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Failed to delete state: %s", getErr.Error()))
	}

	return shim.Success(nil)
}

var internalWorldState InternalWorldState

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	internalWorldState = InternalWorldState{value: make(map[string]float64)}

	// ctx := context.Background()
	// pubsub := rdb.Subscribe(ctx, REDIS_API_PUBSUB_CHAN)

	// // Wait for confirmation that subscription is created before publishing anything.
	// _, err := pubsub.Receive(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	// // Go channel which receives messages.
	// ch := pubsub.Channel()

	// // Consume messages.
	// go func() {
	// 	for msg := range ch {
	// 		var message ShimMessage
	// 		json.Unmarshal([]byte(msg.Payload), &message)

	// 		// update tx status in redis by nonce
	// 		for _, key := range message.Nonces {
	// 			statusInt, err := strconv.Atoi(message.Message)
	// 			if err != nil {
	// 				fmt.Println(err)
	// 			}

	// 			rdbUpdateTxStatus(key, statusInt)
	// 		}

	// 	}
	// }()

	// // Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

	// excc
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

	// protosmart.OverrideCodec("proto")
	// fmt.Println("Start Chaincode server on " + address)
	// err = server.Start()
	// if err != nil {
	// 	fmt.Printf("Error starting Simple chaincode: %s", err)
	// 	return
	// }

}
