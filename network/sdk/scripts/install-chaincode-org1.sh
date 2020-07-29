export CORE_PEER_MSPCONFIGPATH=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

export ORDERER_CA=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export FABRIC_CFG_PATH=/home/nampkh/nampkh/my-fabric/network



export CORE_PEER_TLS_CERT_FILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
export CORE_PEER_TLS_ENABLED=false


CHANNEL_NAME=vnpay-channel
CC_SRC_PATH="/home/nampkh/nampkh/my-fabric/chaincode/high-throughput"
LANGUAGE="golang"
CC_NAME="mycc"
CC_VERSION="1"
PKG_NAME=$CC_NAME$CC_VERSION
PKG_DIR="../data"
ORDERER_HOST="orderer.example.com"



echo "================ PACKAGE CHAINCODE $CC_NAME VERSION $CC_VERSION ================"
../../../new-bin/peer lifecycle chaincode package $PKG_DIR/$PKG_NAME.tar.gz --path $CC_SRC_PATH --lang $LANGUAGE --label $CC_NAME-$CC_VERSION


echo "================ INSTALL CHAINCODE $CC_NAME VERSION $CC_VERSION ON peer0.org1 ================"
../../../new-bin/peer lifecycle chaincode install $PKG_DIR/$PKG_NAME.tar.gz


echo "================ QUERY INSTALLED CHAINCODE TO GET PACKAGE_ID ================"
../../../new-bin/peer lifecycle chaincode queryinstalled >$PKG_DIR/log.txt
cat $PKG_DIR/log.txt
PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' $PKG_DIR/log.txt`
rm $PKG_DIR/log.txt


SEQUENCE=1

echo "================ APPROVE CHAINCODE ================"
../../../new-bin/peer lifecycle chaincode approveformyorg \
-o $ORDERER_HOST:7050 \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--package-id $PACKAGE_ID \
--sequence $SEQUENCE \
--waitForEvent \
# --tls $CORE_PEER_TLS_ENABLED \
# --cafile $ORDERER_CA \


echo "================ CHECK COMMIT READINESS ================"
../../../new-bin/peer lifecycle chaincode checkcommitreadiness \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--sequence $SEQUENCE \
--output json \
# --tls $CORE_PEER_TLS_ENABLED \
# --cafile $ORDERER_CA \




echo "================ COMMIT CHAINCODE ================"
echo "================ This will start chaincode container ================"
../../../new-bin/peer lifecycle chaincode commit \
-o $ORDERER_HOST:7050 \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--sequence $SEQUENCE \
--init-required \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
# --tls $CORE_PEER_TLS_ENABLED \
# --cafile $ORDERER_CA \








echo "================ INVOKE CHAINCODE (INIT FUNCTION) ================"
../../../new-bin/peer chaincode invoke \
-o $ORDERER_HOST:7050 \
--isInit \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
-c '{"Args":["Init"]}' \
--waitForEvent \
# --tls $CORE_PEER_TLS_ENABLED \
# --cafile $ORDERER_CA \

echo "================ TEST INVOKE CHAINCODE ================"
../../../new-bin/peer  chaincode invoke \
-o $ORDERER_HOST:7050 \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
-c '{"Args":["update","myvar","100","+"]}' \
# --waitForEvent \
# --tls $CORE_PEER_TLS_ENABLED \
# --cafile $ORDERER_CA \

echo "================ TEST QUERY CHAINCODE ================"
../../../new-bin/peer chaincode query \
-C $CHANNEL_NAME \
-n $CC_NAME \
-c '{"Args":["get","myvar"]}'

