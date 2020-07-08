package main

import (
	"fmt"
	// "time"
	// "os"
	// "strconv"
	// "math/rand"
	// "crypto/x509"
	// "encoding/pem"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	//  "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	// fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
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
	OrgName := "org1"
	resourceManagerClientContext := sdk.Context(fabsdk.WithUser(OrgAdmin), fabsdk.WithOrg(OrgName))


	fmt.Println("3. Create a resource client instance using its New func, passing the context.")
	// Create new resource management client
	resMgmtClient, err := resmgmt.New(resourceManagerClientContext)
	if err != nil {
	    fmt.Println("failed to create resource client: ", err)
	}

	fmt.Println("3.1 Create a msp client instance using its New func, passing the context.")
	clientProvider := sdk.Context()
	mspClient, err := mspclient.New(clientProvider, mspclient.WithOrg(OrgName))
	if err != nil {
		 fmt.Println( "failed to create MSP client", err)
	}

	fmt.Println("3.2 get admin Identity .")
	adminIdentity, err := mspClient.GetSigningIdentity(OrgAdmin)
	if err != nil {
		 fmt.Println("failed to get admin signing identity: ", err)
	}
	fmt.Println(adminIdentity.Identifier().ID)
	fmt.Println(adminIdentity.Identifier().MSPID)

// Try to get genesis
	// cfgBackend, err := sdk.Config()
	// if err != nil {
	// 		fmt.Printf("failed to get config backend: %s\n", err)
	// }
	// endpointConfig, err := fab.ConfigFromBackend(cfgBackend)
	// if err != nil {
	// 		fmt.Printf("failed to get endpointConfig: %s\n", err)
	// }

	//	fmt.Println("3.3 Test get genesis block .")
	// ordererCfgs := endpointConfig.OrderersConfig()
	// if len(ordererCfgs) == 0 {
	// 	fmt.Printf("failed to find orderer in config:")
	// }
	// randomNumber := rand.Intn(len(ordererCfgs))
	// ordererCfg := &ordererCfgs[randomNumber]
	//
	// ctx, err := resourceManagerClientContext()
	// if err != nil {
	// 		fmt.Printf("Cannot get context : %s\n", err)
	// 		return
	// }
	// ordererResponseTimeout :=  time.Minute * 2
	//
	// fmt.Println(ordererResponseTimeout)
	//
	// ordrReqCtx, ordrReqCtxCancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(ordererResponseTimeout))
	// defer ordrReqCtxCancel()
	//
	// // fmt.Println(ordererCfg)
	// orderer, err := ctx.InfraProvider().CreateOrdererFromConfig(ordererCfg)
	// if err != nil {
	// 		fmt.Printf("Cannot create order : %s\n", err)
	// 		return
	// }
  // fmt.Println(ordrReqCtx)
  // fmt.Println(orderer)
	// genesisBlock, err := resource.GenesisBlockFromOrderer(ordrReqCtx, channelID, orderer)
	// if err != nil {
	// 	fmt.Println("genesis block retrieval failed: ", err)
	// 	return
	// }
	// fmt.Println("Tet")
	// fmt.Println(genesisBlock)
	// return

	// serverName := "peer0.org1.example.com"
  // peerURL := "grpcs://peer0.org1.example.com:7051"

	// serverName := "peer1.org1.example.com"
	// peerURL := "grpcs://peer1.org1.example.com:7056"
	//
	// peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	// if err != nil {
	// 		fmt.Printf("failed to create peer: %s\n", err)
	// 		return
	// }

	// Peer joins channel 'mychannel'
	channelID := "vnpay-channel"
	ordererID := "orderer.example.com"
	// err = resMgmtClient.JoinChannel(channelID, resmgmt.WithTargets(peer))
	err = resMgmtClient.JoinChannel(channelID, resmgmt.WithOrdererEndpoint(ordererID))
	if err != nil {
	    fmt.Printf("failed to join channel: %s\n", err)
			return
	}
	fmt.Println("Finish join channel")
}
