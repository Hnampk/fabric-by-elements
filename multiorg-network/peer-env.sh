PEER=peer0
ORG=org1
export CHANNEL_NAME=vnpay-channel


export CORE_PEER_MSPCONFIGPATH=${PWD}/peer/crypto-config/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com/msp
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=${PEER}.${ORG}.example.com:7051
export CORE_PEER_TLS_ENABLED=true


export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/peer/crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/ca.crt
export CORE_PEER_TLS_CERT_FILE=${PWD}/peer/crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=${PWD}/peer/crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/server.key


export ORDERER_CA=${PWD}/orderer/crypto-config/ordererOrganizations/${ORG}.example.com/orderers/orderer1.${ORG}.example.com/msp/tlscacerts/tlsca.${ORG}.example.com-cert.pem
