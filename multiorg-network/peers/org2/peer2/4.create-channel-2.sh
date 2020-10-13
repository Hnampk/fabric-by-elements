# Peer environment variables
source ./peer-env.sh

CHANNEL_NAME=vnpay-channel-2
CHAN_ARTI_PATH=../../../channel-artifacts-2
ORG=org2
ORDERER=orderer2
ORDERER_PORT=7050

export ORDERER_ADDRESS=$ORDERER.$ORG.example.com:${ORDERER_PORT}
export ORDERER_CA=${PWD}/../../../orderers/crypto-config/ordererOrganizations/${ORG}.example.com/orderers/$ORDERER.${ORG}.example.com/msp/tlscacerts/tlsca.${ORG}.example.com-cert.pem

set -x
../../../bin/peer channel create -o $ORDERER_ADDRESS -c $CHANNEL_NAME -f ${CHAN_ARTI_PATH}/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA
set +x

