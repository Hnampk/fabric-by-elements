source .env

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

export ORDERER_CA=$PWD/orderers/crypto-config/ordererOrganizations/$CONSORTIUM/orderers/$ORDERER.$CONSORTIUM/msp/tlscacerts/tlsca.$CONSORTIUM-cert.pem


CC_SRC_PATH="${PWD}/../chaincodes/dockcc/accounting-cc/src"
LANGUAGE="golang"
CC_NAME="mycc1"
CC_VERSION="1"
PKG_NAME=$CC_NAME$CC_VERSION
PKG_DIR="peers/packed-cc"
CC_LABEL=$CC_NAME-$CC_VERSION
SEQUENCE=1


echo "================ COMMIT CHAINCODE ================"
echo "================ This will start chaincode container ================"
../bin/peer lifecycle chaincode commit \
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
../bin/peer chaincode invoke \
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
wait