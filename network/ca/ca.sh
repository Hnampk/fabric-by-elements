export FABRIC_CA_HOME=/home/nampkh/nampkh/my-fabric/network/ca/data
export FABRIC_CA_SERVER_CA_NAME=ca-org1
export FABRIC_CA_SERVER_CA_CERTFILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem
export FABRIC_CA_SERVER_CA_KEYFILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/ca/priv_sk
export FABRIC_CA_SERVER_TLS_ENABLED=true
export FABRIC_CA_SERVER_TLS_CERTFILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem
export FABRIC_CA_SERVER_TLS_KEYFILE=/home/nampkh/nampkh/my-fabric/network/peer/crypto-config/peerOrganizations/org1.example.com/ca/priv_sk

../../new-bin/fabric-ca-server start -b admin:adminpw