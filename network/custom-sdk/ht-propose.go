// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"example.com/custom-sdk/fabric/usable-inter-nal/peer/common"
// 	"example.com/custom-sdk/fabric/usable-inter-nal/pkg/comm"
// 	"github.com/go-redis/redis/v8"
// 	"github.com/hyperledger/fabric-chaincode-go/shim"
// 	pcommon "github.com/hyperledger/fabric-protos-go/common"
// 	"github.com/hyperledger/fabric-protos-go/peer"
// 	pb "github.com/hyperledger/fabric-protos-go/peer"
// 	signerLib "github.com/hyperledger/fabric/cmd/common/signer"

// 	"github.com/hyperledger/fabric/protoutil"
// 	"google.golang.org/grpc"
// )

// type ClientWorker struct {
// 	id              int
// 	invokeChannel   chan InvokeRequest
// 	responseChannel chan string
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
// 	Prop     RawProposal
// 	Response ProposalResponse
// }
// type RawProposal *peer.Proposal
// type ProposalResponse *pb.ProposalResponse

// const workerNum = 20 // number of Client
// var invokeChannel chan InvokeRequest
// var responseChannel chan string

// const configFile string = "network.yaml"
// const adminUser string = "Admin"
// const OrgName string = "org1"
// const org1User string = "Admin"
// const channelID string = "vnpay-channel"
// const chainCodeID string = "mycc"

// var ctx = context.Background()
// var rdb *redis.Client

// func main() {
// 	// redis setup
// 	rdb = redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})

// 	invokeChannel = make(chan InvokeRequest)
// 	responseChannel = make(chan string)

// 	// create a new handler

// 	for i := 0; i < workerNum; i++ {

// 		worker := ClientWorker{
// 			id:              i,
// 			invokeChannel:   invokeChannel,
// 			responseChannel: responseChannel,
// 		}

// 		go worker.start()

// 		fmt.Println(">>>>>>>>>>>>>>[CUSTOM]Started Client ", worker)
// 	}

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/invoke", invoke)
// 	mux.HandleFunc("/send", send)
// 	mux.HandleFunc("/count", count)
// 	// listen and serve
// 	http.ListenAndServe(":8090", mux)
// }

// func (c *ClientWorker) start() {
// 	for req := range c.invokeChannel {
// 		fcn := req.FuncName
// 		args := [][]byte{[]byte(fcn)}

// 		for _, element := range req.Args {
// 			args = append(args, []byte(element))
// 		}

// 		responseChannel := make(chan string)
// 		triggerStop := make(chan string)

// 		go c.exec(args, responseChannel, triggerStop)
// 		// go timeOutChecker(responseChannel)

// 		result := <-responseChannel

// 		if result == "Timeout" {
// 			triggerStop <- "done job"
// 			c.responseChannel <- "Timeout"
// 		} else {
// 			c.responseChannel <- result
// 		}

// 		close(responseChannel)
// 		close(triggerStop)
// 	}
// }

// func timeOutChecker(c chan string) {
// 	select {
// 	case <-time.After(time.Second * 5):
// 		c <- "Timeout"
// 	}
// }

// func invoke(w http.ResponseWriter, r *http.Request) {
// 	body, err := ioutil.ReadAll(r.Body)

// 	if err != nil {
// 		fmt.Printf("Error reading body: %v", err)
// 		http.Error(w, "can't read body", http.StatusBadRequest)
// 		return
// 	}

// 	var invokeRequest InvokeRequest
// 	json.Unmarshal(body, &invokeRequest)

// 	invokeChannel <- invokeRequest
// 	response := <-responseChannel

// 	if response == "Timeout" {
// 		http.Error(w, "tineout", 500)
// 	} else {
// 		fmt.Fprint(w, response)
// 	}

// 	return
// }

// func (c *ClientWorker) exec(args [][]byte, responseChannel chan string, timeout chan string) {
// 	signerConfig := signerLib.Config{
// 		MSPID:        "Org1MSP",
// 		IdentityPath: "/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
// 		KeyPath:      "/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
// 	}

// 	signer, err := signerLib.NewSigner(signerConfig)

// 	chaincodeLang := "GOLANG"
// 	chaincodeName := "mycc"

// 	if err != nil {
// 		fmt.Println("[ERROR] NewSigner:", err)
// 		responseChannel <- "Timeout" // fix me
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
// 		responseChannel <- "Timeout" // fix me
// 		return
// 	}

// 	// extract the transient field if it exists
// 	var tMap map[string][]byte

// 	cID := "vnpay-channel"
// 	txID := ""

// 	prop, txid, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(pcommon.HeaderType_ENDORSER_TRANSACTION, cID, invocation, creator, txID, tMap)
// 	if err != nil {
// 		fmt.Println("[ERROR]: CreateChaincodeProposalWithTxIDAndTransient", err)
// 		responseChannel <- "Timeout" // fix me
// 		return
// 	}
// 	fmt.Println("TXID:", txid)

// 	signedProp, err := protoutil.GetSignedProposal(prop, signer)
// 	if err != nil {
// 		fmt.Println("[ERROR]: GetSignedProposal", err)
// 		responseChannel <- "Timeout" // fix me
// 		return
// 	}

// 	cc, _ := grpc.Dial("peer0.org1.example.com:7051", grpc.WithInsecure())

// 	endorser := pb.NewEndorserClient(cc)
// 	mockClients := []pb.EndorserClient{endorser}

