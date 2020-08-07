package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"example.com/custom-sdk/fabric/usable-inter-nal/peer/chaincode"
	"example.com/custom-sdk/fabric/usable-inter-nal/peer/common"
	"example.com/custom-sdk/fabric/usable-inter-nal/pkg/comm"
	"example.com/custom-sdk/fabric/usable-inter-nal/pkg/identity"
	"example.com/custom-sdk/nonce"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pcommon "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	signerLib "github.com/hyperledger/fabric/cmd/common/signer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	cmap "github.com/orcaman/concurrent-map"
)

type Proposer struct {
	Id              int
	EndorserClients []pb.EndorserClient
	clientConn      *grpc.ClientConn
	requestChannel  chan ProposalWrapper
}

type ProposalWrapper struct {
	Request                 ProposeRequest
	proposalResponseChannel chan ProposalResponse
}

type ProposeRequest struct {
	FuncName string
	Args     []string
	Nonce    int
}

type ProposalResponse struct {
	Error   error
	TxID    string
	Prop    *peer.Proposal
	Content []*pb.ProposalResponse
	Signer  *signerLib.Signer
}

type Submitter struct {
	Id              int
	broadcastClient *common.BroadcastGRPCClient
	submitChannel   chan SignedProposalWrapper
}

type SignedProposalWrapper struct {
	proposalResponse ProposalResponse
	errorChannel     chan error
}

type SubmissionListener struct {
	Id  int
	Dg  *chaincode.DeliverGroup
	ctx context.Context
}

type TransactionWithStatus struct {
	TxID            string
	IsSubmitted     bool
	ResponseChannel chan bool
}

var concurrency = false

type Nonce struct {
	value map[string]int
	mux   sync.Mutex
}

func (n *Nonce) Lock() {
	n.mux.Lock()
}

func (n *Nonce) Unlock() {
	n.mux.Unlock()
}

// Inc increments the counter for the given key.
func (n *Nonce) Inc(key string) int {
	if concurrency {
		n.mux.Lock()
		defer n.mux.Unlock()
	}
	n.value[key]++
	// Lock so only one goroutine at a time can access the map c.v.
	return n.value[key]
}

// hard-code for test only
const peerAddress = "peer0.org1.example.com:7051"
const ordererAddress = "orderer.example.com:7050"
const peerMSPID = "Org1MSP"

// const rootURL = "/home/nampkh/nampkh/my-fabric/network/"

var rootURL = "/home/ewallet/network/"
var signerConfig signerLib.Config

const chaincodeLang = "GOLANG"
const chaincodeName = "mycc"
const channelID = "vnpay-channel"

var waitForEvent = false

var deliverPeerAddress = []string{"peer0.org1.example.com:7051"}

var workerNum = 10
var requestChannel = make(chan ProposalWrapper)
var submitChannel = make(chan SignedProposalWrapper)
var deliverClients = []pb.DeliverClient{}

/*
	mapping txid and a channel - which let the invokeHandler know that the transaction is submitted
*/
var responseChannelMap = cmap.New()

// var nonce Nonce
var noneManager nonce.NonceManager

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter the Number of connections")
		return
	}

	if len(os.Args) >= 3 {
		rootURL = "/home/nampkh/nampkh/my-fabric/network/"
	}

	signerConfig = signerLib.Config{
		MSPID:        peerMSPID,
		IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
		KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
	}

	var err error
	workerNum, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("An error occurred: ", err)
		return
	}

	initProposerPool(workerNum)
	err = initSubmitterPool(workerNum)
	initListenerPool(1)

	noneManager.Init()

	if err != nil {
		fmt.Println("An error occurred while initSubmitterPool: ", err)
		return
	}

	// HTTP server
	mux := http.NewServeMux()

	mux.HandleFunc("/invoke", invokeHandler)
	mux.HandleFunc("/query", query)
	// listen and serve
	port := "8090"
	fmt.Println("Server listen on port", port)
	http.ListenAndServe(":"+port, mux)
}

func initProposerPool(poolSize int) {
	for i := 0; i < poolSize; i++ {
		proposer, err := initProposer(i, peerAddress)

		if err != nil {
			fmt.Println("[ERROR]initProposerPool:", err)
			continue
		}

		proposer.requestChannel = requestChannel

		go proposer.start()
	}
}

