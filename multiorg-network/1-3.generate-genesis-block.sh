DIR_1=./channel-artifacts-1

echo "##########################################################"
echo "#########  Generating Orderer Genesis block ##############"
echo "##########################################################"
# Note: For some unknown reason (at least for now) the block file can't be
# named orderer.genesis.block or the orderer will fail to launch!
set -x
./bin/configtxgen -profile SampleMultiNodeEtcdRaft -configPath ${DIR_1}  -channelID byfn-sys-channel -outputBlock ${DIR_1}/genesis.block
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate orderer genesis block..."
exit 1
fi


# =====================================================================

DIR_2=./channel-artifacts-2

echo "##########################################################"
echo "#########  Generating Orderer Genesis block ##############"
echo "##########################################################"
# Note: For some unknown reason (at least for now) the block file can't be
# named orderer.genesis.block or the orderer will fail to launch!
set -x
./bin/configtxgen -profile SampleMultiNodeEtcdRaft -configPath ${DIR_2}  -channelID byfn-sys-channel -outputBlock ${DIR_2}/genesis.block
res=$?
set +x
if [ $res -ne 0 ]; then
echo "Failed to generate orderer genesis block..."
exit 1
fi
