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

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	pb "example.com/grpc-fabric-test/helloworld"
	cmap "github.com/orcaman/concurrent-map"

	"google.golang.org/grpc"
)

var (
	peerHost        = "localhost"
	port            = "50051"
	httpPort        = "8090"
	defaultName     = "world"
	isUnary         = true
	requestIDIssuer requestID

	requestChan     = make(chan HiRequest)
	responseChanMap = cmap.New()
)

type GreetingRequest struct {
	Name     string
	Language string
}

type HiRequest struct {
	ID  string
	req GreetingRequest
}

type requestID struct {
	ID int

	mux sync.Mutex
}

// Inc increments the counter for the given key.
func (c *requestID) Inc() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.ID++

	return c.ID
}

/*
	Unary grpc
*/
func hi(ctx context.Context, client pb.GreeterClient, request GreetingRequest) (string, error) {
	// Contact the server and print out its response.
	response, err := client.SayHello(ctx, &pb.HelloRequest{Name: request.Name, Language: request.Language})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return "", err
	}

	return response.GetMessage(), nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Please enter peer host!")
		return
	}

	peerHost = os.Args[1]
	requestIDIssuer = requestID{}

	for i := 0; i < 10; i++ {
		go createPool(i)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/invoke", invokeHandler)

	// listen and serve
	fmt.Println("Server listen on port", httpPort)
	http.ListenAndServe(":"+httpPort, mux)

	return
}

func createPool(id int) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(peerHost+":"+port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	func() {
		for req := range requestChan {
			ctx := context.Background()

			greeting, err := hi(ctx, client, req.req)

			fmt.Println(greeting)

			if tmp, ok := responseChanMap.Get(req.ID); ok {
				responseMap := tmp.(chan string)
				responseChanMap.Remove(req.ID)

				if err != nil {
					log.Fatalf("Error occured", err.Error(), http.StatusInternalServerError)
					fmt.Println("ERROR", err)
					responseMap <- ""

					continue
				}

				responseMap <- greeting
			} else {
				fmt.Println("Response channel not found, id: ", req.ID)
			}

		}
	}()
}

func invokeHandler(res http.ResponseWriter, req *http.Request) {
	greetingRequest, err := resolveHTTPRequest(req)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	responseChan := make(chan string)

	id := strconv.Itoa(requestIDIssuer.Inc())

	responseChanMap.Set(id, responseChan)
	requestChan <- HiRequest{ID: id, req: *greetingRequest}

	select {
	case resp := <-responseChan:
		if resp == "" {
			http.Error(res, "timeout from peer - "+id, http.StatusInternalServerError)
		}

		fmt.Fprint(res, resp)
	case <-time.After(time.Second * 10):
		http.Error(res, "timeout - "+id, http.StatusInternalServerError)
	}

	return
}

func resolveHTTPRequest(req *http.Request) (*GreetingRequest, error) {
	var greetingRequest GreetingRequest

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, err
	}

	json.Unmarshal(body, &greetingRequest)

	return &greetingRequest, nil
}
