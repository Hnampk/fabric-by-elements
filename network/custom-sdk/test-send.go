package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"example.com/custom-sdk/fabric/usable-inter-nal/peer/chaincode"
	"example.com/custom-sdk/fabric/usable-inter-nal/peer/common"
	"example.com/custom-sdk/fabric/usable-inter-nal/pkg/comm"
	"example.com/custom-sdk/fabric/usable-inter-nal/pkg/identity"
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
	TxID     string
	Prop     RawProposal
	Response ProposalResponse
}
type RawProposal *peer.Proposal
type ProposalResponse *pb.ProposalResponse

var workerNum int // number of Client
var loop int
var invokeChannel chan InvokeRequest
var responseChannel chan string

const channelID string = "vnpay-channel"

const rootURL string = "/home/ewallet/network/"

// const rootURL string = "/home/nampkh/nampkh/my-fabric/network/"
const waitForEvent = true

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
		fmt.Println("No more pending proposal response!", err)
		return
	}

	json.Unmarshal(redisData, &proposalWrapper)
	txid := proposalWrapper.TxID
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

		var dg *chaincode.DeliverGroup

		if waitForEvent {
			deliverClients := []pb.DeliverClient{}

			deliverClient, err := common.GetPeerDeliverClientFnc("peer0.org1.example.com:7051", "tlsRootCertFile")
			if err != nil {
				fmt.Println("error getting deliver client for", err)
				return
			}
			deliverClients = append(deliverClients, deliverClient)
			certificate, err := common.GetCertificateFnc()
			// connect to deliver service on all peers
			if err != nil {
				fmt.Println("error GetCertificateFnc", err)
				return
			}

			waitForEventTimeout := 30 * time.Second
			var cancelFunc context.CancelFunc
			ctx, cancelFunc = context.WithTimeout(context.Background(), waitForEventTimeout)
			defer cancelFunc()

			dg = NewDeliverGroup(
				deliverClients,
				[]string{""},
				signer,
				certificate,
				channelID,
				txid,
			)

			// connect to deliver service on all peers
			err = dg.Connect(ctx)
			if err != nil {
				fmt.Println("error Connect", err)
				return
			}
		}

		// send the envelope for ordering
		if err = broadcastClient.Send(env); err != nil {
			fmt.Println("error sending transaction for update function", 500, err)
			return
		}

		// broadcastClient.Close()
		if dg != nil && ctx != nil {
			// wait for event that contains the txid from all peers
			err = dg.Wait(ctx)
			if err != nil {
				fmt.Println("error Wait", 500, err)
				return
			}
		}

		fmt.Println("OK")
	}
	return
}

func NewDeliverGroup(
	deliverClients []pb.DeliverClient,
	peerAddresses []string,
	signer identity.SignerSerializer,
	certificate tls.Certificate,
	channelID string,
	txid string,
) *chaincode.DeliverGroup {
	clients := make([]*chaincode.DeliverClient, len(deliverClients))
	for i, client := range deliverClients {
		dc := &chaincode.DeliverClient{
			Client:  client,
			Address: peerAddresses[i],
		}
		clients[i] = dc
	}

	dg := &chaincode.DeliverGroup{
		Clients:     clients,
		Certificate: certificate,
		ChannelID:   channelID,
		TxID:        txid,
		Signer:      signer,
	}

	return dg
}
