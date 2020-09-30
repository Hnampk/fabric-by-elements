# Peer environment variables
source ./peer-env.sh

ORDERER_ADDRESS=orderer1.org1.example.com:7050

set -x
./bin/peer channel create -o $ORDERER_ADDRESS -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx \ 
# --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA
#  >&log.txt


set +x

