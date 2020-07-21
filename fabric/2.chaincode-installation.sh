
CHANNEL_NAME=vnpay-channel
CC_SRC_PATH=/opt/gopath/src/github.com/chaincode/high-throughput
LANGUAGE="golang"
CC_NAME="mycc"
CC_VERSION="1"
PKG_NAME=$CC_NAME$CC_VERSION

echo "================ PACKAGE CHAINCODE $CC_NAME VERSION $CC_VERSION ================"
docker exec cli \
peer lifecycle chaincode package $PKG_NAME.tar.gz --path $CC_SRC_PATH --lang $LANGUAGE --label $CC_NAME-$CC_VERSION



echo "================ INSTALL CHAINCODE $CC_NAME VERSION $CC_VERSION ON peer0.org1 ================"
docker exec cli \
peer lifecycle chaincode install $PKG_NAME.tar.gz

echo "================ INSTALL CHAINCODE $CC_NAME VERSION $CC_VERSION ON peer1.org1 ================"
docker exec \
-e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp \
-e CORE_PEER_ADDRESS=peer1.org1.example.com:7051 \
-e CORE_PEER_LOCALMSPID="Org1MSP" \
-e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt \
cli peer lifecycle chaincode install $PKG_NAME.tar.gz


echo "================ QUERY INSTALLED CHAINCODE TO GET PACKAGE_ID ================"
docker exec cli \
peer lifecycle chaincode queryinstalled >&log.txt
cat log.txt
PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' log.txt`
rm log.txt

echo "================ APPROVE CHAINCODE ================"
docker exec cli \
peer lifecycle chaincode approveformyorg \
-o orderer.example.com:7050 --tls false \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--package-id $PACKAGE_ID \
--sequence 1 \
--waitForEvent

echo "================ CHECK COMMIT READINESS ================"
docker exec cli \
peer lifecycle chaincode checkcommitreadiness \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--sequence 1 \
--tls false \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
--output json

echo "================ COMMIT CHAINCODE ================"
docker exec cli \
peer lifecycle chaincode commit \
-o orderer.example.com:7050 \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--sequence 1 \
--init-required \
--tls false \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

echo "================ INVOKE CHAINCODE (INIT FUNCTION) ================"
docker exec cli \
peer chaincode invoke \
-o orderer.example.com:7050 \
--isInit \
--tls false \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
-c '{"Args":["Init"]}' \
--waitForEvent

echo "================ TEST INVOKE CHAINCODE ================"
docker exec cli \
peer chaincode invoke \
-o orderer.example.com:7050 \
--tls false \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
-c '{"Args":["update","myvar","100","+"]}' \
--waitForEvent

echo "================ TEST QUERY CHAINCODE ================"
docker exec cli \
peer chaincode query \
-C $CHANNEL_NAME \
-n $CC_NAME \
-c '{"Args":["get","myvar"]}'

