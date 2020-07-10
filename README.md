# fabric-by-elements

network/

0: Xóa dữ liệu sổ cái + các file cấu hình liên quan đến channel

1: Sinh các certificate của orderer, peer và cấu hình channel

2: Khởi chạy orderer trên port 7050

3: Khởi chạy peer trên port 7051

4: Tạo channel dựa theo file cấu hình configtx.yaml => sinh ra file mychannel.block

5: Join peer vào channel sử dụng file mychannel.block

6: Update anchor peer

7: Đóng gói chaincode (Từ version 2.0)

8:  .1 Cài đặt chaincode lên peer, Kiểm tra lại việc cài có thành công hay không bằng command "../bin/peer lifecycle chaincode queryinstalled"
    .2 Approve chaincode đã cài


test fabric-go-sdk

0. Generate artifacts
    - channel genesis.block:
		../bin/configtxgen -profile TwoOrgsOrdererGenesis -channelID byfn-sys-channel -outputBlock ./channel-artifacts/genesis.block
    
    - channel.tx
        ../bin/configtxgen -profile TwoOrgsChannel -channelID vnpay-channel -outputCreateChannelTx ./channel-artifacts/channel.tx

1. start.sh => các container trong fabric/

2. cd vnpay.vn/

3. create channel (go_2) >> join channel (go_3)

4. cài chaincode bằng commandline (go-sdk on pending)
    
    - docker exec -it cli bash

    + environment variable peer0.org1
    CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    CORE_PEER_LOCALMSPID="Org1MSP"
    CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

    + environment variable  peer0.org2
    CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    CORE_PEER_ADDRESS=peer0.org2.example.com:7051
    CORE_PEER_LOCALMSPID="Org2MSP"
    CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt


    - Create channel
        export CHANNEL_NAME=vnpay-channel
        peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

    - Join channel
        peer channel join -b mychannel.block

    - Packaging chaincode
        peer lifecycle chaincode package mycc.tar.gz --path github.com/chaincode/abstore/go/ --lang golang --label mycc_1
        
        # HIGH THOUGHPUT CC: 
        peer lifecycle chaincode package mycc.tar.gz --path /opt/gopath/src/github.com/chaincode/high-throughput --lang golang --label mycc_1


    - Install chaincode
        peer lifecycle chaincode install mycc.tar.gz

    - Query installed chaincode to get packageID
        peer lifecycle chaincode queryinstalled >&log.txt
        cat log.txt
        PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' log.txt`

    - Approve chaincode
        peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls false --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID vnpay-channel --name mycc --version 1.0 --init-required --package-id $PACKAGE_ID --sequence 1 --waitForEvent

    - Check commit readiness
        peer lifecycle chaincode checkcommitreadiness --channelID $CHANNEL_NAME --name mycc --version 1.0 --init-required --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --output json

    - Commit chaincode
        peer lifecycle chaincode commit -o orderer.example.com:7050 --channelID $CHANNEL_NAME --name mycc --version 1.0 --sequence 1 --init-required --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

    - Invoke chaincode (Init function))
    peer chaincode invoke -o orderer.example.com:7050 --isInit --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C $CHANNEL_NAME -n mycc --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["Init","a","100","b","100"]}' --waitForEvent

    - Test Query chaincode
        peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["query","a"]}'

    - Test Invoke chaincode (invoke function)
    peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C $CHANNEL_NAME -n mycc --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["invoke","a","b","10"]}' --waitForEvent

5. Query chaincode (go_8)

6. Invoke chaincode (go_7)

# API server: go run go ht_invoke.go