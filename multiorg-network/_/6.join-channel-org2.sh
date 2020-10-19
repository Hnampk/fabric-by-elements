source .env

# Peer environment variables
export FABRIC_CFG_PATH=./peers

export PEER=peer0
export ORG=org2
export PORT=7051
export CORE_PEER_LOCALMSPID=Org2MSP
export CONSORTIUM=example.com

export ORDERER=orderer0
export ORDERER_PORT=7050

export CHAN_ARTI_PATH=./channel-artifacts
export CHANNEL_NAME=vnpay-channel

export CORE_PEER_TLS_ENABLED=false
export CORE_PEER_ADDRESS=${PEER}.${ORG}.${CONSORTIUM}:${PORT}
export CORE_PEER_MSPCONFIGPATH=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/users/Admin@${ORG}.${CONSORTIUM}/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/ca.crt
export CORE_PEER_TLS_CERT_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/server.key

export ORDERER_CA=${PWD}/orderers/crypto-config/ordererOrganizations/${CONSORTIUM}/orderers/${ORDERER}.${CONSORTIUM}/msp/tlscacerts/tlsca.${CONSORTIUM}-cert.pem


echo
echo "#################################################################"
echo "### Start creating channel ${CHANNEL_NAME} ########################"
echo "#################################################################"
set -x
../bin/peer channel create -o ${ORDERER_ADDRESS} -c ${CHANNEL_NAME} -f ${CHAN_ARTI_PATH}/channel.tx --outputBlock ${CHAN_ARTI_PATH}/${CHANNEL_NAME}.block --tls $CORE_PEER_TLS_ENABLED --cafile ${ORDERER_CA}
set +x
sleep 2

echo
echo "#################################################################"
echo "### Start joining ${PEER}.${ORG} into ${CHANNEL_NAME} #################"
echo "#################################################################"
set -x
../bin/peer channel join -b ${CHAN_ARTI_PATH}/${CHANNEL_NAME}.block
set +x
sleep 5

set -x
../bin/peer channel list
set +x

# ================================

export PEER=peer1
export ORG=org2
export PORT=8051
export CORE_PEER_LOCALMSPID=Org2MSP
export CONSORTIUM=example.com

export CORE_PEER_TLS_ENABLED=false
export CORE_PEER_ADDRESS=${PEER}.${ORG}.${CONSORTIUM}:${PORT}
export CORE_PEER_MSPCONFIGPATH=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/users/Admin@${ORG}.${CONSORTIUM}/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/ca.crt
export CORE_PEER_TLS_CERT_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/server.key

echo
echo "#################################################################"
echo "### Start joining ${PEER}.${ORG} into ${CHANNEL_NAME} #################"
echo "#################################################################"
set -x
../bin/peer channel join -b ${CHAN_ARTI_PATH}/${CHANNEL_NAME}.block
set +x
sleep 5

set -x
../bin/peer channel list
set +x

set -x
../bin/peer channel update -o ${ORDERER_ADDRESS} -c ${CHANNEL_NAME} -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx --tls ${CORE_PEER_TLS_ENABLED} --cafile ${ORDERER_CA}
res=$?
set +x
