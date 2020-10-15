# Generate orderers certificates
# Generate peers certificates

echo
echo "##########################################################"
echo "### Generate orderers certificates using cryptogen tool ##"
echo "##########################################################"

if [ -d orderers/crypto-config ]; then
    rm -Rf orderers/crypto-config
fi

sleep 1

set -x
../bin/cryptogen generate --config=./orderers/crypto-config-orderer.yaml --output=./orderers/crypto-config
res=$?
set +x

if [ $res -ne 0 ]; then
echo "Failed to generate orderers certificates..."
sleep 5
exit 1
fi


echo
echo "##########################################################"
echo "### Generate peers certificates using cryptogen tool #####"
echo "##########################################################"

if [ -d peers/crypto-config ]; then
rm -Rf peers/crypto-config
fi

sleep 1

set -x
../bin/cryptogen generate --config=./peers/crypto-config-peer.yaml --output=./peers/crypto-config
res=$?
set +x

if [ $res -ne 0 ]; then
echo "Failed to generate peers certificates..."
sleep 5
exit 1
fi