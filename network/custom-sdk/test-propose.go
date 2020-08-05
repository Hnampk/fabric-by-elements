// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"sync"
// 	"time"

// 	redis "github.com/go-redis/redis/v8"
// 	pcommon "github.com/hyperledger/fabric-protos-go/common"
// 	"github.com/hyperledger/fabric-protos-go/peer"
// 	pb "github.com/hyperledger/fabric-protos-go/peer"
// 	signerLib "github.com/hyperledger/fabric/cmd/common/signer"

// 	"github.com/hyperledger/fabric/protoutil"
// 	"google.golang.org/grpc"
// )

// type ClientWorker struct {
// 	id int
// 	// invokeChannel   chan InvokeRequest
// 	// responseChannel chan string
// }

// type InvokeRequest struct {
// 	FuncName string
// 	Args     []string
// }

// type QueryRequest struct {
// 	FuncName string
// 	Name     string
// }

// type ProposalWrapper struct {
// 	TxID     string
// 	Prop     RawProposal
// 	Response ProposalResponse
// }
// type RawProposal *peer.Proposal
// type ProposalResponse *pb.ProposalResponse

// var workerNum int // number of Client
// var loop int
// var account string = "nam1"
// var invokeChannel chan InvokeRequest
// var responseChannel chan string

// const chaincodeName string = "mycc"

// const rootURL string = "/home/ewallet/network/"

// // const rootURL string = "/home/nampkh/nampkh/my-fabric/network/"

// var ctx = context.Background()
// var rdb *redis.Client

// func main() {

// 	if len(os.Args) < 2 {
// 		fmt.Println("Please enter the Number of connections")
// 		return
// 	}

// 	var err error
// 	workerNum, err = strconv.Atoi(os.Args[1])
// 	if err != nil {
// 		fmt.Println("An error occurred: ", err)
// 		return
// 	}

// 	if len(os.Args) < 3 {
// 		fmt.Println("Please enter the Number of loop per connection")
// 		return
// 	}

// 	loop, err = strconv.Atoi(os.Args[2])
// 	if err != nil {
// 		fmt.Println("An error occurred: ", err)
// 		return
// 	}

// 	// if len(os.Args) < 4 {
// 	// 	fmt.Println("Please enter the account")
// 	// 	return
// 	// }

// 	// account = os.Args[3]
// 	// if err != nil {
// 	// 	fmt.Println("An error occurred: ", err)
// 	// 	return
// 	// }

// 	// redis setup
// 	rdb = redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})

// 	invokeChannel = make(chan InvokeRequest)
// 	responseChannel = make(chan string)
// 	channelInside := make(chan int)

// 	// create a new handler

// 	var mainWait sync.WaitGroup

// 	for i := 0; i < workerNum; i++ {

// 		worker := ClientWorker{
// 			id: i,
// 			// invokeChannel:   invokeChannel,
// 			// responseChannel: responseChannel,
// 		}

// 		mainWait.Add(1)
// 		go worker.start(&mainWait, channelInside)

// 		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)
// 	}

// 	// mainWait.Wait()
// 	start := time.Now()
// 	for i := 0; i < workerNum; i++ {
// 		<-channelInside
// 	}

// 	fmt.Println("duration: ", time.Now().Sub(start))
// }

// func (c *ClientWorker) start(mainWait *sync.WaitGroup, channelInside chan int) {

// 	defer mainWait.Done()
// 	// var connWait sync.WaitGroup
// 	// connWait.Add(loop)

// 	// invoke
// 	fcn := "update"
// 	var args [][]byte

// 	if chaincodeName == "mycc" {
// 		args = [][]byte{[]byte(fcn), []byte(account), []byte("1"), []byte("+")}
// 	} else {
// 		args = [][]byte{[]byte(fcn)}
// 	}

// 	responseChannel := make(chan []byte)

