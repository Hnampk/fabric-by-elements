# START peer1.org1

export FABRIC_CFG_PATH=./peers
export PEER=peer1
export ORG=org1
export IP_ADDRESS=172.16.79.8
export PORT=8051
export CC_PORT=8052
export CONSORTIUM=example.com

export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_FILESYSTEMPATH=./data1
export CORE_OPERATIONS_LISTENADDRESS=0.0.0.0:10443

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
export CORE_PEER_GOSSIP_BOOTSTRAP=peer0.${ORG}.${CONSORTIUM}:7051
export CORE_PEER_GOSSIP_EXTERNALENDPOINT=${PEER}.${ORG}.${CONSORTIUM}:${PORT}


../bin/peer node start &