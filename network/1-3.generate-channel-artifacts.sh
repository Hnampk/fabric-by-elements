
CHANNEL_NAME=vnpay-channel

echo "##########################################################"
echo "#########  Generating Orderer Genesis block ##############"
echo "##########################################################"
# Note: For some unknown reason (at least for now) the block file can't be
# named orderer.genesis.block or the orderer will fail to launch!
set -x
../bin/configtxgen -profile TwoOrgsOrdererGenesis -configPath ./  -channelID byfn-sys-channel -outputBlock ./channel-artifacts/genesis.block
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate orderer genesis block..."
exit 1
fi
echo
echo "#################################################################"
echo "### Generating channel configuration transaction 'channel.tx' ###"
echo "#################################################################"
set -x
../bin/configtxgen -profile TwoOrgsChannel -configPath ./ -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate channel configuration transaction..."
exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org1MSP   ##########"
echo "#################################################################"
set -x
../bin/configtxgen -profile TwoOrgsChannel -configPath ./ -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for Org1MSP..."
exit 1
fi