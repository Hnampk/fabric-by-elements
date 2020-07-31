# Peer environment variables
source ./peer-env.sh

export CHANNEL_NAME=vnpay-channel
export ORDERER_CA=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem

set -x
../bin/peer channel create -o localhost:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx \ 
# --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA
#  >&log.txt


set +x

