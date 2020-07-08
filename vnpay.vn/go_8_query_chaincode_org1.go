package main

import (
	"fmt"
  "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
  "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
  // "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	// mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	// "github.com/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	 fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
)


func main() {
	fmt.Println("1. Instantiate a fabsdk instance using a configuration.")
	// Create SDK setup for the integration tests
	configFile := "network.yaml"
	configProvider := config.FromFile(configFile)
  sdk, err1 := fabsdk.New(configProvider)
	if err1 != nil {
			fmt.Println("failed to create sdk",err1)
			return
	}
	defer sdk.Close()

	fmt.Println("2. Prepare client context.")
	fmt.Println("2.1 Create a context based on a user and organization, using your fabsdk instance.")

	org1User := "admin"
	OrgName := "org1"
	channelID := "vnpay-channel"
	channelProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(OrgName))


	fmt.Println("3. Create a client instance using its New func, passing the context.")
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
	peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	if err != nil {
			fmt.Printf("failed to create peer: %s\n", err)
			return
	}


		fmt.Println("4. Use the funcs provided by each client to create your solution.")
		chainCodeID := "test1"
		fcn := "query"
		args :=  [][]byte{[]byte("b")}
		req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}
		response, err := client.Query(req, channel.WithTargets(peer))
		if err != nil {
		    fmt.Println("failed to query chaincode: ", err)
				return
		}

		fmt.Println(string(response.Payload))

		fmt.Println("5. Finish query")

}
