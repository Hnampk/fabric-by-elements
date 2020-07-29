package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"example.com/custom-sdk/fabric/usable-inter-nal/peer/common"
	"example.com/custom-sdk/fabric/usable-inter-nal/pkg/comm"
	"github.com/go-redis/redis/v8"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	signerLib "github.com/hyperledger/fabric/cmd/common/signer"

	"github.com/hyperledger/fabric/protoutil"
)

type ClientWorker struct {
	id int
	// invokeChannel   chan InvokeRequest
	// responseChannel chan string
}

type InvokeRequest struct {
	FuncName string
	Args     []string
}

type QueryRequest struct {
	FuncName string
	Name     string
}

type ProposalWrapper struct {
	Prop     RawProposal
	Response ProposalResponse
}
type RawProposal *peer.Proposal
type ProposalResponse *pb.ProposalResponse

var workerNum int // number of Client
var loop int
var invokeChannel chan InvokeRequest
var responseChannel chan string

const configFile string = "network.yaml"
const adminUser string = "Admin"
const OrgName string = "org1"
const org1User string = "Admin"
const channelID string = "vnpay-channel"
const chainCodeID string = "mycc"

const rootURL string = "/home/ewallet/network/"

// const rootURL string = "/home/nampkh/nampkh/my-fabric/network/"

var ctx = context.Background()
var rdb *redis.Client

func main() {
	var err error
	workerNum, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Please enter the Number of connections")
		return
	}

	loop, err = strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Please enter the Number of loops per connection")
		return
	}

	// redis setup
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	invokeChannel = make(chan InvokeRequest)
	responseChannel = make(chan string)
	channelInside := make(chan int)

	for i := 0; i < workerNum; i++ {

		worker := ClientWorker{
			id: i,
			// invokeChannel:   invokeChannel,
			// responseChannel: responseChannel,
		}

		go worker.start(channelInside)

		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)
	}
	fmt.Println("START")

	// create a new handler
	start := time.Now()

	for i := 0; i < workerNum; i++ {
		<-channelInside
	}

	fmt.Println("duration: ", time.Now().Sub(start))
}

func (c *ClientWorker) start(channelInside chan int) {

	address := "orderer.example.com:7050"
	// address := "10.22.7.230:7050"
	// override := ""
	clientConfig := comm.ClientConfig{}
	clientConfig.Timeout = 3 * time.Second
	secOpts := comm.SecureOptions{}
	clientConfig.SecOpts = secOpts

	gClient, err := comm.NewGRPCClient(clientConfig)

	if err != nil {
		fmt.Println("[ERROR] failed to load config for OrdererClient")
		return
	}

	oc := &common.OrdererClient{
		CommonClient: common.CommonClient{
			GRPCClient: gClient,
			Address:    address,
			// sn:         override
		}}

	bc, err := oc.Broadcast()

	broadcastClient := &common.BroadcastGRPCClient{Client: bc}
	defer broadcastClient.Close()

	signerConfig := signerLib.Config{
		MSPID:        "Org1MSP",
		IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
		KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
	}

	signer, err := signerLib.NewSigner(signerConfig)
	if err != nil {
		fmt.Println("ERROR NewSigner", err)
		return
	}

	// send
	for i := 0; i < loop; i++ {
		send(broadcastClient, signer)
	}

	channelInside <- 1
}

func send(broadcastClient *common.BroadcastGRPCClient, signer *signerLib.Signer) {
	// send the first signed proposal response
	var proposalWrapper ProposalWrapper
	redisData, err := rdb.LPop(ctx, "pending-proposals").Bytes()

	if err != nil {
		fmt.Println("No more pending proposal response!", 500)
		return
	}

	json.Unmarshal(redisData, &proposalWrapper)
	rawProposal := proposalWrapper.Prop
	proposalResponse := proposalWrapper.Response

	if proposalResponse != nil {
		if proposalResponse.Response.Status >= shim.ERRORTHRESHOLD {
			fmt.Println("shim.ERRORTHRESHOLD", 500)
			return
		}

		// assemble a signed transaction (it's an Envelope message)
		env, err := protoutil.CreateSignedTx(rawProposal, signer, proposalResponse)
		if err != nil {
			fmt.Println("ERROR CreateSignedTx", err)
			return
		}

		// var dg *DeliverGroup
		// var ctx context.Context
		// if waitForEvent {
		// 	var cancelFunc context.CancelFunc
		// 	ctx, cancelFunc = context.WithTimeout(context.Background(), waitForEventTimeout)
		// 	defer cancelFunc()

		// 	dg = NewDeliverGroup(
		// 		deliverClients,
		// 		peerAddresses,
		// 		signer,
		// 		certificate,
		// 		channelID,
		// 		txid,
		// 	)
		// 	// connect to deliver service on all peers
		// 	err := dg.Connect(ctx)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }

		// send the envelope for ordering
		if err = broadcastClient.Send(env); err != nil {
			fmt.Println("error sending transaction for update function", 500)
			// broadcastClient.Close()
			return
		}
		// broadcastClient.Close()
		// if dg != nil && ctx != nil {
		// 	// wait for event that contains the txid from all peers
		// 	err = dg.Wait(ctx)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
		fmt.Println("OK")
	}
	return
}
