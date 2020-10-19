# START peer0.org2
source .env

export FABRIC_CFG_PATH=./peers
export PEER=peer0
export ORG=org2
export IP_ADDRESS=$ORG2_IP
export PORT=7051
export CC_PORT=7052
export CONSORTIUM=example.com

export CORE_PEER_LOCALMSPID=Org2MSP
export CORE_PEER_FILESYSTEMPATH=./data0
export CORE_OPERATIONS_LISTENADDRESS=0.0.0.0:9443

#Generic peer variables
export CORE_VM_ENDPOINT=unix://var/run/docker.sock

# the following setting starts chaincode containers on the same
# bridge network as the peers
# https://docs.docker.com/compose/networking/
export CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=my_fabric_test
export FABRIC_LOGGING_SPEC=INFO
#export FABRIC_LOGGING_SPEC=DEBUG
export CORE_PEER_TLS_ENABLED=false

export CORE_PEER_PROFILE_ENABLED=false
export CORE_PEER_MSPCONFIGPATH=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/msp
export CORE_PEER_TLS_CERT_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/ca.crt
export CORE_PEER_TLS_CLIENTROOTCAS_FILES=[${PWD}/peers/crypto-config/peerOrganizations/${ORG}.${CONSORTIUM}/peers/${PEER}.${ORG}.${CONSORTIUM}/tls/ca.crt]
# Peer specific variabes
export CORE_PEER_ID=${PEER}.${ORG}.${CONSORTIUM}
export CORE_PEER_ADDRESS=${PEER}.${ORG}.${CONSORTIUM}:${PORT}
export CORE_PEER_LISTENADDRESS=${PEER}.${ORG}.${CONSORTIUM}:${PORT}
export CORE_PEER_CHAINCODEADDRESS=${IP_ADDRESS}:${CC_PORT}
export CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:${CC_PORT}
export CORE_PEER_GOSSIP_BOOTSTRAP=peer1.${ORG}.${CONSORTIUM}:8051
export CORE_PEER_GOSSIP_EXTERNALENDPOINT=${PEER}.${ORG}.${CONSORTIUM}:${PORT}


../bin/peer node start &
