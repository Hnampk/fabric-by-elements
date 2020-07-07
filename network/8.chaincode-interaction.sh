verifyResult() {
  if [ $1 -ne 0 ]; then
    echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
    echo
    exit 1
  fi
}

# queryInstalled PEER ORG
queryInstalled() {
  PEER=$1
  ORG=$2
  set -x
  ../bin/peer lifecycle chaincode queryinstalled >&log.txt
  res=$?
  set +x
  cat log.txt
  PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' log.txt`
  verifyResult $res "Query installed on peer${PEER}.org${ORG} has failed"
  echo PackageID is ${PACKAGE_ID}
  echo "===================== Query installed successful on peer${PEER}.org${ORG} on channel ===================== "
  echo
}

# approveForMyOrg VERSION PEER ORG
approveForMyOrg() {
  VERSION=$1
  PEER=$2
  ORG=$3

#   if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
#     set -x
#     peer lifecycle chaincode approveformyorg --channelID $CHANNEL_NAME --name mycc --version ${VERSION} --init-required --package-id ${PACKAGE_ID} --sequence ${VERSION} --waitForEvent >&log.txt
#     set +x
#   else
  set -x
  ../bin/peer lifecycle chaincode approveformyorg --tls true --cafile $ORDERER_CA --channelID $CHANNEL_NAME --name mycc --version ${VERSION} --init-required --package-id ${PACKAGE_ID} --sequence ${VERSION} --waitForEvent >&log.txt
  set +x
#   fi
  cat log.txt
  verifyResult $res "Chaincode definition approved on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' failed"
  echo "===================== Chaincode definition approved on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
  echo
}

export ORDERER_CA=/home/nampkh/nampkh/my-fabric/network/orderer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export CORE_PEER_TLS_ENABLED=true

source ./peer-env.sh

queryInstalled 0 1

## approve the definition for org1
approveForMyOrg 1 0 1