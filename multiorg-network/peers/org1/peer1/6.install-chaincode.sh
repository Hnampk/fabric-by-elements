
# Peer environment variables
source ./peer-env.sh


CHANNEL_NAME=vnpay-channel-1
CC_SRC_PATH="${PWD}/../../../chaincodes/dockcc/accounting-cc/src"

LANGUAGE="golang"
CC_NAME="mycc1"
CC_VERSION="1"
PKG_NAME=$CC_NAME$CC_VERSION
PKG_DIR="../peer1/packed-cc"
CC_LABEL=$CC_NAME-$CC_VERSION

SEQUENCE=2


if [ ! -f $PKG_DIR/$PKG_NAME.tar.gz ]; then
echo "================ PACKAGE CHAINCODE $CC_NAME VERSION $CC_VERSION ================"
../../../bin/peer lifecycle chaincode package $PKG_DIR/$PKG_NAME.tar.gz --path $CC_SRC_PATH --lang $LANGUAGE --label $CC_LABEL
fi

echo "================ INSTALL CHAINCODE $CC_NAME VERSION $CC_VERSION ON $PEER.org1 ================"
../../../bin/peer lifecycle chaincode install $PKG_DIR/$PKG_NAME.tar.gz


echo "================ QUERY INSTALLED CHAINCODE TO GET PACKAGE_ID ================"
../../../bin/peer lifecycle chaincode queryinstalled | grep $CC_LABEL >$PKG_DIR/log.txt
cat $PKG_DIR/log.txt
PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' $PKG_DIR/log.txt`
rm $PKG_DIR/log.txt



echo "================ APPROVE CHAINCODE ================"
../../../bin/peer lifecycle chaincode approveformyorg \
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
../../../bin/peer lifecycle chaincode checkcommitreadiness \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--sequence $SEQUENCE \
--output json


echo "================ COMMIT CHAINCODE ================"
echo "================ This will start chaincode container ================"
../../../bin/peer lifecycle chaincode commit \
-o $ORDERER_ADDRESS \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--sequence $SEQUENCE \
--init-required \
--peerAddresses $CORE_PEER_ADDRESS \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \





echo "================ INVOKE CHAINCODE (INIT FUNCTION) ================"
../../../bin/peer chaincode invoke \
-o $ORDERER_ADDRESS \
--isInit \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses $CORE_PEER_ADDRESS \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
-c '{"Args":["Init"]}' \
--waitForEvent \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \

echo "================ TEST INVOKE CHAINCODE ================"
../../../bin/peer  chaincode invoke \
-o $ORDERER_ADDRESS \
-C $CHANNEL_NAME \
-n $CC_NAME \
--peerAddresses $CORE_PEER_ADDRESS \
--tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE \
-c '{"Args":["update","myvar","100","+"]}' \
--waitForEvent \
--tls $CORE_PEER_TLS_ENABLED \
--cafile $ORDERER_CA \

#echo "================ TEST QUERY CHAINCODE ================"
#../../../bin/peer chaincode query \
#-C $CHANNEL_NAME \
#-n $CC_NAME \
#-c '{"Args":["get","myvar"]}'