// 	cc, _ := grpc.Dial("peer0.org1.example.com:7051", grpc.WithInsecure())
// 	defer cc.Close()
// 	endorser := pb.NewEndorserClient(cc)
// 	endorserClients := []pb.EndorserClient{endorser}

// 	for i := 0; i < loop; i++ {
// 		func() {
// 			// defer connWait.Done()
// 			c.exec(args, responseChannel, endorserClients)
// 		}()
// 	}
// 	// connWait.Wait()

// 	channelInside <- 1
// }

// func (c *ClientWorker) exec(args [][]byte, responseChannel chan []byte, endorserClients []pb.EndorserClient) {
// 	signerConfig := signerLib.Config{
// 		MSPID:        "Org1MSP",
// 		IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
// 		KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
// 	}

// 	signer, err := signerLib.NewSigner(signerConfig)

// 	chaincodeLang := "GOLANG"

// 	if err != nil {
// 		fmt.Println("[ERROR] NewSigner:", err)
// 		return
// 	}

// 	testInput := pb.ChaincodeInput{
// 		IsInit: false,
// 		Args:   args,
// 	}

// 	spec := &pb.ChaincodeSpec{
// 		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
// 		ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
// 		Input:       &testInput,
// 	}

// 	// Build the ChaincodeInvocationSpec message
// 	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

// 	creator, err := signer.Serialize()
// 	if err != nil {
// 		fmt.Println("[ERROR] Serialize:", err)
// 		return
// 	}

// 	// extract the transient field if it exists
// 	var tMap map[string][]byte

// 	cID := "vnpay-channel"
// 	txID := ""

// 	prop, txid, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(pcommon.HeaderType_ENDORSER_TRANSACTION, cID, invocation, creator, txID, tMap)
// 	if err != nil {
// 		fmt.Println("[ERROR]: CreateChaincodeProposalWithTxIDAndTransient", err)
// 		return
// 	}

// 	signedProp, err := protoutil.GetSignedProposal(prop, signer)
// 	if err != nil {
// 		fmt.Println("[ERROR]: GetSignedProposal", err)
// 		return
// 	}

// 	// response payload
// 	responses, err := processProposals(endorserClients, *signedProp)
// 	if err != nil || len(responses) < 1 {
// 		// 	// responseChannel <- "Timeout" // fix me
// 		fmt.Println("ERROR occured!", err)
// 		return
// 	}
// 	//
// 	fmt.Println("TXID:", time.Now(), txid)
// 	rawProposal := RawProposal(prop)
// 	proposalResponse := ProposalResponse(responses[0])
// 	response := ProposalWrapper{TxID: txid, Prop: rawProposal, Response: proposalResponse}

// 	// save byte slice into Redis
// 	responseByte, err := json.Marshal(response)

// 	if err != nil {
// 		fmt.Println("error:", err.Error())
// 		return
// 	}

// 	rdb.RPush(ctx, "pending-proposals", responseByte)
// 	return
// }

// // processProposals sends a signed proposal to a set of peers, and gathers all the responses.
// func processProposals(endorserClients []pb.EndorserClient, signedProposal pb.SignedProposal) ([]*pb.ProposalResponse, error) {
// 	responsesCh := make(chan *pb.ProposalResponse, len(endorserClients))
// 	errorCh := make(chan error, len(endorserClients))
// 	wg := sync.WaitGroup{}
// 	for _, endorser := range endorserClients {
// 		wg.Add(1)
// 		go func(endorser pb.EndorserClient) {
// 			defer wg.Done()
// 			proposalResp, err := endorser.ProcessProposal(context.Background(), &signedProposal)
// 			if err != nil {
// 				errorCh <- err
// 				return
// 			}
// 			responsesCh <- proposalResp
// 		}(endorser)
// 	}
// 	wg.Wait()
// 	close(responsesCh)
// 	close(errorCh)
// 	for err := range errorCh {
// 		return nil, err
// 	}
// 	var responses []*pb.ProposalResponse
// 	for response := range responsesCh {
// 		responses = append(responses, response)
// 	}
// 	return responses, nil
// }
