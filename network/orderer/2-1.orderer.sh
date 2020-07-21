
export FABRIC_CFG_PATH=../

export FABRIC_LOGGING_SPEC=INFO
export ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
export ORDERER_GENERAL_LISTENPORT=7050
export ORDERER_GENERAL_GENESISMETHOD=file
export ORDERER_GENERAL_GENESISFILE=/home/nampkh/nampkh/my-fabric/network/channel-artifacts/genesis.block
export ORDERER_GENERAL_LOCALMSPID=OrdererMSP
export ORDERER_GENERAL_LOCALMSPDIR=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp
# enabled TLS
export ORDERER_GENERAL_TLS_ENABLED=true
export ORDERER_GENERAL_TLS_PRIVATEKEY=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key
export ORDERER_GENERAL_TLS_CERTIFICATE=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
export ORDERER_GENERAL_TLS_ROOTCAS=[/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt]
export ORDERER_KAFKA_TOPIC_REPLICATIONFACTOR=1
# export ORDERER_KAFKA_VERBOSE=true
# export ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
# export ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key
# export ORDERER_GENERAL_CLUSTER_ROOTCAS=[/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt]


../../new-bin/orderer
