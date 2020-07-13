package main

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
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

func main() {
	fmt.Println("1. Instantiate a fabsdk instance using a configuration.")
	// Create SDK setup for the integration tests
	configFile := "network.yaml"
	configProvider := config.FromFile(configFile)
	sdk, err1 := fabsdk.New(configProvider)
	if err1 != nil {
		fmt.Println("failed to create sdk", err1)
		return
	}
	defer sdk.Close()

	fmt.Println("2. Prepare client context.")
	fmt.Println("2.1 Create a context based on a user and organization, using your fabsdk instance.")

	adminUser := "admin"
	// normalUser := "admin"
	OrgName := "org1"
	channelID := "vnpay-channel"
	clientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(adminUser), fabsdk.WithOrg(OrgName))
	// clientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(normalUser))

	fmt.Println("3. Create a client instance using its New func, passing the context.", OrgName)
	// Create channel client
	client, err := channel.New(clientContext)
	if err != nil {
		fmt.Println("failed to create channel client: ", err)
		return
	}

	//Try to get genesis
	// cfgBackend, err := sdk.Config()
	// if err != nil {
	// 		fmt.Printf("failed to get config backend: %s\n", err)
	// }
	// endpointConfig, err := fab.ConfigFromBackend(cfgBackend)
	// if err != nil {
	// 		fmt.Printf("failed to get endpointConfig: %s\n", err)
	// }
	//
	// serverName := "peer0.org1.example.com"
	// peerURL := "grpcs://peer0.org1.example.com:7051"
	//
	// peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	// if err != nil {
	// 		fmt.Printf("failed to create peer: %s\n", err)
	// 		return
	// }

	fmt.Println("4. Use the funcs provided by each client to create your solution.")
	chainCodeID := "mycc"
	fcn := "update"
	args := [][]byte{[]byte("myvar"), []byte("100"), []byte("+")}

	//
	eventListener, err := event.New(clientContext, event.WithBlockEvents())
	if err != nil {
		fmt.Println(err, "failed to create new event client")
		return
	}

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in hello invoke")

	reg, notifier, err := eventListener.RegisterBlockEvent()
	if err != nil {
		fmt.Println("failed to register block event", err)
		return
	}
	defer eventListener.Unregister(reg)

	fmt.Println("block event registered successfully")

	reg2, notifier2, err2 := eventListener.RegisterChaincodeEvent(chainCodeID, "updateEvent")
	if err2 != nil {
		return
	}
	defer eventListener.Unregister(reg2)
	fmt.Println("RegisterChaincodeEvent event registered successfully")

	req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}
	// response, err := client.Execute(req, channel.WithTargets(peer))
	response, err := client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
	if err != nil {
		fmt.Println("failed to query chaincode: ", err)
		return
	}

	fmt.Println(string(response.TransactionID))

	go func() {
		select {
		case ccEvent := <-notifier2:
			fmt.Println("notifier2: ", ccEvent)
			// fmt.Println("Descriptor: ", ccEvent.Block.Descriptor)
		case <-time.After(time.Second * 20):
			fmt.Println("did NOT receive CC event for eventId: ")
			return
		}
	}()

	fmt.Println("5. Finish execute")
	// Wait for the result of the submission
	select {
	case ccEvent := <-notifier:
		fmt.Println("notifier: ", ccEvent.Block.GetMetadata())
		// .GetHeader().String())
		// fmt.Println("Descriptor: ", ccEvent.Block.Descriptor)
	case <-time.After(time.Second * 20):
		fmt.Println("did NOT receive CC event for eventId: ")
		return
	}

}
