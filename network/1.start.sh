echo "============START!============"


echo "============GENERATE ORDERER CERTIFICATES============"
cd orderer/
sleep 1
./1-1.generate-cert.sh
cd ../
sleep 1

echo "============GENERATE PEER CERTIFICATES============"
cd peer/
./1-2.generate-cert.sh
cd ../
sleep 1

echo "============GENERATE CHANNEL ARTIFACTS============"
./1-3.generate-channel-artifacts.sh
