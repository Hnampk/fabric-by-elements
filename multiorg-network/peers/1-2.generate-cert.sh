echo
echo "##########################################################"
echo "##### Generate certificates using cryptogen tool #########"
echo "##########################################################"

if [ -d "crypto-config" ]; then
rm -Rf crypto-config
fi

sleep 2 

set -x
../bin/cryptogen generate --config=./crypto-config-org1.yaml
res=$?
../bin/cryptogen generate --config=./crypto-config-org2.yaml
res=$?
../bin/cryptogen generate --config=./crypto-config-org3.yaml
res=$?
set +x

if [ $res -ne 0 ]; then
echo "Failed to generate certificates..."
exit 1
fi