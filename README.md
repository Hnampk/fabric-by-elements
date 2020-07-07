# fabric-by-elements

0: Xóa dữ liệu sổ cái + các file cấu hình liên quan đến channel

1: Sinh các certificate của orderer, peer và cấu hình channel

2: Khởi chạy orderer trên port 7050

3: Khởi chạy peer trên port 7051

4: Tạo channel dựa theo file cấu hình configtx.yaml => sinh ra file mychannel.block

5: Join peer vào channel sử dụng file mychannel.block

6: Update anchor peer

7: Đóng gói chaincode (Từ version 2.0)

8:  .1 Cài đặt chaincode lên peer, Kiểm tra lại việc cài có thành công hay không bằng command "../bin/peer lifecycle chaincode queryinstalled"
    .2 Approve chaincode đã cài
