package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	pcommon "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	signerLib "github.com/hyperledger/fabric/cmd/common/signer"

	"github.com/hyperledger/fabric/protoutil"
	"google.golang.org/grpc"
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

const workerNum = 200 // number of Client
var invokeChannel chan InvokeRequest
var responseChannel chan string

const configFile string = "network.yaml"
const adminUser string = "Admin"
const OrgName string = "org1"
const org1User string = "Admin"
const channelID string = "vnpay-channel"
const chainCodeID string = "mycc"

var ctx = context.Background()
var rdb *redis.Client

func main() {
	// redis setup
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	invokeChannel = make(chan InvokeRequest)
	responseChannel = make(chan string)
	channelInside := make(chan int)

	// create a new handler

	start := time.Now()
	for i := 0; i < workerNum; i++ {

		worker := ClientWorker{
			id: i,
			// invokeChannel:   invokeChannel,
			// responseChannel: responseChannel,
		}

		go worker.start(channelInside)

		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)
	}

	for i := 0; i < workerNum; i++ {
		<-channelInside
	}

	fmt.Println("duration: ", time.Now().Sub(start))
}

func (c *ClientWorker) start(channelInside chan int) {
	loop := 300

	// invoke
	fcn := "update"
	args := [][]byte{[]byte(fcn), []byte("nam4"), []byte("1"), []byte("+")}

	responseChannel := make(chan []byte)

	for i := 0; i < loop; i++ {
		c.exec(args, responseChannel)
	}

	channelInside <- 1
}

func (c *ClientWorker) exec(args [][]byte, responseChannel chan []byte) {
	signerConfig := signerLib.Config{
		MSPID:        "Org1MSP",
		IdentityPath: "/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
		KeyPath:      "/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
	}

	signer, err := signerLib.NewSigner(signerConfig)

	chaincodeLang := "GOLANG"
	chaincodeName := "mycc"

	if err != nil {
		fmt.Println("[ERROR] NewSigner:", err)
		return
	}

	testInput := pb.ChaincodeInput{
		IsInit: false,
		Args:   args,
	}

	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
		ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input:       &testInput,
	}

	// Build the ChaincodeInvocationSpec message
	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	creator, err := signer.Serialize()
	if err != nil {
		fmt.Println("[ERROR] Serialize:", err)
		return
	}

	// extract the transient field if it exists
	var tMap map[string][]byte

	cID := "vnpay-channel"
	txID := ""

	prop, txid, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(pcommon.HeaderType_ENDORSER_TRANSACTION, cID, invocation, creator, txID, tMap)
	if err != nil {
		fmt.Println("[ERROR]: CreateChaincodeProposalWithTxIDAndTransient", err)
		return
	}
	fmt.Println("TXID:", txid)

	signedProp, err := protoutil.GetSignedProposal(prop, signer)
	if err != nil {
		fmt.Println("[ERROR]: GetSignedProposal", err)
		return
	}

	// cc, _ := grpc.Dial("peer0.org1.example.com:7051", grpc.WithInsecure())
	cc, _ := grpc.Dial("10.22.7.231:7051", grpc.WithInsecure())
	endorser := pb.NewEndorserClient(cc)
	mockClients := []pb.EndorserClient{endorser}

	// response payload
	responses, err := processProposals(mockClients, signedProp)
	if err != nil || len(responses) < 1 {
		// responseChannel <- "Timeout" // fix me
		cc.Close()
		return
	}
	cc.Close()

	rawProposal := RawProposal(prop)
	proposalResponse := ProposalResponse(responses[0])
	response := ProposalWrapper{Prop: rawProposal, Response: proposalResponse}

	// save byte slice into Redis
	responseByte, err := json.Marshal(response)

	if err != nil {
		fmt.Println("error:", err.Error())
		// responseChannel <- "Timeout" // fix me
		return
	}
	// fmt.Println(responseByte)

	rdb.RPush(ctx, "pending-proposals", responseByte)
	// responseChannel <- responseByte
	return
}

// processProposals sends a signed proposal to a set of peers, and gathers all the responses.
func processProposals(endorserClients []pb.EndorserClient, signedProposal *pb.SignedProposal) ([]*pb.ProposalResponse, error) {
	responsesCh := make(chan *pb.ProposalResponse, len(endorserClients))
	errorCh := make(chan error, len(endorserClients))
	wg := sync.WaitGroup{}
	for _, endorser := range endorserClients {
		wg.Add(1)
		go func(endorser pb.EndorserClient) {
			defer wg.Done()
			proposalResp, err := endorser.ProcessProposal(context.Background(), signedProposal)
			if err != nil {
				errorCh <- err
				return
			}
			responsesCh <- proposalResp
		}(endorser)
	}
	wg.Wait()
	close(responsesCh)
	close(errorCh)
	for err := range errorCh {
		return nil, err
	}
	var responses []*pb.ProposalResponse
	for response := range responsesCh {
		responses = append(responses, response)
	}
	return responses, nil
}
