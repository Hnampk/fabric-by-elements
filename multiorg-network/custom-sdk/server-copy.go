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

	"multiorg-network/custom-sdk/fabric/usable-inter-nal/peer/chaincode"
	"multiorg-network/custom-sdk/fabric/usable-inter-nal/peer/common"
	"multiorg-network/custom-sdk/fabric/usable-inter-nal/pkg/comm"
	"multiorg-network/custom-sdk/fabric/usable-inter-nal/pkg/identity"

	redis "github.com/go-redis/redis/v8"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pcommon "github.com/hyperledger/fabric-protos-go/common"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	signerLib "github.com/hyperledger/fabric/cmd/common/signer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	cmap "github.com/orcaman/concurrent-map"
)

type Proposer struct {
	Id                 int
	EndorserClients    []pb.EndorserClient
	clientConn         *grpc.ClientConn
	requestChannel     chan ProposalWrapper
	tartgetPeerAddress string // merge this
	peerMSPID          string // merge this
	org                string // merge this
	channelID          string
}

type ProposalWrapper struct {
	RequestArgs       [][]byte
	responseChannelId string
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
var peerAddress = ""

const ordererAddress = "orderer1.org1.example.com:7050"

// Define Status codes for the response
const (
	OK                    = 200
	ERROR                 = 500
	chaincodeLang         = "GOLANG"
	REDIS_HOST            = "172.16.79.8"
	REDIS_PORT            = "6379"
	RDB_TXSTATUS_WAITING  = "0"
	RDB_TXSTATUS_SUCCESS  = "1"
	RDB_TXSTATUS_FAILED   = "2"
	REDIS_API_PUBSUB_CHAN = "api-chaincode-channel"
)

var (
	rdb            *redis.Client
	channelID      = "vnpay-channel"
	rootURL        = "/home/ewallet/network/"
	chaincodeName  = "mycc1"
	port           = "8090"
	waitForEvent   = false
	workerNum      = 10
	invokeChannel  = make(chan ProposalWrapper)
	queryChannel   = make(chan ProposalWrapper)
	submitChannel  = make(chan SignedProposalWrapper)
	deliverClients = []pb.DeliverClient{}
)

/*
	mapping txid and a channel - which help the invokeHandler know that the transaction is submitted
*/
var responseChannelMap = cmap.New()
var proposalResponseChannelMap = cmap.New()

type ProposalResponseChannelID struct {
	id int

	mutex sync.Mutex
}

func (c *ProposalResponseChannelID) Inc() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.id++

	return c.id
}

var proposalResponseChannelID = ProposalResponseChannelID{}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter {0: nampkh, 1: server test}")
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Please enter port (ex: 8090)")
		return
	}

	if len(os.Args) < 4 {
		fmt.Println("Please enter channel name")
		return
	}

	if len(os.Args) < 5 {
		fmt.Println("Please enter chaincode name")
		return
	}

	if len(os.Args) < 6 {
		fmt.Println("Please enter the Number of connections")
		return
	}

	if os.Args[1] == "0" {
		rootURL = "/home/nampkh/nampkh/my-fabric/multiorg-network/"
	}

	port = os.Args[2]
	channelID = os.Args[3]
	chaincodeName = os.Args[4]

	var err error
	workerNum, err = strconv.Atoi(os.Args[5])
	if err != nil {
		fmt.Println("An error occurred: ", err)
		return
	}

	initProposerPool(workerNum)
	err = initSubmitterPool(workerNum)
	initListenerPool(workerNum)

	if err != nil {
		fmt.Println("An error occurred while initSubmitterPool: ", err)
		return
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/invoke", invokeHandler)
	mux.HandleFunc("/invoke-only", invokeOnlyHandler)
	mux.HandleFunc("/query", queryHandler)
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/length", lengthHandler)
	// listen and serve
	fmt.Println("Server listen on port", port)
	http.ListenAndServe(":"+port, mux)
}

