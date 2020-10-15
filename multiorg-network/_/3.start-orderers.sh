
export FABRIC_CFG_PATH=./orderers
export ORDERER=orderer0
export ORG=org1
export PORT=7050
export CHAN_ARTI_PATH=./channel-artifacts

export ORDERER_GENERAL_LOCALMSPID=OrdererOrgMSP

export FABRIC_LOGGING_SPEC=INFO
# export FABRIC_LOGGING_SPEC=DEBUG
export ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
export ORDERER_GENERAL_LISTENPORT=${PORT}
export ORDERER_GENERAL_GENESISMETHOD=file
export ORDERER_GENERAL_GENESISFILE=${PWD}/$CHAN_ARTI_PATH/genesis.block
export ORDERER_GENERAL_LOCALMSPDIR=${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/msp
# enabled TLS
export ORDERER_GENERAL_TLS_ENABLED=false

export ORDERER_GENERAL_TLS_PRIVATEKEY=${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/tls/server.key
export ORDERER_GENERAL_TLS_CERTIFICATE=${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/tls/server.crt
export ORDERER_GENERAL_TLS_ROOTCAS=[${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/tls/ca.crt]
# export ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/tls/server.crt
# export ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/tls/server.key
# export ORDERER_GENERAL_CLUSTER_ROOTCAS=[${PWD}/orderers/crypto-config/ordererOrganizations/example.com/orderers/${ORDERER}.example.com/tls/ca.crt]


../bin/orderer &
