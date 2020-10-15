# Generate orderer genesis block
# Generate channel configuration transaction - channel.tx
# Generate anchor peer update for Orgs

DIR=./channel-artifacts
CHANNEL_NAME=vnpay-channel
ORGMSP=Org1MSP

if [ -d ${DIR} ]; then
    rm -f ${DIR}/*.tx
    rm -f ${DIR}/*.block
fi


echo "##########################################################"
echo "#########  Generating Orderer Genesis block ##############"
echo "##########################################################"
# Note: For some unknown reason (at least for now) the block file can't be
# named orderer.genesis.block or the orderer will fail to launch!
set -x
../bin/configtxgen -profile SoloOrdererGenesis -configPath ${DIR}  -channelID byfn-sys-channel -outputBlock ${DIR}/genesis.block
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate orderer genesis block..."
sleep 5
exit 1
fi

echo
echo "#################################################################"
echo "### Generating channel configuration transaction 'channel.tx' ###"
echo "#################################################################"
set -x
../bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR} -outputCreateChannelTx ${DIR}/channel.tx -channelID $CHANNEL_NAME
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate channel configuration transaction..."
sleep 5
exit 1
fi

echo "#################################################################"
echo "#######    Generating anchor peer update for ${ORGMSP}   ########"
echo "#################################################################"
set -x
../bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR} -outputAnchorPeersUpdate ${DIR}/${ORGMSP}anchors.tx -channelID $CHANNEL_NAME -asOrg ${ORGMSP}
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for ${ORGMSP}..."
sleep 5
exit 1
fi


ORGMSP=Org2MSP

echo "#################################################################"
echo "#######    Generating anchor peer update for ${ORGMSP}   ########"
echo "#################################################################"
set -x
../bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR} -outputAnchorPeersUpdate ${DIR}/${ORGMSP}anchors.tx -channelID $CHANNEL_NAME -asOrg ${ORGMSP}
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for ${ORGMSP}..."
sleep 5
exit 1
fi