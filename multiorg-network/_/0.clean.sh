# Clean orderer data
# Clean peer data


echo
echo "##########################################################"
echo "### Clean orderers data ##################################"
echo "##########################################################"

if [ -d orderers/data ]; then
    rm -Rf orderers/data
fi

if [ -d orderers/etcdraft ]; then
    rm -Rf orderers/etcdraft
fi

sleep 1


echo
echo "##########################################################"
echo "### Clean peers data #####################################"
echo "##########################################################"

if [ -d peers/data0 ]; then
    rm -Rf peers/data0
fi

if [ -d peers/data1 ]; then
    rm -Rf peers/data1
fi

if [ -d peers/packed-cc ]; then
    rm -Rf peers/packed-cc/*
fi

