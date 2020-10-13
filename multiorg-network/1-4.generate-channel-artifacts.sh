CHANNEL_NAME_1=vnpay-channel-1
DIR_1=./channel-artifacts-1

echo
echo "#################################################################"
echo "### Generating channel configuration transaction 'channel.tx' ###"
echo "#################################################################"
set -x
./bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR_1} -outputCreateChannelTx ${DIR_1}/channel.tx -channelID $CHANNEL_NAME_1
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate channel configuration transaction..."
exit 1
fi

echo "#################################################################"
echo "#######    Generating anchor peer update for Org1MSP   ##########"
echo "#################################################################"
set -x
./bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR_1} -outputAnchorPeersUpdate ${DIR_1}/Org1MSPanchors.tx -channelID $CHANNEL_NAME_1 -asOrg Org1MSP
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for Org1MSP..."
exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org2MSP   ##########"
echo "#################################################################"
set -x
./bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR_1} -outputAnchorPeersUpdate ${DIR_1}/Org2MSPanchors.tx -channelID $CHANNEL_NAME_1 -asOrg Org2MSP
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for Org2MSP..."
exit 1
fi


# =====================================================================

CHANNEL_NAME_2=vnpay-channel-2
DIR_2=./channel-artifacts-2

echo
echo "#################################################################"
echo "### Generating channel configuration transaction 'channel.tx' ###"
echo "#################################################################"
set -x
./bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR_2} -outputCreateChannelTx ${DIR_2}/channel.tx -channelID $CHANNEL_NAME_2
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate channel configuration transaction..."
exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org3MSP   ##########"
echo "#################################################################"
set -x
./bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR_2} -outputAnchorPeersUpdate ${DIR_2}/Org3MSPanchors.tx -channelID $CHANNEL_NAME_2 -asOrg Org3MSP
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for Org3MSP..."
exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org2MSP   ##########"
echo "#################################################################"
set -x
./bin/configtxgen -profile TwoOrgsChannel -configPath ${DIR_2} -outputAnchorPeersUpdate ${DIR_2}/Org2MSPanchors.tx -channelID $CHANNEL_NAME_2 -asOrg Org2MSP
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate anchor peer update for Org2MSP..."
exit 1
fi
