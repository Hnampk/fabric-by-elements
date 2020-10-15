Các script dùng để start một cụm blockchain gồm 02 org:
    (1) org1: gồm 01 orderer chạy mode Solo, 02 peers
    (2) org2: gồm 02 peers

    - peer1.org1 và peer1.org2 sẽ được cài đặt chaincode
    - 02 org này sẽ được cài đặt trên 02 server khác nhau
    - 1 channel được tạo ra gồm cả 04 peer này
    - 02 chaincode được cài đặt sẽ trỏ đến 02 accounting service khác nhau bằng cách gửi tx đến peer cài chaincode tương ứng
