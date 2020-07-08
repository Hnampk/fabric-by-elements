package main

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	// mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	//  "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	// fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

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

	// fmt.Println("3.1 Create a msp client instance using its New func, passing the context.")
	// clientProvider := sdk.Context()
	// mspClient, err := mspclient.New(clientProvider, mspclient.WithOrg(OrgName))
	// if err != nil {
	// 	 fmt.Println( "failed to create MSP client", err)
	// }
	// fmt.Println("3.2 get admin Identity .")
	//
	//
	// adminIdentity, err := mspClient.GetSigningIdentity(OrgAdmin)
	// if err != nil {
	// 	 fmt.Println("failed to get admin signing identity: ", err)
	// }
	// fmt.Println(adminIdentity.Identifier().ID)
	// fmt.Println(adminIdentity.Identifier().MSPID)

// Try to get genesis
	// cfgBackend, err := sdk.Config()
	// if err != nil {
	// 		fmt.Printf("failed to get config backend: %s\n", err)
	// }
	// endpointConfig, err := fab.ConfigFromBackend(cfgBackend)
	// if err != nil {
	// 		fmt.Printf("failed to get endpointConfig: %s\n", err)
	// }

	// serverName := "peer0.org1.example.com"
  // peerURL := "grpcs://peer0.org1.example.com:7051"

	// serverName := "peer1.org1.example.com"
	// peerURL := "grpcs://peer1.org1.example.com:7056"
	// peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	// if err != nil {
	// 		fmt.Printf("failed to create peer: %s\n", err)
	// 		return
	// }

	chaincodeGoPath := "/Users/nguyenthanhbinh/Blockchain/test/hyperledger/fabric-samples/balance-transfer/artifacts"
	// chaincodePath := "github.com/test_cc/go"
	chaincodePath := "vnpay.vn/bilateralchannel"

 	// Create the chaincode package that will be sent to the peers
 	ccPkg, err := packager.NewCCPackage(chaincodePath, chaincodeGoPath)
 	if err != nil {
 		fmt.Println(err, "failed to create chaincode package")
 	}
 	fmt.Println("ccPkg created")

	chainCodeID := "bilateralchannel"
	version := "v0"

	req := resmgmt.InstallCCRequest{Name: chainCodeID, Version: version, Path: chaincodePath, Package: ccPkg}
	// _, err = resMgmtClient.InstallCC(req, resmgmt.WithTargets(peer),  resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	_, err = resMgmtClient.InstallCC(req,  resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
	    fmt.Printf("failed to install chaincode: %s\n", err)
			return
	}
	fmt.Println("Chaincode installed")
}
