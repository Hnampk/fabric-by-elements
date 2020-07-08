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
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
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
		 return
	}
	fmt.Println("3.2 get admin Identity .")
	adminIdentity, err := mspClient.GetSigningIdentity(OrgAdmin)
	if err != nil {
		 fmt.Println("failed to get admin signing identity: ", err)
	}
	fmt.Println(adminIdentity.Identifier().ID)
	fmt.Println(adminIdentity.Identifier().MSPID)

	// fmt.Println("3.3 Test get certificate  .")
  // pubPEMData := adminIdentity.EnrollmentCertificate()
	// s := string(pubPEMData)
	// fmt.Println(s)
	// block, rest := pem.Decode(pubPEMData)
	// fmt.Println(block)
	// fmt.Println(rest)

	// if block == nil || block.Type != "CERTIFICATE" {
	// 	fmt.Println("failed to decode PEM block containing certificate")
	// }
	//
	// cert, err := x509.ParseCertificate(block.Bytes)
	// if err != nil {
	// 	panic("failed to parse certificate: " + err.Error())
	// }
	// fmt.Println(cert)

	// r, err := os.Open(channelConfigTxPath)
	// if err != nil {
	//     fmt.Println("failed to open channel config: ", err)
	// 		return
	// }
	// defer r.Close()
	// Read channel configuration tx
	channelID := "vnpay-channel"
	channelConfigTxPath := "../../../artifacts/channel/vnpay-channel.tx"

	// Create new channel 'mychannel'
	ordererID := "orderer.example.com"
	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfigPath: channelConfigTxPath, SigningIdentities: []msp.SigningIdentity{adminIdentity}}

	fmt.Println("3.3.1 Call request to create channel   .")
	_, err = resMgmtClient.SaveChannel(req, resmgmt.WithOrdererEndpoint(ordererID))
	if err != nil {
	    fmt.Println("failed to save channel:", err)
			return
	}

	fmt.Println("Finish create channel")
}
