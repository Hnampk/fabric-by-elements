package main

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"

	// fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	// "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func randomUsername() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}

func main() {
	fmt.Println("1. Instantiate a fabsdk instance using a configuration.")
	// Create SDK setup for the integration tests
	networkConfigFile := "network.yaml"
	configProvider := config.FromFile(networkConfigFile)

	// configBackend, err := configProvider()
	// if err != nil {
	// 	fmt.Println(err, "unable to load config backend")
	// }
	// defEndpointConfig, err := fabImpl.ConfigFromBackend(configBackend...)
	// if err != nil {
	// 	fmt.Println(err, "unable create endpoint from config")
	// }
	// fmt.Println(defEndpointConfig)

	// defCryptoConfig := cryptosuite.ConfigFromBackend(configBackend...)
	// c, err := cryptosuite.BuildCryptoSuiteConfigFromOptions(cryptoConfigs...)
	// if err != nil {
	// 		fmt.Println(err, "unable create cryptosuit")
	// }
	// fmt.Println(defCryptoConfig)
	sdk, err1 := fabsdk.New(configProvider)
	if err1 != nil {
		fmt.Println("failed to create sdk", err1)
		return
	}
	defer sdk.Close()

	fmt.Println("2. Prepare client context.")
	fmt.Println("2.1 Create a context based on a user and organization, using your fabsdk instance.")

	OrgName := "org1"

	clientContext := sdk.Context()

	fmt.Println("3. Create a client instance using its New func, passing the context.")
	fmt.Println("3.1 Create msp client.")

	// Create msp client
	mspClient, err := mspclient.New(clientContext, mspclient.WithOrg(OrgName))
	if err != nil {
		fmt.Println("failed to create msp client: ", err)
		return
	}
	//
	fmt.Println("3.2 Enroll admin user.")
	OrgAdmin := "admin"
	OrgAdminPassword := "adminpw"
	err = mspClient.Enroll(OrgAdmin, mspclient.WithSecret(OrgAdminPassword), mspclient.WithLabel("admin"))
	if err != nil {
		fmt.Printf("failed to enroll admin user: %s\n", err)
		return
	}

	// fmt.Println("4. Use the funcs provided by each client to create your solution.")
	// username := "abc3"
	// fmt.Println("4.1 Register user: " + username)

	// enrollmentSecret, err := mspClient.Register(&mspclient.RegistrationRequest{Name: username, Affiliation: OrgName})
	// if err != nil {
	// 	fmt.Printf("Register return error %s\n", err)
	// 	return
	// }

	// fmt.Println("4.2 Enroll user.")
	// err = mspClient.Enroll(username, mspclient.WithSecret(enrollmentSecret))
	// if err != nil {
	// 	fmt.Printf("failed to enroll user: %s\n", err)
	// 	return
	// }
	// fmt.Println("4.3 Enroll user is completed")

	// fmt.Println("5. Call fabsdk.Close() to release resources and caches.")
}