func initProposer(id int, tartgetPeerAddress string) (*Proposer, error) {
	cc, err := grpc.Dial(tartgetPeerAddress, grpc.WithInsecure()) // Without TLS, for test only

	if err != nil {
		fmt.Println("[ERROR]initProposer: Dial:", err)
		return nil, err
	}

	endorser := pb.NewEndorserClient(cc)
	endorserClients := []pb.EndorserClient{endorser}

	fmt.Println("[Custom-SDK] Proposer started:", id)

	return &Proposer{
		Id:              id,
		EndorserClients: endorserClients,
		clientConn:      cc,
	}, nil
}

func (p Proposer) closeConnection() {
	p.clientConn.Close()
}

func (p Proposer) start() {
	defer p.closeConnection()

	for wrapper := range p.requestChannel {
		// when receive a proposal, forward it to peer (propose)
		var args [][]byte
		fcn := wrapper.Request.FuncName
		args = [][]byte{[]byte(fcn)}

		for _, element := range wrapper.Request.Args {
			args = append(args, []byte(element))
		}

		response, err := p.propose(args)

		if err != nil {
			wrapper.proposalResponseChannel <- ProposalResponse{Error: err}
			continue
		}

		wrapper.proposalResponseChannel <- *response
	}
}

func (p Proposer) propose(args [][]byte) (*ProposalResponse, error) {
	signer, err := signerLib.NewSigner(signerConfig)

	if err != nil {
		fmt.Println("[ERROR]propose: NewSigner:", err)
		return nil, err
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
		fmt.Println("[ERROR] propose: Serialize:", err)
		return nil, err
	}

	// extract the transient field if it exists
	var tMap map[string][]byte

	// Nampkh: must feed empty txID
	txID := ""

	prop, txID, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(pcommon.HeaderType_ENDORSER_TRANSACTION, channelID, invocation, creator, txID, tMap)
	if err != nil {
		fmt.Println("[ERROR] propose: CreateChaincodeProposalWithTxIDAndTransient:", err)
		return nil, err
	}

	signedProp, err := protoutil.GetSignedProposal(prop, signer)
	if err != nil {
		fmt.Println("[ERROR] propose: GetSignedProposal:", err)
		return nil, err
	}

	responses, err := processProposals(p.EndorserClients, *signedProp)
	if err != nil || len(responses) < 1 {
		fmt.Println("[ERROR] propose: processProposals:", err)
		return nil, err
	}

	return &ProposalResponse{
		Prop:    prop,
		Content: responses,
		TxID:    txID,
		Signer:  signer,
	}, nil
}

