verifyResult() {
  if [ $1 -ne 0 ]; then
    echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
    echo
    exit 1
  fi
}

packageChaincode() {
  VERSION=$1
  PEER=$2
  ORG=$3
  set -x
  ../bin/peer lifecycle chaincode package mycc.tar.gz --path ${CC_SRC_PATH} --lang ${CC_RUNTIME_LANGUAGE} --label mycc_${VERSION} >&log.txt
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode packaging on peer${PEER}.org${ORG} has failed"
  echo "===================== Chaincode is packaged on peer${PEER}.org${ORG} ===================== "
  echo
}

installChaincode() {
  PEER=$1
  ORG=$2
  set -x
  ../bin/peer lifecycle chaincode install mycc.tar.gz >&log.txt
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode installation on peer${PEER}.org${ORG} has failed"
  echo "===================== Chaincode is installed on peer${PEER}.org${ORG} ===================== "
  echo
}

source ./peer-env.sh

CC_RUNTIME_LANGUAGE=golang
# CC_RUNTIME_LANGUAGE=node
# CC_SRC_PATH="/home/nampkh/nampkh/my-fabric/chaincode/abstore/javascript"
CC_SRC_PATH="/home/nampkh/nampkh/my-fabric/chaincode/abstore/go"
packageChaincode 1 0 1

sleep 2

echo "Installing chaincode on peer0.org1..."
installChaincode 0 1
