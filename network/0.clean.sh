
echo "============CLEAN ORDERER DATA============"
cd orderer/
if [ -d "data" ]; then
rm -Rf data
fi

if [ -d "etcdraft" ]; then
rm -Rf etcdraft
fi

sleep 2


echo "============CLEAN PEER DATA============"
cd ../peer
if [ -d "data" ]; then
rm -Rf data
fi

sleep 2

echo "============CLEAN CHANNEL DATA============"
cd ../
if [ -d "channel-artifacts" ]; then
rm -Rf channel-artifacts
fi
