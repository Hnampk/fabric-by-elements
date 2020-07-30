package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const workerNum = 200 // number of Client
const loop = 50

var configFile string = "network.yaml"
var adminUser string = "Admin"
var OrgName string = "org1"
var org1User string = "Admin"
var channelID string = "vnpay-channel"
var chainCodeID string = "mycc4"
var fcn string = "update"

type ClientWorker struct {
	Id     int
	Client *channel.Client
	// responseChannel chan string
	EventListener *event.Client
}

func main() {
	configProvider := config.FromFile(configFile)
	workerPool := []ClientWorker{}

	var mainWait sync.WaitGroup
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
			return
		}

		eventListener, err := event.New(clientContext, event.WithBlockEvents())

		if err != nil {
			fmt.Println(err, "failed to create new event client")
			return
		}

		worker := ClientWorker{
			Id:     i,
			Client: client,
			// responseChannel: responseChannel,
			EventListener: eventListener,
		}
		mainWait.Add(1)
		workerPool = append(workerPool, worker)
		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)
	}

	fmt.Println("START")

	for _, worker := range workerPool {
		go worker.start(&mainWait)
	}

	start := time.Now()
	defer func() {
		fmt.Println("duration: ", time.Now().Sub(start))
		fmt.Println("count: ", count)
	}()

	mainWait.Wait()
}

func (c ClientWorker) start(mainWait *sync.WaitGroup) {
	defer mainWait.Done()

	var connWait sync.WaitGroup
	connWait.Add(loop)

	registration, notifier, err := c.EventListener.RegisterChaincodeEvent(chainCodeID, "updateEvent")

	if err != nil {
		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to RegisterChaincodeEvent: ", err, c.Id)
		// responseChannel <- err.Error()
		// c.eventListener.Unregister(registration)
		return
	}
	defer c.EventListener.Unregister(registration)

	for i := 0; i < loop; i++ {

		c.exec(&connWait, notifier)
	}
	connWait.Wait()
}

var count = 0

func (c *ClientWorker) exec(connWait *sync.WaitGroup, notifier <-chan *fab.CCEvent) {

	defer connWait.Done()

	var req channel.Request
	if chainCodeID == "mycc" {
		req = channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: [][]byte{[]byte("account5"), []byte("1"), []byte("+")}}
	} else {
		req = channel.Request{ChaincodeID: chainCodeID, Fcn: fcn, Args: [][]byte{}}
	}

	response, err := c.Client.Execute(req, channel.WithTargetEndpoints("peer0.org1.example.com"))
	if err != nil {
		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]failed to Execute: ", err)
		// responseChannel <- err.Error()
		return
	}
	// fmt.Println(response.TransactionID)

	func() {
		for {
			select {
			case ccEvent := <-notifier:

				if ccEvent.TxID != string(response.TransactionID) {
					continue
				}
				count++
				fmt.Println(time.Now().String() + "txid: " + string(response.TransactionID))

				return
			}
		}
	}()
}
