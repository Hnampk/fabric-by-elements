
# Peer environment variables
export FABRIC_CFG_PATH=./peers

export PEER=peer1
export ORG=org2
export PORT=8051
export CORE_PEER_LOCALMSPID=Org2MSP
export CONSORTIUM=example.com

export CHANNEL_NAME=vnpay-channel

export CORE_PEER_TLS_ENABLED=false
export CORE_PEER_ADDRESS=$PEER.$ORG.$CONSORTIUM:$PORT
export CORE_PEER_MSPCONFIGPATH=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/users/Admin@$ORG.$CONSORTIUM/msp
export CORE_PEER_TLS_ROOTCERT_FILE=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/peers/$PEER.$ORG.$CONSORTIUM/tls/ca.crt
export CORE_PEER_TLS_CERT_FILE=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/peers/$PEER.$ORG.$CONSORTIUM/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=$PWD/peers/crypto-config/peerOrganizations/$ORG.$CONSORTIUM/peers/$PEER.$ORG.$CONSORTIUM/tls/server.key


CC_NAME="mycc1"


echo "================ update-account-service ================"
../bin/peer chaincode query \
-C $CHANNEL_NAME \
-n $CC_NAME \
-c '{"Args":["update-account-service","10.22.7.237","50052"]}'
