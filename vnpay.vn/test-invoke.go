package main

import (
	"fmt"
	"runtime"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type ClientWorker struct {
	id     int
	client *channel.Client
}

var configFile string = "network.yaml"
var adminUser string = "admin"
var OrgName string = "org1"
var org1User string = "admin"
var channelID string = "vnpay-channel"
var chainCodeID string = "mycc"

const workerNum = 100 // number of Client

func main() {
	runtime.GOMAXPROCS(4)
	configProvider := config.FromFile(configFile)

	channelInside := make(chan int)

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

		// eventListener, err := event.New(clientContext, event.WithBlockEvents())
		// if err != nil {
		// 	fmt.Println(err, "failed to create new event client")
		// 	// return
		// }

		worker := ClientWorker{
			id:     i,
			client: client,
		}

		if err != nil {
			fmt.Println("failed to create channel client: ", err)
			return
		}

		go worker.start(channelInside)
		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)

	}

	for i := 0; i < workerNum; i++ {
		<-channelInside
	}

}

func (c *ClientWorker) start(channelInside chan int) {
	fcn := "update"
	argsStr := []string{"nam2", "1", "+"}
	args := [][]byte{}

	for _, arg := range argsStr {
		args = append(args, []byte(arg))
	}

	req := channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: args}

	for i := 0; i < 100; i++ {
		 c.client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
//		if err != nil {
//			fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to query chaincode: ", err)
//		}
//		fmt.Println(string("txid: " + response.TransactionID))
	}

	channelInside <- 9999
}