// 	// response payload
// 	responses, err := processProposals(mockClients, signedProp)
// 	if err != nil || len(responses) < 1 {
// 		fmt.Println("[ERROR]: processProposals", err)
// 		responseChannel <- "Timeout" // fix me
// 		return
// 	}

// 	rawProposal := RawProposal(prop)
// 	proposalResponse := ProposalResponse(responses[0])
// 	response := ProposalWrapper{Prop: rawProposal, Response: proposalResponse}

// 	// save byte slice into Redis
// 	responseByte, err := json.Marshal(response)

// 	if err != nil {
// 		fmt.Println("error:", err.Error())
// 		responseChannel <- "Timeout" // fix me
// 		return
// 	}

// 	rdb.RPush(ctx, "pending-proposals", responseByte)
// 	responseChannel <- string(responses[0].Response.Payload)
// 	fmt.Println("return deeeee")
// 	return
// }

// // processProposals sends a signed proposal to a set of peers, and gathers all the responses.
// func processProposals(endorserClients []pb.EndorserClient, signedProposal *pb.SignedProposal) ([]*pb.ProposalResponse, error) {
// 	responsesCh := make(chan *pb.ProposalResponse, len(endorserClients))
// 	errorCh := make(chan error, len(endorserClients))
// 	wg := sync.WaitGroup{}
// 	for _, endorser := range endorserClients {
// 		wg.Add(1)
// 		go func(endorser pb.EndorserClient) {
// 			defer wg.Done()
// 			proposalResp, err := endorser.ProcessProposal(context.Background(), signedProposal)
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

// func send(w http.ResponseWriter, r *http.Request) {
// 	// send the first signed proposal response
// 	var proposalWrapper ProposalWrapper
// 	redisData, err := rdb.LPop(ctx, "pending-proposals").Bytes()

// 	if err != nil {
// 		http.Error(w, "No more pending proposal response!", 500)
// 		return
// 	}

// 	json.Unmarshal(redisData, &proposalWrapper)
// 	rawProposal := proposalWrapper.Prop
// 	proposalResponse := proposalWrapper.Response

// 	if &proposalResponse != nil {
// 		if proposalResponse.Response.Status >= shim.ERRORTHRESHOLD {
// 			http.Error(w, "shim.ERRORTHRESHOLD", 500)
// 			return
// 		}

// 		signerConfig := signerLib.Config{
// 			MSPID:        "Org1MSP",
// 			IdentityPath: "/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
// 			KeyPath:      "/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
// 		}

// 		signer, err := signerLib.NewSigner(signerConfig)
// 		if err != nil {
// 			fmt.Println("ERROR NewSigner", err)
// 			http.Error(w, "ERROR NewSigner", 500)
// 			return
// 		}

// 		// assemble a signed transaction (it's an Envelope message)
// 		env, err := protoutil.CreateSignedTx(rawProposal, signer, proposalResponse)
// 		if err != nil {
// 			fmt.Println("ERROR CreateSignedTx", err)
// 			http.Error(w, "could not assemble transaction", 500)
// 			return
// 		}

// 		// var dg *DeliverGroup
// 		// var ctx context.Context
// 		// if waitForEvent {
// 		// 	var cancelFunc context.CancelFunc
// 		// 	ctx, cancelFunc = context.WithTimeout(context.Background(), waitForEventTimeout)
// 		// 	defer cancelFunc()

// 		// 	dg = NewDeliverGroup(
// 		// 		deliverClients,
// 		// 		peerAddresses,
// 		// 		signer,
// 		// 		certificate,
// 		// 		channelID,
// 		// 		txid,
// 		// 	)
// 		// 	// connect to deliver service on all peers
// 		// 	err := dg.Connect(ctx)
// 		// 	if err != nil {
// 		// 		return nil, err
// 		// 	}
// 		// }

// 		address := "orderer.example.com:7050"
// 		// override := ""
// 		clientConfig := comm.ClientConfig{}
// 		clientConfig.Timeout = 3 * time.Second
// 		secOpts := comm.SecureOptions{}
// 		clientConfig.SecOpts = secOpts

// 		if err != nil {
// 			fmt.Println("[ERROR] failed to load config for OrdererClient")
// 			return
// 		}

// 		gClient, err := comm.NewGRPCClient(clientConfig)

// 		oc := &common.OrdererClient{
// 			CommonClient: common.CommonClient{
// 				GRPCClient: gClient,
// 				Address:    address,
// 				// sn:         override
// 			}}

// 		bc, err := oc.Broadcast()

// 		broadcastClient := &common.BroadcastGRPCClient{Client: bc}
// 		// send the envelope for ordering
// 		if err = broadcastClient.Send(env); err != nil {
// 			http.Error(w, "error sending transaction for update function", 500)
// 			return
// 		}

// 		// if dg != nil && ctx != nil {
// 		// 	// wait for event that contains the txid from all peers
// 		// 	err = dg.Wait(ctx)
// 		// 	if err != nil {
// 		// 		return nil, err
// 		// 	}
// 		// }
// 	}
// 	fmt.Println("OK")
// 	fmt.Fprint(w, "OK")
// 	return
// }

// func count(w http.ResponseWriter, r *http.Request) {
// 	// fmt.Println("proposalResponses has", len(proposalResponses), "elements")
// 	// fmt.Fprint(w, len(proposalResponses))
// }
