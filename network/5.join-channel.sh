source ./peer-env.sh

set -x
../bin/peer channel join -b ${CHANNEL_NAME}.block

echo "============Wait for join channel response========="
sleep 2

../bin/peer channel list
set +x

