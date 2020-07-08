package main

import (
	"fmt"
  "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
  // "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
  // "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	// mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
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
	OrgAdmin := "admin"
	OrgName := "org2"
	resourceManagerClientContext := sdk.Context(fabsdk.WithUser(OrgAdmin), fabsdk.WithOrg(OrgName))

	fmt.Println("3. Create a resource client instance using its New func, passing the context.")
	// Create new resource management client
	resMgmtClient, err := resmgmt.New(resourceManagerClientContext)
	if err != nil {
			fmt.Println("failed to create resource client: ", err)
	}

	fmt.Println("4. Create peer config.")
	// clientProvider := sdk.Context()
	// mspClient, err := mspclient.New(clientProvider, mspclient.WithOrg(OrgName))
	// if err != nil {
	// 	 fmt.Println( "failed to create MSP client", err)
	// }
	// adminIdentity, err := mspClient.GetSigningIdentity(OrgAdmin)
	// if err != nil {
	// 	 fmt.Println("failed to get admin signing identity: ", err)
	// }
	// fmt.Println(adminIdentity.Identifier().ID)
	// fmt.Println(adminIdentity.Identifier().MSPID)


	cfgBackend, err := sdk.Config()
	if err != nil {
			fmt.Printf("failed to get config backend: %s\n", err)
	}
	endpointConfig, err := fab.ConfigFromBackend(cfgBackend)
	if err != nil {
			fmt.Printf("failed to get endpointConfig: %s\n", err)
	}

	serverName := "peer0.org2.example.com"
	peerURL := "grpcs://peer0.org2.example.com:8051"

	// serverName := "peer1.org2.example.com"
	// peerURL := "grpcs://peer1.org2.example.com:8056"

	peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	if err != nil {
			fmt.Printf("failed to create peer: %s\n", err)
			return
	}
	fmt.Println("5. Query peer info.")
	response, err := resMgmtClient.QueryChannels(resmgmt.WithTargets(peer))
	if err != nil {
	    fmt.Printf("failed to query channels: %s\n", err)
			return
	}
	if response != nil {
	    fmt.Println("Retrieved channels")
			for i, channel := range response.Channels {
						fmt.Println(i, channel)
			}
	}
  fmt.Println("End.")
}
