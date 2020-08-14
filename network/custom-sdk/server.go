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
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pcommon "github.com/hyperledger/fabric-protos-go/common"

	// "github.com/hyperledger/fabric-protos-go/peer"
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
	RequestArgs             [][]byte
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
	Prop    *pb.Proposal
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

type EventResponse struct {
	txID    string
	status  bool
	errCode pb.TxValidationCode
}

// hard-code for testing only
const peerAddress = "peer0.org1.example.com:7051"
const ordererAddress = "orderer.example.com:7050"
const peerMSPID = "Org1MSP"
const chaincodeLang = "GOLANG"
const channelID = "vnpay-channel"

var rootURL = "/home/ewallet/network/"
var signerConfig signerLib.Config
var chaincodeName = "mycc1"

var waitForEvent = false

var deliverPeerAddress = []string{"peer0.org1.example.com:7051"}

var workerNum = 10
var invokeChannel = make(chan ProposalWrapper)
var queryChannel = make(chan ProposalWrapper)
var submitChannel = make(chan SignedProposalWrapper)
var deliverClients = []pb.DeliverClient{}

/*
	mapping txid and a channel - which help the invokeHandler know that the transaction is submitted
*/
var responseChannelMap = cmap.New()

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter the Number of connections")
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Please enter chaincode name")
		return
	}

	if len(os.Args) < 4 {
		fmt.Println("Please enter {0: nampkh, 1: server test}")
		return
	}

	var err error
	workerNum, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("An error occurred: ", err)
		return
	}

	chaincodeName = os.Args[2]

	if os.Args[3] == "0" {
		rootURL = "/home/nampkh/nampkh/my-fabric/network/"
	}

	signerConfig = signerLib.Config{
		MSPID:        peerMSPID,
		IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
		KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
	}

	initProposerPool(workerNum)
	err = initSubmitterPool(workerNum)
	initListenerPool(1)

	// nonce = Nonce{value: make(map[string]int)}

	if err != nil {
		fmt.Println("An error occurred while initSubmitterPool: ", err)
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/invoke", invokeHandler)
	mux.HandleFunc("/query", queryHandler)
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

		proposer.requestChannel = invokeChannel

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

func (p *Proposer) closeConnection() {
	p.clientConn.Close()
}

func (p *Proposer) start() {
	defer p.closeConnection()

	for wrapper := range p.requestChannel {
		// when receive a proposal, forward it to peer (propose)

		response, err := p.propose(wrapper.RequestArgs)

		if err != nil {
			wrapper.proposalResponseChannel <- ProposalResponse{Error: err}
			continue
		}

		wrapper.proposalResponseChannel <- *response
	}
}

func (p *Proposer) propose(args [][]byte) (*ProposalResponse, error) {
	signer, err := signerLib.NewSigner(signerConfig)

	if err != nil {
		fmt.Println("[ERROR]propose: NewSigner:", err)
		return nil, err
	}

	invokeInput := pb.ChaincodeInput{
		IsInit: false,
		Args:   args,
	}

	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
		ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input:       &invokeInput,
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

func sendToSubmitter(proposalResponse ProposalResponse) error {

	responseChannel := make(chan EventResponse)
	responseChannelMap.Set(proposalResponse.TxID, responseChannel)
	defer responseChannelMap.Remove(proposalResponse.TxID)

	errChan := make(chan error)

	submitChannel <- SignedProposalWrapper{
		proposalResponse: proposalResponse,
		errorChannel:     errChan,
	}

	if err := <-errChan; err != nil {
		return err
	}

	select {
	case <-time.After(time.Second * 30):
		return errors.Errorf("submit timeout!")
	case submissionEvent := <-responseChannel:
		if !submissionEvent.status {
			// should add revert logic here to update Chaincode internal world state

			return errors.Errorf("submit failed!, TxValidationCode: %s", submissionEvent.errCode)
		}
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

func (s *Submitter) start() {
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

func (s *Submitter) submit(env *pcommon.Envelope) error {
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

func (l *SubmissionListener) start() error {
	if l.Dg != nil && l.ctx != nil {
		// wait for event that contains the txid from all peers
		submissionChannel := make(chan *pb.FilteredTransaction)

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

func onReceiveSubmissionEvent(tx *pb.FilteredTransaction) {
	if tmp, ok := responseChannelMap.Get(tx.Txid); ok {
		responseChannel := tmp.(chan EventResponse)
		status := true

		if tx.TxValidationCode != pb.TxValidationCode_VALID {
			status = false
		}

		responseChannel <- EventResponse{
			txID:    tx.Txid,
			status:  status,
			errCode: tx.TxValidationCode,
		}

		return
	}
}

func invokeHandler(res http.ResponseWriter, req *http.Request) {
	// =========================== PROPOSE PHASE ===========================
	proposeRequest, err := resolveHttpRequest(req)

	if err != nil {
		http.Error(res, "can't read body", http.StatusBadRequest)
		return
	}

	responseChan := make(chan ProposalResponse)
	fmt.Println(proposeRequest)

	args := packArgs(proposeRequest, false)

	proposalWrapper := ProposalWrapper{
		RequestArgs:             args,
		proposalResponseChannel: responseChan,
	}

	var proposalResponse ProposalResponse

	// send proposal to proposer
	invokeChannel <- proposalWrapper
	proposalResponse = <-responseChan
	proposalResp := proposalResponse.Content[0]

	if proposalResp.Response.Status >= shim.ERRORTHRESHOLD {
		http.Error(res, proposalResp.Response.Message, http.StatusInternalServerError)
		return
	}

	if proposalResponse.Error != nil {
		http.Error(res, proposalResponse.Error.Error(), http.StatusInternalServerError)
		return
	}

	//
	// =========================== SUBMIT PHASE ===========================

	// submitErr := sendToSubmitter(proposalResponse)

	// if submitErr != nil {
	// 	http.Error(res, submitErr.Error(), http.StatusInternalServerError)
	// 	fmt.Println("submit Error ", submitErr)
	// 	return
	// }

	// fmt.Println(string(proposalResponse.Content[0].Response.Payload))
	fmt.Fprint(res, "submitted, txid: "+proposalResponse.TxID)
	return
}

func queryHandler(res http.ResponseWriter, req *http.Request) {
	// =========================== PROPOSE PHASE ===========================
	proposeRequest, err := resolveHttpRequest(req)

	if err != nil {
		http.Error(res, "can't read body", http.StatusBadRequest)
		return
	}

	responseChan := make(chan ProposalResponse)
	fmt.Println(proposeRequest)

	args := packArgs(proposeRequest, false)

	proposalWrapper := ProposalWrapper{
		RequestArgs:             args,
		proposalResponseChannel: responseChan,
	}

	var proposalResponse ProposalResponse

	// send proposal to proposer
	invokeChannel <- proposalWrapper // may change to query channel, if needed!
	proposalResponse = <-responseChan

	proposalResp := proposalResponse.Content[0]

	if proposalResp.Response.Status >= shim.ERRORTHRESHOLD {
		http.Error(res, proposalResp.Response.Message, http.StatusInternalServerError)
		return
	}

	fmt.Fprint(res, string(proposalResp.Response.Payload))
	return
}

func resolveHttpRequest(req *http.Request) (*ProposeRequest, error) {
	var proposeRequest ProposeRequest

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, err
	}

	json.Unmarshal(body, &proposeRequest)

	return &proposeRequest, nil
}

func packArgs(proposeRequest *ProposeRequest, isWithNonce bool) [][]byte {
	fcn := proposeRequest.FuncName
	args := [][]byte{[]byte(fcn)}

	for _, element := range proposeRequest.Args {
		args = append(args, []byte(element))
	}

	if isWithNonce {
		// only invoke-request need nonce
		args = append(args, []byte(strconv.Itoa(proposeRequest.Nonce)))
	}

	return args
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
