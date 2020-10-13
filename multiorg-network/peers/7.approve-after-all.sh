

for i in $(seq 1 3)
do
    source ../peer$i/peer-env.sh

    CHANNEL_NAME=vnpay-channel-2
    CC_SRC_PATH="${PWD}/../../chaincodes/dockcc/simple-chaincode/src"
    LANGUAGE="golang"
    CC_NAME="mycc2"
    CC_VERSION="1"
    PKG_NAME=$CC_NAME$CC_VERSION
    PKG_DIR="../peer1/packed-cc"

    SEQUENCE=8
    ../../bin/peer lifecycle chaincode queryinstalled >$PKG_DIR/log.txt
    cat $PKG_DIR/log.txt
    PACKAGE_ID=`sed -n '/Package/{s/^Package ID: //; s/, Label:.*$//; p;}' $PKG_DIR/log.txt | tail -n 1`
    rm $PKG_DIR/log.txt
    echo $PACKAGE_ID

    ORDERER=orderer2
    ORDERER_PORT=8050

    export ORDERER_ADDRESS=$ORDERER.$ORG.example.com:${ORDERER_PORT}



    echo "================ APPROVE CHAINCODE ================"
    ../../bin/peer lifecycle chaincode approveformyorg \
    -o $ORDERER_ADDRESS \
    --channelID $CHANNEL_NAME \
    --name $CC_NAME \
    --version $CC_VERSION \
    --init-required \
    --package-id $PACKAGE_ID \
    --sequence $SEQUENCE \
    --waitForEvent \
    --tls $CORE_PEER_TLS_ENABLED \
    --cafile $ORDERER_CA
done

