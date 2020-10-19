Các script dùng để start một cụm blockchain gồm 02 org:

    (1) org1: gồm 01 orderer chạy mode Solo, 02 peers
    (2) org2: gồm 02 peers

    - 02 org này sẽ được cài đặt trên 02 server khác nhau
    - 1 channel được tạo ra gồm cả 04 peer này
    - peer1.org1 và peer1.org2 sẽ được cài đặt chaincode
    - 02 chaincode được cài đặt sẽ trỏ đến 02 accounting service khác nhau bằng cách gửi tx đến peer cài chaincode tương ứng


- Bước 1: Cài đặt file hosts ánh xạ địa chỉ các peers + orderer:
    Ví dụ:
        10.22.7.235     orderer0.example.com
        10.22.7.235     peer0.org1.example.com
        10.22.7.235     peer1.org1.example.com

        10.22.7.236     peer0.org2.example.com
        10.22.7.236     peer1.org2.example.com

- Bước 2: Cập nhật địa chỉ orderer và org trong file .env
    Ví dụ
        ORDERER_ADDRESS=10.22.7.235:7050
        ORG1_IP=10.22.7.235
        ORG2_IP=10.22.7.236

- Bước 3: với orderer solo mode, nếu gặp vấn đề phân giải IP, hardcode địa chỉ orderer trong file channel-artifacts/configtx.yaml. Ví dụ:

        OrdererEndpoints:
            - orderer0.example.com:7050 
        
    =>  OrdererEndpoints:
            - 10.22.7.235:7050

- Bước 4: chạy lần lượt các script ở 02 server tương ứng với 02 org. Ví dụ: 

        - [235-org1] 0.clean.sh
        - [235-org1] 1.generate-certificates.sh
        - [235-org1] 2.generate-channel-artifacts.sh
        - [235-org1] 3.start-orderers.sh
        - [235-org1] 4.start-peer0-org1.sh
        - [235-org1] 4.start-peer1-org1.sh
        - [235-org1] 5.create-and-join-channel-org1.sh
        - [235-org1] copy folder channel-artifacts sang vị trí tương ứng ở [236]

        - [236-org2] 4.start-peer0-org2.sh
        - [236-org2] 4.start-peer1-org2.sh
        - [236-org2] 6.join-channel-org2.sh

        - [235-org1] 7.install-chaincode-org1.sh
        - [235-org1] copy folder peers/packed-cc sang vị trí tương ứng ở [236]

        - [236-org2] 7.install-chaincode-org2.sh

        - [235-org1] 8.commit-chaincode-org1.sh

    Đến đây, mạng blockchain đã được dựng, các peer1 đã được cài đặt chaincode

- Bước 5: cập nhật địa chỉ của Accounting service được trỏ đến bới chaincode. Khi này, 02 chaincode sẽ gọi đến 02 Accounting service khác nhau

        - [235-org1] 9.update-accounting-service-endpoint-org1.sh
        - [236-org2] 9.update-accounting-service-endpoint-org2.sh