func initProposerPool(poolSize int) {
	for i := 0; i < poolSize; i++ {
		var peerMSPID string
		var org string
		var channel string

		var tlsRootCertFile string
		var clientKeyFile string
		var clientCertFile string

		if i%poolSize == 0 {
			peerAddress = "peer0.org1.example.com:7051"
			peerMSPID = "Org1MSP"
			org = "org1.example.com"
			channel = "vnpay-channel"
			tlsRootCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			clientKeyFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key"
			clientCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt"
		} else if i%poolSize == 1 {
			peerAddress = "peer1.org1.example.com:7051"
			peerMSPID = "Org1MSP"
			org = "org1.example.com"
			channel = "vnpay-channel"
			tlsRootCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt"
			clientKeyFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/server.key"
			clientCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/server.crt"
		} else if i%poolSize == 2 {
			peerAddress = "peer2.org1.example.com:7051"
			peerMSPID = "Org1MSP"
			org = "org1.example.com"
			channel = "vnpay-channel"
			tlsRootCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer2.org1.example.com/tls/ca.crt"
			clientKeyFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer2.org1.example.com/tls/server.key"
			clientCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer2.org1.example.com/tls/server.crt"
		} else if i%poolSize == 3 {
			peerAddress = "peer3.org1.example.com:7051"
			peerMSPID = "Org1MSP"
			org = "org1.example.com"
			channel = "vnpay-channel"
			tlsRootCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer3.org1.example.com/tls/ca.crt"
			clientKeyFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer3.org1.example.com/tls/server.key"
			clientCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer3.org1.example.com/tls/server.crt"
		} else {
			peerAddress = "peer4.org1.example.com:7051"
			peerMSPID = "Org1MSP"
			org = "org1.example.com"
			channel = "vnpay-channel"
			tlsRootCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer4.org1.example.com/tls/ca.crt"
			clientKeyFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer4.org1.example.com/tls/server.key"
			clientCertFile = "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer4.org1.example.com/tls/server.crt"
		}

		proposer, err := initProposer(i, peerAddress, peerMSPID, org, channel, tlsRootCertFile, clientKeyFile, clientCertFile)

		if err != nil {
			fmt.Println("[ERROR]initProposerPool:", err)
			continue
		}

		proposer.requestChannel = invokeChannel

		go proposer.start()
	}
}

func initProposer(id int, tartgetPeerAddress string, peerMSPID string, org string, channel string, tlsRootCertFile string, clientKeyFile string, clientCertFile string) (*Proposer, error) {
	// cc, err := grpc.Dial(tartgetPeerAddress, grpc.WithInsecure()) // Without TLS, for test only
	// cc, err := grpc.Dial(tartgetPeerAddress)

	// if err != nil {
	// 	fmt.Println("[ERROR]initProposer: Dial:", err)
	// 	return nil, err
	// }

	// endorser := pb.NewEndorserClient(cc)

	clientConfig := comm.ClientConfig{}
	clientConfig.Timeout = 3 * time.Second

	// if clientConfig.Timeout == time.Duration(0) {
	// 	clientConfig.Timeout = defaultConnTimeout
	// }

	secOpts := comm.SecureOptions{
		UseTLS:            true,
		RequireClientCert: true,
	}

	if secOpts.RequireClientCert {
		keyPEM, err := ioutil.ReadFile(clientKeyFile)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to load peer.tls.clientKey.file")
		}
		secOpts.Key = keyPEM
		certPEM, err := ioutil.ReadFile(clientCertFile)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to load peer.tls.clientCert.file")
		}
		secOpts.Certificate = certPEM
	}
	clientConfig.SecOpts = secOpts

	if clientConfig.SecOpts.UseTLS {
		if tlsRootCertFile == "" {
			return nil, errors.New("tls root cert file must be set")
		}
		caPEM, res := ioutil.ReadFile(tlsRootCertFile)
		if res != nil {
			return nil, errors.WithMessagef(res, "unable to load TLS root cert file from %s", tlsRootCertFile)
		}
		clientConfig.SecOpts.ServerRootCAs = [][]byte{caPEM}
	}

	pc, err := common.NewPeerClientForClientConfig(tartgetPeerAddress, "", clientConfig)
	if err != nil {
		fmt.Println("failed to create NewPeerClientForAddress: ", tartgetPeerAddress)
		return nil, err
	}

	e, err := pc.Endorser()
	if err != nil {
		fmt.Println("failed to get Endorser: ", tartgetPeerAddress)
		return nil, err
	}
	endorserClients := []pb.EndorserClient{e}

	fmt.Println("[Custom-SDK] Proposer started:", id)

	deliverClient, err := pc.PeerDeliver()

	if err != nil {
		fmt.Println("failed to get PeerDeliver: ", tartgetPeerAddress)
		return nil, err
	}

	if len(deliverClients) == 0 {
		deliverClients = append(deliverClients, deliverClient)
	}

	return &Proposer{
		Id:                 id,
		EndorserClients:    endorserClients,
		clientConn:         nil,
		tartgetPeerAddress: tartgetPeerAddress,
		peerMSPID:          peerMSPID,
		org:                org,
		channelID:          channel,
	}, nil
}

