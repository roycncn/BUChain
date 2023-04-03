# BUChain
BLOCKCHAIN Project
```
22428682 HUANG Weiyi
22435956 LIU Bicheng
22443606 WU Chen
22407677 YAO Ruilin
22467610 ZHANG Yatao
```


#Running
```
cd .cmd/chaind
go build
./chaind start -f ../../config.yaml 
```

# 分工
- App Framework  : Roy
- Blockchain - Blockchain Prototype: 怡姐
- Blockchain - Mining, Consensus: BC
- Blockchain - TX，UTXO，Mint Coin: 琛哥 
- Network - Client 和 Server: 老园

# Test PK - Miner
8001 priv: 20674f94d76abbeaea279b356e0fba8a9cf8c9e415d2dc13ad34b2fe3a2dbff4

8001 Pub: 039d95813a47234fe889729ff94efa4bb170e4190ba4157f837a68621458018638

8002 priv: 1c9da4ab0f90cf07f8f016d4a69d2df2d15918b877df911e5da62baf33584d2c

8002 Pub: 02e90c589d434fe3b70f11bea48bf54922c46646a82d4418d3d9ed16f258a7f88b

8003 priv: 68cc1a46d5d7af78e000b0780be910442789426c78ba207ef586a1681bd5e48a

8003 Pub: 03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4

#Test PK - Wallet 

A priv: 14f54846ccd49322305a44b4202ebb2d85d99aa321e1ff1a4d7026d876362db6

A Pub: 03a45af63249565ddd4185bc228e838b5d4d13f69e324abdda91b9bd538d80cc37

B priv: 73ae44e014506847a2490808510d4db30f8c585b2cbe3e74ea89f05aab5e3722

B Pub: 03bb1659d5cadb49b4d2ea88fe965e405e7316b10a7bb802824a0b2fb131ea104c

C priv: 1d020987ad4d22448595f4f422ce13dbddefce7d0a036558c25fd2b5bc0cd6a9

C Pub: 03d33caa21f026cfbd54a7877bec29227cc9926b557ebe67f73b696423a94fb90a



#CURL测试用例
POST 转账请求
```
curl --location --request POST 'http://127.0.0.1:8001/wallet' \
--header 'Content-Type: application/json' \
--data-raw '{
"FromAddr": "039d95813a47234fe889729ff94efa4bb170e4190ba4157f837a68621458018638",
"ToAddr": "03a45af63249565ddd4185bc228e838b5d4d13f69e324abdda91b9bd538d80cc37",
"Amount": 50,
"PrivKey": "20674f94d76abbeaea279b356e0fba8a9cf8c9e415d2dc13ad34b2fe3a2dbff4"
}'
```
