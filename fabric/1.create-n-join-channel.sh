
CHANNEL_NAME=vnpay-channel

echo "================ CREATE CHANNEL $CHANNEL_NAME ================"
docker exec cli \
peer channel create -o orderer.example.com:7050 \
-c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
sleep 5

echo "================ JOIN Peer0.Org1 INTO CHANNEL $CHANNEL_NAME ================"
docker exec cli \
peer channel join -b $CHANNEL_NAME.block
sleep 2

echo "================ JOIN Peer0.Org1 INTO CHANNEL $CHANNEL_NAME ================"
docker exec \
-e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp \
-e CORE_PEER_ADDRESS=peer1.org1.example.com:7051 \
-e CORE_PEER_LOCALMSPID="Org1MSP" \
-e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt \
cli peer channel join -b $CHANNEL_NAME.block
sleep 2
