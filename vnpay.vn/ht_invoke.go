package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	// "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	// mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	// "github.com/pkg/errors"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	//  fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	// "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
)

type InvokeRequest struct {
	FuncName string
	Args     []string
}

type QueryRequest struct {
	FuncName string
	Name     string
}

var configFile string = "network.yaml"
var adminUser string = "admin"
var OrgName string = "org1"
var org1User string = "admin"
var channelID string = "vnpay-channel"
var chainCodeID string = "mycc"

const workerNum = 20 // number of Client

type ClientWorker struct {
	id              int
	client          *channel.Client
	invokeChannel   chan InvokeRequest
	responseChannel chan string
	eventListener   *event.Client
}

var invokeChannel chan InvokeRequest
var responseChannel chan string

func main() {

	invokeChannel = make(chan InvokeRequest)
	responseChannel = make(chan string)

	// create a new handler

	configProvider := config.FromFile(configFile)

	for i := 0; i < workerNum; i++ {

		sdk, err := fabsdk.New(configProvider)
		if err != nil {
			fmt.Println("failed to create sdk", err)
			return
		}

		defer sdk.Close()

		clientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(adminUser), fabsdk.WithOrg(OrgName))
		client, err := channel.New(clientContext)

		if err != nil {
			fmt.Println(err, "failed to create new client")
		}

		eventListener, err := event.New(clientContext, event.WithBlockEvents())
		if err != nil {
			fmt.Println(err, "failed to create new event client")
			// return
		}

		worker := ClientWorker{
			id:              i,
			client:          client,
			invokeChannel:   invokeChannel,
			responseChannel: responseChannel,
			eventListener:   eventListener,
		}

		if err != nil {
			fmt.Println("failed to create channel client: ", err)
			return
		}

		go worker.start()

		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/invoke", invoke)
	mux.HandleFunc("/query", query)
	// mux.HandleFunc("/oldInvoke", oldInvoke)
	// listen and serve
	http.ListenAndServe(":8090", mux)
}

func (c *ClientWorker) start() {
	for req := range c.invokeChannel {
		fcn := req.FuncName
		args := [][]byte{}

		for _, element := range req.Args {
			args = append(args, []byte(element))
		}

		responseChannel := make(chan string)
		triggerStop := make(chan string)

		go c.exec(fcn, args, responseChannel, triggerStop)
		go timeOutChecker(responseChannel)

		result := <-responseChannel
		fmt.Println("GOT SOMETHING", result)

		if result == "Timeout" {
			triggerStop <- "done job"
			c.responseChannel <- "Timeout"
		} else {
			c.responseChannel <- result
		}

		continue

		// registration, notifier, err := c.eventListener.RegisterChaincodeEvent(chainCodeID, "updateEvent")

		// if err != nil {
		// 	fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to query chaincode: ", err)
		// 	c.responseChannel <- err.Error()
		// 	c.eventListener.Unregister(registration)
		// 	continue
		// }
		// // defer c.eventListener.Unregister(registration)

		// req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}

		// response, err := c.client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
		// if err != nil {
		// 	fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to query chaincode: ", err)
		// 	c.responseChannel <- err.Error()
		// 	c.eventListener.Unregister(registration)
		// 	continue
		// }

		// fmt.Println(string("txid: " + response.TransactionID))
		// // iterate until receive the exact transactionID event

		// var count int
		// func() {
		// 	timeOutChannel := make(chan string)

		// 	go timeOutChecker(timeOutChannel)
		// 	// select {
		// 	// case <-timeOutChannel:
		// 	// 	c.responseChannel <- "Timeout"
		// 	// 	c.eventListener.Unregister(registration)
		// 	// 	return

		// 	// }

		// 	for {
		// 		select {
		// 		case ccEvent := <-notifier:

		// 			if ccEvent.TxID != string(response.TransactionID) {
		// 				count++
		// 				continue
		// 			}

		// 			c.responseChannel <- "OK after receive " + strconv.Itoa(count) + " events, served by clientId: " + strconv.Itoa(c.id)
		// 			c.eventListener.Unregister(registration)

		// 			return
		// 			// case <-time.After(time.Millisecond * 5):
		// 			// 	c.responseChannel <- "Timeout"
		// 			// 	c.eventListener.Unregister(registration)

		// 			// 	return
		// 		}
		// 	}
		// }()

	}
}

func timeOutChecker(c chan string) {
	select {
	case <-time.After(time.Second * 5):
		c <- "Timeout"
	}
}

