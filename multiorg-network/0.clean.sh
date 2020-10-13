
docker rm $(docker ps -aq)

echo "============CLEAN ORDERER DATA============"
cd orderers/
if [ -d orderers/org1/orderer1/data ]; then
rm -Rf orderers/org1/orderer1/data
rm -Rf orderers/org1/orderer1/etcdraft
fi

if [ -d orderers/org2/orderer1/etcdraft ]; then
rm -Rf orderers/org2/orderer1/data
rm -Rf orderers/org2/orderer1/etcdraft
fi

if [ -d orderers/org2/orderer2/etcdraft ]; then
rm -Rf orderers/org2/orderer2/data
rm -Rf orderers/org2/orderer2/etcdraft
fi

if [ -d orderers/org3/orderer1/etcdraft ]; then
rm -Rf orderers/org3/orderer1/data
rm -Rf orderers/org3/orderer1/etcdraft
fi

sleep 2


echo "============CLEAN PEER DATA============"
cd ../peers
if [ -d peer0/data ] || [ -d peer1/data ] || [ -d peer2/data ] || [ -d peer3/data ]; then
rm -rf peer*/data
fi

sleep 2

# echo "============CLEAN CHANNEL DATA============"
# cd ../
# if [ -d "channel-artifacts" ]; then
# rm -Rf channel-artifacts
# fi
