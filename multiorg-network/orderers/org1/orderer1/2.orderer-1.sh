
ORDERER=orderer1
ORG=org1
PORT=7050
CHAN_ARTI_PATH=../../../channel-artifacts-1

export ORDERER_GENERAL_LOCALMSPID=OrdererOrg1MSP

export FABRIC_LOGGING_SPEC=INFO
# export FABRIC_LOGGING_SPEC=DEBUG
export ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
export ORDERER_GENERAL_LISTENPORT=${PORT}
export ORDERER_GENERAL_GENESISMETHOD=file
export ORDERER_GENERAL_GENESISFILE=${PWD}/$CHAN_ARTI_PATH/genesis.block
export ORDERER_GENERAL_LOCALMSPDIR=${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/msp
# enabled TLS
export ORDERER_GENERAL_TLS_ENABLED=true

export ORDERER_GENERAL_TLS_PRIVATEKEY=${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/tls/server.key
export ORDERER_GENERAL_TLS_CERTIFICATE=${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/tls/server.crt
export ORDERER_GENERAL_TLS_ROOTCAS=[${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/tls/ca.crt]
export ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/tls/server.crt
export ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/tls/server.key
export ORDERER_GENERAL_CLUSTER_ROOTCAS=[${PWD}/../../crypto-config/ordererOrganizations/${ORG}.example.com/orderers/${ORDERER}.${ORG}.example.com/tls/ca.crt]


../../../bin/orderer
