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
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	// mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	// fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	// contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	// packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	// fabpeer "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-protos-go/msp"
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
	OrgAdmin := "admin"
	OrgName := "org1"
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
	// peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	// if err != nil {
	// 		fmt.Printf("failed to create peer: %s\n", err)
	// 		return
	// }

	//Try to get genesis
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
	//
	// peer, err := fabpeer.New(endpointConfig,fabpeer.WithURL(peerURL),fabpeer.WithMSPID(OrgName), fabpeer.WithServerName(serverName))
	// if err != nil {
	// 		fmt.Printf("failed to create peer: %s\n", err)
	// 		return
	// }

	// chaincodeGoPath := "/Users/nguyenthanhbinh/Blockchain/test/hyperledger/fabric-samples/balance-transfer/artifacts"
	chaincodeGoPath := "../chaincode/abstore"

	// Create the chaincode package that will be sent to the peers
	// ccPkg, err := packager.NewCCPackage(chaincodePath, chaincodeGoPath)
	// if err != nil {
	// 	fmt.Println(err, "failed to create chaincode package")
	// }
	// fmt.Println("ccPkg created")

	chainCodeID := "mycc"

	// Set up chaincode policy
	// ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP","Org2MSP"})
	ccPolicy := cauthdsl.SignedByNOutOfGivenRole(int32(2), msp.MSPRole_MEMBER, []string{"Org1MSP", "Org2MSP"})

	channelID := "vnpay-channel"
	args := [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte("200")}

	for i, t := range args {
		fmt.Println(i, string(t))
	}
	version := "v0"
	req := resmgmt.InstantiateCCRequest{Name: chainCodeID, Path: chaincodeGoPath, Version: version, Args: args, Policy: ccPolicy}
	// resp, err := resMgmtClient.InstantiateCC(channelID, req, resmgmt.WithTargets(peer))
	resp, err := resMgmtClient.InstantiateCC(channelID, req)
	if err != nil || resp.TransactionID == "" {
		fmt.Println(err, "failed to instantiate the chaincode")
		return
	}
	fmt.Println("Chaincode instantiated")
}