func (p *Proposer) closeConnection() {
	p.clientConn.Close()
}

func (p *Proposer) start() {
	// defer p.closeConnection()

	for wrapper := range p.requestChannel {
		// when receive a proposal, forward it to peer (propose)

		go func(wrapper ProposalWrapper) {
			response, err := p.propose(wrapper.RequestArgs)

			if tmp, ok := proposalResponseChannelMap.Get(wrapper.responseChannelId); ok {
				proposalResponseChannel := tmp.(chan ProposalResponse)

				if err != nil {
					proposalResponseChannel <- ProposalResponse{Error: err}
					return
				}

				proposalResponseChannel <- *response
			} else {
				fmt.Println("Channel", wrapper.responseChannelId, "does not exist!")
			}

		}(wrapper)

	}
}

func (p *Proposer) propose(args [][]byte) (*ProposalResponse, error) {

	var signerConfig = signerLib.Config{
		MSPID:        p.peerMSPID,
		IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/" + p.org + "/users/Admin@" + p.org + "/msp/signcerts/Admin@" + p.org + "-cert.pem",
		KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/" + p.org + "/users/Admin@" + p.org + "/msp/keystore/priv_sk",
	}

	signer, err := signerLib.NewSigner(signerConfig)

	if err != nil {
		// fmt.Println("[ERROR]propose: NewSigner:", err)
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
		// fmt.Println("[ERROR] propose: Serialize:", err)
		return nil, err
	}

	// extract the transient field if it exists
	var tMap map[string][]byte

	// Nampkh: must feed empty txID
	txID := ""

	prop, txID, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(pcommon.HeaderType_ENDORSER_TRANSACTION, p.channelID, invocation, creator, txID, tMap)
	if err != nil {
		// fmt.Println("[ERROR] propose: CreateChaincodeProposalWithTxIDAndTransient:", err)
		return nil, err
	}

	signedProp, err := protoutil.GetSignedProposal(prop, signer)
	if err != nil {
		// fmt.Println("[ERROR] propose: GetSignedProposal:", err)
		return nil, err
	}

	responses, err := processProposals(p.EndorserClients, *signedProp)
	if err != nil || len(responses) < 1 {
		// fmt.Println("[ERROR] propose: processProposals:", err)
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
	for i := 0; i < poolSize; i++ {
		submitter := Submitter{
			Id: i,
		}

		var ordererAddress string
		var tlsRootCertFile string
		if i%poolSize == 0 {
			ordererAddress = "orderer1.org1.example.com:7050"
			tlsRootCertFile = "../orderer/crypto-config/ordererOrganizations/org1.example.com/orderers/orderer1.org1.example.com/tls/ca.crt"
		} else {
			ordererAddress = "orderer2.org1.example.com:7050"
			tlsRootCertFile = "../orderer/crypto-config/ordererOrganizations/org1.example.com/orderers/orderer2.org1.example.com/tls/ca.crt"
		}

		err := submitter.connectToOrderer(ordererAddress, tlsRootCertFile)

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

func (s *Submitter) connectToOrderer(tartgetOrdererAddress string, tlsRootCertFile string) error {

	clientKeyFile := "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key"
	clientCertFile := "../peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt"

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

	secOpts := comm.SecureOptions{
		UseTLS:            true,
		RequireClientCert: true,
	}

	if secOpts.UseTLS {
		caPEM, res := ioutil.ReadFile(tlsRootCertFile)
		if res != nil {
			err := errors.WithMessage(res,
				fmt.Sprintf("unable to load orderer.tls.rootcert.file"))
			return err
		}
		secOpts.ServerRootCAs = [][]byte{caPEM}
	}
	if secOpts.RequireClientCert {
		keyPEM, res := ioutil.ReadFile(clientKeyFile)
		if res != nil {
			err := errors.WithMessage(res,
				fmt.Sprintf("unable to load orderer.tls.clientKey.file"))
			return err
		}
		secOpts.Key = keyPEM
		certPEM, res := ioutil.ReadFile(clientCertFile)
		if res != nil {
			err := errors.WithMessage(res,
				fmt.Sprintf("unable to load orderer.tls.clientCert.file"))
			return err
		}
		secOpts.Certificate = certPEM
	}
	clientConfig.SecOpts = secOpts
	clientConfig.KaOpts = KaOpts

	gClient, err := comm.NewGRPCClient(clientConfig)

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
			// if err != nil {
			// 	// retry to connect to orderer 1 time! (the connection may be be disrupted)
			// 	err = s.connectToOrderer(ordererAddress)
			// 	if err != nil {
			// 		wrapper.errorChannel <- err
			// 		return
			// 	}

			// 	err = s.submit(env)
			// 	// if still got error after retry => orderer problem
			// 	if err != nil {
			// 		wrapper.errorChannel <- err
			// 		return
			// 	}
			// }
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
		var peerAddress string
		var channelID string
		var signerConfig signerLib.Config

		if i%poolSize == 0 {
			peerAddress = "peer0.org1.example.com"
			channelID = "vnpay-channel"
			signerConfig = signerLib.Config{
				MSPID:        "Org1MSP",
				IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
				KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
			}

		} else if i%poolSize == 1 {
			peerAddress = "peer1.org1.example.com"
			channelID = "vnpay-channel"
			signerConfig = signerLib.Config{
				MSPID:        "Org1MSP",
				IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
				KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
			}
		} else if i%poolSize == 2 {
			peerAddress = "peer2.org1.example.com"
			channelID = "vnpay-channel"
			signerConfig = signerLib.Config{
				MSPID:        "Org1MSP",
				IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
				KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
			}
		} else if i%poolSize == 3 {
			peerAddress = "peer3.org1.example.com"
			channelID = "vnpay-channel"
			signerConfig = signerLib.Config{
				MSPID:        "Org1MSP",
				IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
				KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
			}
		} else {
			peerAddress = "peer4.org1.example.com"
			channelID = "vnpay-channel"
			signerConfig = signerLib.Config{
				MSPID:        "Org1MSP",
				IdentityPath: rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem",
				KeyPath:      rootURL + "peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk",
			}
		}

		listener, err := initSubmissionListener(i, peerAddress, channelID, signerConfig)

		if err != nil {

		}

		go listener.start()
	}
}

func initSubmissionListener(id int, peerAddress string, channelID string, signerConfig signerLib.Config) (*SubmissionListener, error) {
	ctx := context.Background()

	dg, err := createDeliverGroup(ctx, "", peerAddress, channelID, signerConfig)

	if err != nil {
		return nil, err
	}

	return &SubmissionListener{
		Id:  id,
		Dg:  dg,
		ctx: ctx,
	}, nil
}

func createDeliverGroup(ctx context.Context, txID string, peerAddress string, channelID string, signerConfig signerLib.Config) (*chaincode.DeliverGroup, error) {
	signer, err := signerLib.NewSigner(signerConfig)

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
		// []string{"peer0.org1.example.com", "peer0.org2.example.com", "peer0.org3.example.com", "peer0.org4.example.com", "peer0.org5.example.com"},
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

type SuccessMessage struct {
	Message string
	Nonces  []string
}

func invokeHandler(res http.ResponseWriter, req *http.Request) {
	// =========================== PROPOSE PHASE ===========================
	proposeRequest, err := resolveHttpRequest(req)

	if err != nil {
		http.Error(res, "can't read body", http.StatusBadRequest)
		return
	}

	args := packArgs(proposeRequest, false)

	responseChan := make(chan ProposalResponse)
	id := strconv.Itoa(proposalResponseChannelID.Inc())
	proposalResponseChannelMap.Set(id, responseChan)
	defer proposalResponseChannelMap.Remove(id)

	proposalWrapper := ProposalWrapper{
		RequestArgs:       args,
		responseChannelId: id,
	}
	var proposalResponse ProposalResponse

	// send proposal to proposer
	invokeChannel <- proposalWrapper
	proposalResponse = <-responseChan

	// fmt.Println("proposalResponse", proposalResponse)
	if proposalResponse.Error != nil {
		http.Error(res, proposalResponse.Error.Error(), http.StatusInternalServerError)
		return
	}

	proposalResp := proposalResponse.Content[0]
	if proposalResp.Response.Status >= shim.ERRORTHRESHOLD {
		http.Error(res, proposalResp.Response.Message, http.StatusInternalServerError)
		return
	}
	// var response SuccessMessage

	// err = json.Unmarshal(proposalResponse.Content[0].Response.Payload, &response)
	// if err != nil {
	// 	fmt.Println("failed to unmarshal payload")
	// 	// return
	// }

	//
	// =========================== SUBMIT PHASE ===========================

	_ = sendToSubmitter(proposalResponse)

	// var ctx = context.Background()

	// if submitErr != nil {
	// 	http.Error(res, submitErr.Error(), http.StatusInternalServerError)
	// 	// Publish a message.
	// 	// payload, _ := json.Marshal(SuccessMessage{Nonce: response.Nonce, Message: RDB_TXSTATUS_FAILED})
	// 	go rdb.Publish(ctx, REDIS_API_PUBSUB_CHAN, SuccessMessage{Nonces: response.Nonces, Message: RDB_TXSTATUS_FAILED})

	// 	fmt.Println("submit Error ", submitErr)
	// 	return
	// }

	// payload, _ := json.Marshal(SuccessMessage{Nonces: response.Nonces, Message: RDB_TXSTATUS_SUCCESS})
	// go rdb.Publish(ctx, REDIS_API_PUBSUB_CHAN, payload)

	fmt.Fprint(res, "submitted, txid: "+proposalResponse.TxID)
	return
}

type Mylist struct {
	proposalResponses []ProposalResponse

	mux sync.Mutex
}

func (m *Mylist) push(proposalResponse ProposalResponse) {
	// m.mux.Lock()
	// defer m.mux.Unlock()

	m.proposalResponses = append(m.proposalResponses, proposalResponse)
}

func (m *Mylist) get() ProposalResponse {
	m.mux.Lock()
	defer m.mux.Unlock()

	proposalResponse := m.proposalResponses[0]
	m.proposalResponses = m.proposalResponses[1:]

	return proposalResponse
}

var proposalResponses = Mylist{}

func invokeOnlyHandler(res http.ResponseWriter, req *http.Request) {
	// =========================== PROPOSE PHASE ===========================
	proposeRequest, err := resolveHttpRequest(req)

	if err != nil {
		http.Error(res, "can't read body", http.StatusBadRequest)
		return
	}

	fmt.Println(proposeRequest)

	args := packArgs(proposeRequest, false)

	var proposalResponse ProposalResponse
	responseChan := make(chan ProposalResponse)
	id := strconv.Itoa(proposalResponseChannelID.Inc())
	proposalResponseChannelMap.Set(id, responseChan)
	defer proposalResponseChannelMap.Remove(id)

	proposalWrapper := ProposalWrapper{
		RequestArgs:       args,
		responseChannelId: id,
	}

	// send proposal to proposer
	invokeChannel <- proposalWrapper
	proposalResponse = <-responseChan
	if proposalResponse.Error != nil {
		http.Error(res, proposalResponse.Error.Error(), http.StatusInternalServerError)
		return
	}

	proposalResp := proposalResponse.Content[0]
	if proposalResp.Response.Status >= shim.ERRORTHRESHOLD {
		http.Error(res, proposalResp.Response.Message, http.StatusInternalServerError)
		return
	}

	// save signed proposal
	// proposalResponses.push(proposalResponse)

	// fmt.Println(string(proposalResponse.Content[0].Response.Payload))
	fmt.Fprint(res, "proposed, txid: "+proposalResponse.TxID)
	return
}

func submitHandler(res http.ResponseWriter, req *http.Request) {
	proposalResponse := proposalResponses.get()

	submitErr := sendToSubmitter(proposalResponse)

	if submitErr != nil {
		http.Error(res, submitErr.Error(), http.StatusInternalServerError)
		fmt.Println("submit Error ", submitErr)
		return
	}

	// fmt.Println(string(proposalResponse.Content[0].Response.Payload))
	fmt.Fprint(res, "submitted, txid: "+proposalResponse.TxID)
	return
}

func lengthHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, len(proposalResponses.proposalResponses))
	return
}

func queryHandler(res http.ResponseWriter, req *http.Request) {
	// =========================== PROPOSE PHASE ===========================
	proposeRequest, err := resolveHttpRequest(req)

	if err != nil {
		http.Error(res, "can't read body", http.StatusBadRequest)
		return
	}

	fmt.Println(proposeRequest)

	args := packArgs(proposeRequest, false)

	var proposalResponse ProposalResponse
	responseChan := make(chan ProposalResponse)
	id := strconv.Itoa(proposalResponseChannelID.Inc())
	proposalResponseChannelMap.Set(id, responseChan)
	defer proposalResponseChannelMap.Remove(id)

	proposalWrapper := ProposalWrapper{
		RequestArgs:       args,
		responseChannelId: id,
	}

	// send proposal to proposer
	invokeChannel <- proposalWrapper // may change to query channel, if needed!
	proposalResponse = <-responseChan
	if proposalResponse.Error != nil {
		http.Error(res, proposalResponse.Error.Error(), http.StatusInternalServerError)
		return
	}

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