// processProposals sends a signed proposal to a set of peers, and gathers all the responses.
func processProposals(endorserClients []pb.EndorserClient, signedProposal pb.SignedProposal) ([]*pb.ProposalResponse, error) {
	responsesCh := make(chan *pb.ProposalResponse, len(endorserClients))
	errorCh := make(chan error, len(endorserClients))
	wg := sync.WaitGroup{}
	for _, endorser := range endorserClients {
		wg.Add(1)
		go func(endorser pb.EndorserClient) {
			defer wg.Done()
			proposalResp, err := endorser.ProcessProposal(context.Background(), &signedProposal)
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

func initSubmitterPool(poolSize int) error {
	// connection to deliver peer will be establish very soon
	err := createPeerDeliverClient()

	if err != nil {
		fmt.Println("[ERROR]initSubmitterPool: createPeerDeliverClient", err)
		return err
	}

	for i := 0; i < poolSize; i++ {
		submitter := Submitter{
			Id: i,
		}
		err := submitter.connectToOrderer(ordererAddress)

		if err != nil {
			fmt.Println("[ERROR]connectToOrderer:", err)
			continue
		}
		submitter.submitChannel = submitChannel

		go submitter.start()
	}

	return nil
}

func createPeerDeliverClient() error {
	for _, deliverClientAddr := range deliverPeerAddress {
		deliverClient, err := common.GetPeerDeliverClientFnc(deliverClientAddr, "tlsRootCertFile")
		if err != nil {
			fmt.Println("[ERROR]createPeerDeliverClient: GetPeerDeliverClientFnc", err)
			return errors.Errorf("[ERROR]createPeerDeliverClient: GetPeerDeliverClientFnc", err)
		}

		deliverClients = append(deliverClients, deliverClient)
	}

	return nil
}

func (s *Submitter) connectToOrderer(tartgetOrdererAddress string) error {

	clientConfig := comm.ClientConfig{}
	clientConfig.Timeout = 30 * time.Second
	clientConfig.AsyncConnect = true

	// Nampkh: if set ClientInterval too low, orderer will disconnect the connection
	// while client ping too frequently
	// see orderer.yaml >> General >> Keepalive >> ServerMinInterval
	KaOpts := comm.KeepaliveOptions{
		// ClientInterval is the duration after which if the client does not see
		// any activity from the server it pings the server to see if it is alive
		ClientInterval: 60 * time.Second,
		// ClientTimeout is the duration the client waits for a response
		// from the server after sending a ping before closing the connection
		// ClientTimeout: 10 * time.Second,
		// // // ServerInterval is the duration after which if the server does not see
		// // // any activity from the client it pings the client to see if it is alive
		// ServerInterval: 5 * time.Second,
		// // // ServerTimeout is the duration the server waits for a response
		// // // from the client after sending a ping before closing the connection
		// ServerTimeout: 10 * time.Second,
		// // // ServerMinInterval is the minimum permitted time between client pings.
		// // // If clients send pings more frequently, the server will disconnect them
		// ServerMinInterval: 5 * time.Second,
	}

	secOpts := comm.SecureOptions{}

	clientConfig.SecOpts = secOpts
	clientConfig.KaOpts = KaOpts

	gClient, err := comm.NewGRPCClient(clientConfig)

	if err != nil {
		fmt.Println("[ERROR] failed to load config for OrdererClient")
		return err
	}

	oc := &common.OrdererClient{
		CommonClient: common.CommonClient{
			GRPCClient: gClient,
			Address:    tartgetOrdererAddress,
			// sn:         override
		}}

	bc, err := oc.Broadcast()
	fmt.Println("[Custom-SDK] Submitter started:", s.Id)
	broadcastClient := &common.BroadcastGRPCClient{Client: bc}

	if err != nil {
		fmt.Println("ERROR NewSigner", err)
		return err
	}

	s.broadcastClient = broadcastClient

	return nil
}

func (s Submitter) start() {
	defer s.broadcastClient.Close()

	for wrapper := range s.submitChannel {
		func(wrapper SignedProposalWrapper) {
			defer close(wrapper.errorChannel)
			env, err := protoutil.CreateSignedTx(wrapper.proposalResponse.Prop, wrapper.proposalResponse.Signer, wrapper.proposalResponse.Content...)
			if err != nil {
				fmt.Println("ERROR CreateSignedTx", err)
				wrapper.errorChannel <- err
				return
			}

			err = s.submit(env)
			if err != nil {
				// retry to connect to orderer 1 time! (the connection may be be disrupted)
				err = s.connectToOrderer(ordererAddress)
				if err != nil {
					wrapper.errorChannel <- err
					return
				}

				err = s.submit(env)
				// if still got error after retry => orderer problem
				if err != nil {
					wrapper.errorChannel <- err
					return
				}
			}

			// wrapper.errorChannel <- errors.Errorf("DONE")
		}(wrapper)
		// free the Submitter after submit tx to orderer
	}
}

func (s Submitter) submit(env *pcommon.Envelope) error {
	// send the envelope for ordering
	if err := s.broadcastClient.Send(env); err != nil {
		fmt.Println("error sending transaction for update function", 500, err)
		return err
	}
	return nil
}

func initListenerPool(poolSize int) {
	for i := 0; i < poolSize; i++ {
		listener, err := initSubmissionListener(i)

		if err != nil {

		}

		go listener.start()
	}
}

func initSubmissionListener(id int) (*SubmissionListener, error) {
	ctx := context.Background()

	dg, err := createDeliverGroup(ctx, "")

	if err != nil {
		return nil, err
	}

	return &SubmissionListener{
		Id:  id,
		Dg:  dg,
		ctx: ctx,
	}, nil
}

func createDeliverGroup(ctx context.Context, txID string) (*chaincode.DeliverGroup, error) {
	signer, err := signerLib.NewSigner(signerConfig)

	deliverClients := []pb.DeliverClient{}
	deliverClient, err := common.GetPeerDeliverClientFnc(deliverPeerAddress[0], "tlsRootCertFile")
	if err != nil {
		fmt.Println("[ERROR]createPeerDeliverClient: GetPeerDeliverClientFnc", err)
		return nil, errors.Errorf("[ERROR]createPeerDeliverClient: GetPeerDeliverClientFnc", err)
	}

	deliverClients = append(deliverClients, deliverClient)

	if err != nil {
		fmt.Println("[ERROR]propose: NewSigner:", err)
		return nil, err
	}

	var dg *chaincode.DeliverGroup

	certificate, err := common.GetCertificateFnc()
	// connect to deliver service on all peers
	if err != nil {
		fmt.Println("error GetCertificateFnc", err)
		return nil, err
	}

	dg = NewDeliverGroup(
		deliverClients,
		[]string{peerAddress},
		signer,
		certificate,
		channelID,
		txID,
	)

	// connect to deliver service on all peers
	err = dg.Connect(ctx)
	if err != nil {
		fmt.Println("error Connect", err)
		return nil, err
	}

	return dg, nil
}

func (l SubmissionListener) start() error {
	if l.Dg != nil && l.ctx != nil {
		// wait for event that contains the txid from all peers
		submissionChannel := make(chan string)

		// Fix me: did not handle exception
		go l.Dg.Listen(l.ctx, submissionChannel)
		// if err != nil {
		// 	fmt.Println("Listen", err.Error())
		// 	return err
		// }

		for msg := range submissionChannel {

			// Fix me: what happened when we received an error?
			onReceiveSubmissionEvent(msg)
		}

	}
	return nil
}

func onReceiveSubmissionEvent(txID string) {
	if tmp, ok := responseChannelMap.Get(txID); ok {
		responseChannel := tmp.(chan bool)
		responseChannel <- true

		return
	}
}

func invokeHandler(res http.ResponseWriter, req *http.Request) {
	// =========================== PROPOSE PHASE ===========================
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(res, "can't read body", http.StatusBadRequest)
		return
	}

	var proposeRequest ProposeRequest
	json.Unmarshal(body, &proposeRequest)
	accountId := proposeRequest.Args[0]

	// Nampkh: the below logic does not handle the order of incoming request
	//	It just make sure that only 01 (request of an account) will be executed at a time
	var increasedNonce int

	increasedNonce, err = tryToGetNonce(accountId)

	if err != nil {
		http.Error(res, "[ERROR] invokeHandler: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if increasedNonce < 0 {
		// busy service
		pubsub := noneManager.SubscribeChannel(accountId)
		pubsubChan := pubsub.Channel()

		for _ = range pubsubChan {
			// msg maybe error
			increasedNonce, err = tryToGetNonce(accountId)
			// fmt.Println(msg.Payload, "mynonce:", increasedNonce)

			if err != nil {
				http.Error(res, "[ERROR] invokeHandler: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if increasedNonce > -1 {
				break
			}
			// still busy
		}
	}

	proposeRequest.Nonce = increasedNonce

	responseChan := make(chan ProposalResponse)
	fmt.Println(proposeRequest)
	proposalWrapper := ProposalWrapper{
		Request:                 proposeRequest,
		proposalResponseChannel: responseChan,
	}

	var proposalResponse ProposalResponse

	// send proposal to proposer
	requestChannel <- proposalWrapper
	proposalResponse = <-responseChan

	err = noneManager.IncreaseLatestNonce(accountId)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	if proposalResponse.Error != nil {
		http.Error(res, proposalResponse.Error.Error(), http.StatusInternalServerError)
		return
	}

	//
	// =========================== SUBMIT PHASE ===========================

	proposalResp := proposalResponse.Content[0]

	if proposalResp.Response.Status >= shim.ERRORTHRESHOLD {
		http.Error(res, "shim.ERRORTHRESHOLD", http.StatusInternalServerError)
		return
	}

	responseChannel := make(chan bool)
	responseChannelMap.Set(proposalResponse.TxID, responseChannel)

	errChan := make(chan error)

	submitChannel <- SignedProposalWrapper{
		proposalResponse: proposalResponse,
		errorChannel:     errChan,
	}

	if err := <-errChan; err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	<-responseChannel
	responseChannelMap.Remove(proposalResponse.TxID)

	fmt.Fprint(res, "submitted, txid: "+proposalResponse.TxID)
	return
}

func tryToGetNonce(accountId string) (int, error) {

	currentNonceInt, err := noneManager.GetLatestNonce(accountId)

	if err != nil {
		if !noneManager.IsNilValue(err) {
			return -1, err
		}

		currentNonceInt = -1
	}

	increasedNonce := currentNonceInt + 1
	isMyTurn, err := noneManager.SetNX(accountId + "-" + strconv.Itoa(increasedNonce))

	if err != nil {
		return -1, err
	}

	if !isMyTurn {
		return -1, nil
	}

	return increasedNonce, nil
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

func query(res http.ResponseWriter, req *http.Request) {

}
