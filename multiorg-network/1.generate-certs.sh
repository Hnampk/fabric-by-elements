echo "============START!============"


echo "============GENERATE ORDERER CERTIFICATES============"
cd orderers/
sleep 1
./1-1.generate-cert.sh
cd ../
sleep 1

echo "============GENERATE PEER CERTIFICATES============"
cd peers/
./1-2.generate-cert.sh
cd ../
sleep 1

echo "============GENERATE CHANNEL ARTIFACTS============"
./1-3.generate-genesis-block.sh
