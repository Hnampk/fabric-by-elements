Một số rule trong mạng multiorg-network

1. các peer chẵn (peer0, peer2) của org2 thực hiện tạo channel
2. chỉ các peer lẻ (peer1, peer3) mới cài đặt chaincode
3. Các peer chẵn không cài đặt chaincode, chỉ đảm nhận việc sync transaction đóng block
4. API được cài đặt trên cùng server với peer chẵn, sẽ lắng nghe việc đóng block từ các peer này

5. channel vnpay-channel-2 dùng để test sharding chaincode - Accounting Service, với chaincode trên peer3 (236) được connect tới Accounting service 232
6. service custom-sdk/server-236-with-org3.go được hardcode để chạy test sharding vnpay-channel-2  (T_T)

Luồng chạy test sharding:
1. generate certificates
2. start orderer2.org2 và orderer1.org3
3. generate channel-artifacts (1-4)
4. start các peer: peer2.org2, peer3.org2, peer0.org3, peer1.org3
5. create channel ở peer2.org2
6. join channel ở peer2.org2
7. copy file vnoay-channel-2.block từ peer2.org2 cho các peer khác
8. join channel cho các peer còn lại
9. cài đặt chaincode ở peer3.org2 (bỏ qua commit nếu cần nhiều đồng thuận - check file configtx.yaml > Application > Policies > Endorsement)
10. copy file nén chaincode trong folder peers/org2/peer3/packed-cc/* cho peer1.org3
11. cài đặt chaincode ở peer1.org3, thực hiện commit
