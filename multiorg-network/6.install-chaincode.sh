
# Peer environment variables
source ./peer-env.sh

ORDERER_ADDRESS=orderer1.org1.example.com:7050


CHANNEL_NAME=vnpay-channel
CC_SRC_PATH="${PWD}/chaincodes/dockcc/simple-chaincode/src"
LANGUAGE="golang"
CC_NAME="mycc"
CC_VERSION="1"
PKG_NAME=$CC_NAME$CC_VERSION
PKG_DIR="peer/packed-cc"



if [ -f $PKG_DIR/$PKG_NAME.tar.gz ]; then
echo "================ PACKAGE CHAINCODE $CC_NAME VERSION $CC_VERSION ================"
./bin/peer lifecycle chaincode package $PKG_DIR/$PKG_NAME.tar.gz --path $CC_SRC_PATH --lang $LANGUAGE --label $CC_NAME-$CC_VERSION
fi

echo "================ INSTALL CHAINCODE $CC_NAME VERSION $CC_VERSION ON peer0.org1 ================"
./bin/peer lifecycle chaincode install $PKG_DIR/$PKG_NAME.tar.gz


echo "================ QUERY INSTALLED CHAINCODE TO GET PACKAGE_ID ================"
./bin/peer lifecycle chaincode queryinstalled >$PKG_DIR/log.txt
cat $PKG_DIR/log.txt
PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' $PKG_DIR/log.txt`
rm $PKG_DIR/log.txt


SEQUENCE=1

echo "================ APPROVE CHAINCODE ================"
./bin/peer lifecycle chaincode approveformyorg \
-o $ORDERER_ADDRESS \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--package-id $PACKAGE_ID \
--sequence $SEQUENCE \
--waitForEvent \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \


echo "================ CHECK COMMIT READINESS ================"
./bin/peer lifecycle chaincode checkcommitreadiness \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--sequence $SEQUENCE \
--output json


echo "================ COMMIT CHAINCODE ================"
echo "================ This will start chaincode container ================"
./bin/peer lifecycle chaincode commit \
-o $ORDERER_ADDRESS \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--sequence $SEQUENCE \
--init-required \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \





echo "================ INVOKE CHAINCODE (INIT FUNCTION) ================"
./bin/peer chaincode invoke \
-o $ORDERER_ADDRESS \
--isInit \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
-c '{"Args":["Init"]}' \
--waitForEvent \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \

echo "================ TEST INVOKE CHAINCODE ================"
./bin/peer  chaincode invoke \
-o $ORDERER_ADDRESS \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
-c '{"Args":["update","myvar","100","+"]}' \
--waitForEvent \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \

echo "================ TEST QUERY CHAINCODE ================"
./bin/peer chaincode query \
-C $CHANNEL_NAME \
-n $CC_NAME \
-c '{"Args":["get","myvar"]}'