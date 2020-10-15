# Peer environment variables
export FABRIC_CFG_PATH=./peers

export PEER=peer1
export ORG=org1
export PORT=8051
export CORE_PEER_LOCALMSPID=Org1MSP
export CONSORTIUM=example.com

export ORDERER=orderer0
export ORDERER_PORT=7050

export CHAN_ARTI_PATH=./channel-artifacts
export CHANNEL_NAME=vnpay-channel

export CORE_PEER_TLS_ENABLED=false
export CORE_PEER_ADDRESS=$PEER.$ORG.$CONSORTIUM:$PORT
export CORE_PEER_MSPCONFIGPATH=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/users/Admin@$ORG.$CONSORTIUM/msp
export CORE_PEER_TLS_ROOTCERT_FILE=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/peers/$PEER.$ORG.$CONSORTIUM/tls/ca.crt
export CORE_PEER_TLS_CERT_FILE=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/peers/$PEER.$ORG.$CONSORTIUM/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/peers/$PEER.$ORG.$CONSORTIUM/tls/server.key

export ORDERER_ADDRESS=$ORDERER.$CONSORTIUM:$ORDERER_PORT
export ORDERER_CA=$PWD/orderers/crypto-config/ordererOrganizations/$CONSORTIUM/orderers/$ORDERER.$CONSORTIUM/msp/tlscacerts/tlsca.$CONSORTIUM-cert.pem


CC_SRC_PATH="${PWD}/../chaincodes/dockcc/accounting-cc/src"
LANGUAGE="golang"
CC_NAME="mycc1"
CC_VERSION="1"
PKG_NAME=$CC_NAME$CC_VERSION
PKG_DIR="peers/packed-cc"
CC_LABEL=$CC_NAME-$CC_VERSION
SEQUENCE=1


if [ ! -f $PKG_DIR/$PKG_NAME.tar.gz ]; then
echo "================ PACKAGE CHAINCODE $CC_NAME VERSION $CC_VERSION ================"
../bin/peer lifecycle chaincode package $PKG_DIR/$PKG_NAME.tar.gz --path $CC_SRC_PATH --lang $LANGUAGE --label $CC_LABEL
fi

echo "================ INSTALL CHAINCODE $CC_NAME VERSION $CC_VERSION ON $PEER.$ORG ================"
../bin/peer lifecycle chaincode install $PKG_DIR/$PKG_NAME.tar.gz


echo "================ QUERY INSTALLED CHAINCODE TO GET PACKAGE_ID ================"
../bin/peer lifecycle chaincode queryinstalled | grep $CC_LABEL >$PKG_DIR/log.txt
cat $PKG_DIR/log.txt
PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' $PKG_DIR/log.txt`
rm $PKG_DIR/log.txt
echo $PACKAGE_ID



echo "================ APPROVE CHAINCODE ================"
../bin/peer lifecycle chaincode approveformyorg \
-o $ORDERER_ADDRESS \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--package-id $PACKAGE_ID \
--sequence $SEQUENCE \
--waitForEvent \
#--tls $CORE_PEER_TLS_ENABLED \
#--cafile $ORDERER_CA \


echo "================ CHECK COMMIT READINESS ================"
../bin/peer lifecycle chaincode checkcommitreadiness \
--channelID $CHANNEL_NAME \
--name $CC_NAME \
--version $CC_VERSION \
--init-required \
--sequence $SEQUENCE \
--output json
