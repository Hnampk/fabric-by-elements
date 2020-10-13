PEER=peer0
ORG=org3
IP_ADDRESS=10.22.7.239
PORT=7051
CC_PORT=7052

export CORE_PEER_LOCALMSPID=Org3MSP

#Generic peer variables
export CORE_VM_ENDPOINT=unix://var/run/docker.sock

# the following setting starts chaincode containers on the same
# bridge network as the peers
# https://docs.docker.com/compose/networking/
export CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=my_fabric_test
export FABRIC_LOGGING_SPEC=INFO
#export FABRIC_LOGGING_SPEC=DEBUG
export CORE_PEER_TLS_ENABLED=true

export CORE_PEER_PROFILE_ENABLED=true
export CORE_PEER_MSPCONFIGPATH=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/msp
export CORE_PEER_TLS_CERT_FILE=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/ca.crt
# Peer specific variabes
export CORE_PEER_ID=${PEER}.${ORG}.example.com
export CORE_PEER_ADDRESS=${PEER}.${ORG}.example.com:${PORT}
export CORE_PEER_LISTENADDRESS=${PEER}.${ORG}.example.com:${PORT}
export CORE_PEER_CHAINCODEADDRESS=${IP_ADDRESS}:${CC_PORT}
export CORE_PEER_CHAINCODELISTENADDRESS=${IP_ADDRESS}:${CC_PORT}
export CORE_PEER_GOSSIP_BOOTSTRAP=peer0.${ORG}.example.com:7051
export CORE_PEER_GOSSIP_EXTERNALENDPOINT=${PEER}.${ORG}.example.com:${PORT}

../../../bin/peer node start
#../bin/main node start
