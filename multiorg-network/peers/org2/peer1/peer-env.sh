PEER=peer1
ORG=org2
PORT=7051
ORDERER=orderer1
ORDERER_PORT=7050

export CORE_PEER_LOCALMSPID=Org2MSP

export ORDERER_ADDRESS=$ORDERER.$ORG.example.com:${ORDERER_PORT}
export ORDERER_CA=${PWD}/../../../orderers/crypto-config/ordererOrganizations/${ORG}.example.com/orderers/$ORDERER.${ORG}.example.com/msp/tlscacerts/tlsca.${ORG}.example.com-cert.pem

export CORE_PEER_MSPCONFIGPATH=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com/msp
export CORE_PEER_ADDRESS=${PEER}.${ORG}.example.com:${PORT}
export CORE_PEER_TLS_ENABLED=true

export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/ca.crt
export CORE_PEER_TLS_CERT_FILE=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=${PWD}/../../crypto-config/peerOrganizations/${ORG}.example.com/peers/${PEER}.${ORG}.example.com/tls/server.key

