package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Name string
}

var configFile string = "network.yaml"
var adminUser string = "admin"
var OrgName string = "org1"
var org1User string = "admin"
var channelID string = "vnpay-channel"
var chainCodeID string = "mycc3"

func main() {
	// create a new handler
	mux := http.NewServeMux()
	mux.HandleFunc("/invoke", invoke)
	mux.HandleFunc("/query", query)
	// listen and serve
	http.ListenAndServe(":8090", mux)
}

func invoke(w http.ResponseWriter, r *http.Request) {
	configProvider := config.FromFile(configFile)

	var err error

	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		fmt.Println("failed to create sdk", err)
		return
	}
	defer sdk.Close()

	clientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(adminUser), fabsdk.WithOrg(OrgName))

	client, err := channel.New(clientContext)
	if err != nil {
		fmt.Println("failed to create channel client: ", err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var invokeRequest InvokeRequest
	json.Unmarshal(body, &invokeRequest)

	fcn := invokeRequest.FuncName

	args := [][]byte{}
	for _, element := range invokeRequest.Args {
		args = append(args, []byte(element))
	}

	eventListener, err := event.New(clientContext, event.WithBlockEvents())
	if err != nil {
		fmt.Println(err, "failed to create new event client")
		return
	}

	reg2, notifier2, err2 := eventListener.RegisterChaincodeEvent(chainCodeID, "updateEvent")
	if err2 != nil {
		return
	}
	defer eventListener.Unregister(reg2)

	fmt.Println("RegisterChaincodeEvent event registered successfully")

	req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}

	response, err := client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
	if err != nil {
		fmt.Println("failed to query chaincode: ", err)
		return
	}

	fmt.Println(string("txid: " + response.TransactionID))

	select {
	case ccEvent := <-notifier2:
		fmt.Println("notifier2: ", ccEvent.TxID)
		if ccEvent.TxID == string(response.TransactionID) {
			fmt.Fprint(w, "OK")
		}
		// fmt.Println("Descriptor: ", ccEvent.Block.Descriptor)
	case <-time.After(time.Second * 20):
		fmt.Println("did NOT receive CC event for eventId: ")
		fmt.Fprint(w, "Timeout")
		return
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
	peerURL := "grpcs://peer0.org1.example.com:7051"
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

	fcn := "get"
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
