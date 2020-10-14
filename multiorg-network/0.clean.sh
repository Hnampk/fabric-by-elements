
docker rm $(docker ps -aq)

echo "============CLEAN ORDERER DATA============"
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
if [ -d peers/org1/peer0/data ] || [ -d peers/org1/peer1/data ]; then
rm -rf peers/org1/peer*/data
fi
if [ -d peers/org2/peer0/data ] || [ -d peers/org2/peer1/data ] || [ -d peers/org2/peer2/data ] || [ -d peers/org2/peer3/data ]; then
rm -rf peers/org2/peer*/data
fi
if [ -d peers/org3/peer0/data ] || [ -d peers/org3/peer1/data ]; then
rm -rf peers/org3/peer*/data
fi