func (c *ClientWorker) exec(fcn string, args [][]byte, responseChannel chan string, timeout chan string) {

	registration, notifier, err := c.eventListener.RegisterChaincodeEvent(chainCodeID, "updateEvent")

	if err != nil {
		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to query chaincode: ", err)
		responseChannel <- err.Error()
		c.eventListener.Unregister(registration)
		return
	}
	// defer c.eventListener.Unregister(registration)

	req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}

	response, err := c.client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
	if err != nil {
		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to query chaincode: ", err)
		responseChannel <- err.Error()
		c.eventListener.Unregister(registration)
		return
	}

	fmt.Println(string("txid: " + response.TransactionID))
	// iterate until receive the exact transactionID event

	var count int
	func() {

		for {
			select {
			case ccEvent := <-notifier:

				if ccEvent.TxID != string(response.TransactionID) {
					count++
					continue
				}

				fmt.Println("HEREEE")
				responseChannel <- "OK after receive " + strconv.Itoa(count) + " events, served by clientId: " + strconv.Itoa(c.id)
				c.eventListener.Unregister(registration)
				return
			case <-timeout:
				fmt.Println("timeout triggered")
				c.eventListener.Unregister(registration)
				return
			}
		}
	}()

}

func invoke(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var invokeRequest InvokeRequest
	json.Unmarshal(body, &invokeRequest)

	invokeChannel <- invokeRequest
	response := <-responseChannel

	if response == "Timeout" {
		http.Error(w, "tineout", 500)
	} else {
		fmt.Fprint(w, response)

	}
}

func query(w http.ResponseWriter, r *http.Request) {
	configProvider := config.FromFile(configFile)
	sdk, err1 := fabsdk.New(configProvider)
	if err1 != nil {
		fmt.Println("failed to create sdk", err1)
		return
	}
	defer sdk.Close()

	channelProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(OrgName))

	// Create channel client
	client, err := channel.New(channelProvider)
	if err != nil {
		fmt.Println("failed to create channel client: ", err)
		return
	}

	cfgBackend, err := sdk.Config()
	if err != nil {
		fmt.Printf("failed to get config backend: %s\n", err)
	}
	endpointConfig, err := fab.ConfigFromBackend(cfgBackend)
	if err != nil {
		fmt.Printf("failed to get endpointConfig: %s\n", err)
	}
	serverName := "peer0.org1.example.com"
	peerURL := "grpcs://peer0.org1.example.com:7056"
	peer, err := fabpeer.New(endpointConfig, fabpeer.WithURL(peerURL), fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	if err != nil {
		fmt.Printf("failed to create peer: %s\n", err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var queryRequest QueryRequest
	json.Unmarshal(body, &queryRequest)

	fcn := queryRequest.FuncName
	args := [][]byte{[]byte(queryRequest.Name)}

	req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}
	response, err := client.Query(req, channel.WithTargets(peer))
	if err != nil {
		fmt.Println("failed to query chaincode: ", err)
		return
	}

	fmt.Println(string(response.Payload))

	fmt.Fprint(w, string(response.Payload))

	fmt.Println("5. Finish query")

}

/*
	invoke Chaincode without using client pool
*/
// func oldInvoke(w http.ResponseWriter, r *http.Request) {
// 	configProvider := config.FromFile(configFile)

// 	var err error

// 	sdk, err := fabsdk.New(configProvider)
// 	if err != nil {
// 		fmt.Println("failed to create sdk", err)
// 		return
// 	}
// 	defer sdk.Close()

// 	clientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(adminUser), fabsdk.WithOrg(OrgName))

// 	client, err := channel.New(clientContext)
// 	if err != nil {
// 		fmt.Println("failed to create channel client: ", err)
// 		return
// 	}

// 	body, err := ioutil.ReadAll(r.Body)

// 	if err != nil {
// 		fmt.Printf("Error reading body: %v", err)
// 		http.Error(w, "can't read body", http.StatusBadRequest)
// 		return
// 	}

// 	var invokeRequest InvokeRequest
// 	json.Unmarshal(body, &invokeRequest)

// 	fcn := invokeRequest.FuncName

// 	args := [][]byte{}
// 	for _, element := range invokeRequest.Args {
// 		args = append(args, []byte(element))
// 	}

// 	eventListener, err := event.New(clientContext, event.WithBlockEvents())
// 	if err != nil {
// 		fmt.Println(err, "failed to create new event client")
// 		return
// 	}

// 	reg2, notifier2, err2 := eventListener.RegisterChaincodeEvent(chainCodeID, "updateEvent")
// 	if err2 != nil {
// 		return
// 	}
// 	defer eventListener.Unregister(reg2)

// 	fmt.Println("RegisterChaincodeEvent event registered successfully")

// 	req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}

// 	response, err := client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
// 	if err != nil {
// 		fmt.Println("failed to query chaincode: ", err)
// 		return
// 	}

// 	fmt.Println(string("txid: " + response.TransactionID))

// 	for {
// 		select {
// 		case ccEvent := <-notifier2:

// 			if ccEvent.TxID != string(response.TransactionID) {
// 				fmt.Println("notifier2: ", response.TransactionID, "\t", ccEvent.TxID)
// 				continue
// 			}

// 			fmt.Fprint(w, "OK")
// 			return

// 			// fmt.Println("Descriptor: ", ccEvent.Block.Descriptor)
// 		case <-time.After(time.Millisecond * 50):
// 			fmt.Println("did NOT receive CC event for eventId: ")
// 			fmt.Fprint(w, "Timeout")
// 			return
// 		}
// 	}

// 	fmt.Println("DONEE", response.TransactionID)
// }
