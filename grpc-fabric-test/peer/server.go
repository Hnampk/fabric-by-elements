/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "example.com/grpc-fabric-test/helloworld"
	cmap "github.com/orcaman/concurrent-map"

	"google.golang.org/grpc"
)

const (
	peerServerPort = ":50051"
)

var (
	chaincodeServerHost = "localhost"
	chaincodeServerPort = ":50053"
)

type requestID struct {
	ID int32

	mux sync.Mutex
}

// Inc increments the counter for the given key.
func (c *requestID) Inc() int32 {
	c.mux.Lock()
	defer c.mux.Unlock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.ID++

	return c.ID
}

type chaincodeRequest struct {
	ID       int32
	Language string
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

var workerNum uint32 = 1
var greetingChanMap = cmap.New()
var chaincodeChan = make(chan chaincodeRequest)
var requestIDIssuer requestID

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	id := requestIDIssuer.Inc()
	idStr := strconv.FormatInt(int64(id), 10)
	waitChan := make(chan string)

	greetingChanMap.Set(idStr, waitChan)
	defer greetingChanMap.Remove(idStr)

	chaincodeChan <- chaincodeRequest{ID: id, Language: in.Language}
	select {
	case greeting := <-waitChan:
		return &pb.HelloReply{Message: greeting + " " + in.Name + "!"}, nil
	case <-time.After(time.Second * 10):
		return &pb.HelloReply{Message: ""}, nil
	}
}

func newServer() *server {
	s := &server{}
	return s
}

func connectToChaincodeServer(client pb.GreeterClient, id int) {
	ctx := context.Background()
	stream, err := client.LanguageService(ctx)

	if err != nil {
		fmt.Println("Error while create stream", stream)
	}

	// on receive response from chaincode service
	go func() {
		for {
			in, err := stream.Recv()
			fmt.Println("received", id, in)

			if err == io.EOF {
				fmt.Println("OEF received")
				continue
				// return
			}

			if err != nil {
				fmt.Println("err stream.Recv", err)
				continue
				// log.Fatalf("Failed to receive a note : %v", err)
			}

			id := strconv.FormatInt(int64(in.Id), 10)

			if tmp, ok := greetingChanMap.Get(id); ok {
				greetingChan := tmp.(chan string)
				greetingChan <- in.Greeting
			} else {
				fmt.Println("Channel not exist: ", in.Id)
			}

		}
	}()

	// send request to chaincode stream
	go func() {
		for req := range chaincodeChan {
			if err := stream.Send(&pb.LanguageRequest{Id: req.ID, Language: req.Language}); err != nil {
				fmt.Printf("Failed to send %v\n", req)
			}
			// log.Printf("Sent request: %v", req)
		}
	}()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter chaincodeServerHost!")

		return
	}

	chaincodeServerHost = os.Args[1]

	var wg sync.WaitGroup
	wg.Add(1)

	for i := 0; i < 2; i++ {
		// Chaincode Client
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())

		conn, err := grpc.Dial(chaincodeServerHost+chaincodeServerPort, opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		client := pb.NewGreeterClient(conn)
		go connectToChaincodeServer(client, i)
	}

	requestIDIssuer = requestID{}

	// Peer server
	lis, err := net.Listen("tcp4", peerServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serverOpts := []grpc.ServerOption{}
	s := grpc.NewServer(serverOpts...)

	pb.RegisterGreeterServer(s, newServer())

	log.Println("Ready to serve!")
	log.Println("===========================")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
