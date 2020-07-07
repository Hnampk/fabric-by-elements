export FABRIC_CFG_PATH=${PWD}/../
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
export CORE_PEER_MSPCONFIGPATH=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp
export CORE_PEER_TLS_CERT_FILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
# Peer specific variabes
export CORE_PEER_ID=peer0.org1.example.com
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_LISTENADDRESS=localhost:7051
export CORE_PEER_CHAINCODEADDRESS=localhost:7052
export CORE_PEER_CHAINCODELISTENADDRESS=localhost:7052
export CORE_PEER_GOSSIP_BOOTSTRAP=localhost:7051
export CORE_PEER_GOSSIP_EXTERNALENDPOINT=localhost:7051
export CORE_PEER_LOCALMSPID=Org1MSP

../../bin/peer node start
