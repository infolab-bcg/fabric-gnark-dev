# Hyperledger Fabric Quick Start

## 准备工作

### git, curl, jq

```shell
sudo apt install git curl jq
```

### docker

#### 安装

```shell
sudo apt install docker.io docker-compose
```

#### 普通用户运行权限

大多数情况下在安装时，docker 用户组已经创建，这一步基本不需要

```yaml
sudo groupadd docker
```

**添加到 docker 用户组**

```shell
sudo usermod -aG docker $USER
```

**开机启动**

```shell
sudo systemctl enable --now docker
```

#### docker pull 代理设置

```shell
sudo mkdir -p /etc/systemd/system/docker.service.d
```

新建文件并编辑：

```shell
sudo vim /etc/systemd/system/docker.service.d/http-proxy.conf
```

写入以下内容：

```shell
[Service]
Environment="HTTP_PROXY=http://xx.xx.xx.xx:7890"
Environment="HTTPS_PROXY=https://xx.xx.xx.xx:7890"
```

代理生效：

```shell
sudo systemctl daemon-reload
sudo systemctl restart docker
```

更多设置参考[官方文档](https://docs.docker.com/engine/daemon/proxy/#environment-variables)


### go

1. 首先确保当前系统中没有其他旧版本 go：

```shell
which go
```

2. 下载最新版 golang 二进制文件

```shell
VERSION=$(curl -s https://go.dev/dl/ | grep -oP 'go[0-9]+\.[0-9]+\.[0-9]+' | head -1) && wget "https://go.dev/dl/${VERSION}.linux-amd64.tar.gz"
```

3. 卸载旧版本 golang, 若之前没有安装 golang，可跳过

```shell
sudo rm -rf /usr/local/go
```

4. 将压缩文件解压到 `/usr/local`

```shell
GOFILE=$(ls | grep -P 'go[0-9]+\.[0-9]+\.[0-9]+') && sudo tar -C /usr/local/ -xzf ${GOFILE}  
```

5. 添加环境变量

编辑 `~/.zshrc`​,追加以下内容并运行 `source ~/.zshrc`​ 使其生效

```shell
export PATH=$PATH:/usr/local/go/bin
```

6. 最后运行 `go version` ​ 命令验证安装是否成功

#### 设置 golang 镜像

目前一直在用七牛云的 golang 镜像,体验还算不错

```shell
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

运行 `go env`​ 命令查看设置是否生效。

‍

## 安装环境

```shell
curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
```

如果下载失败，可能是代理问题，设置终端代理即可。

下载docker镜像，clone fabric-samples仓库，下载二进制文件，官方脚本一键搞定

```shell
./install-fabric.sh docker samples binary
```

## 运行测试网络

```shell
cd fabric-samples/test-network
./network.sh down
```

会拉取一个镜像，暂时用不到

### 启动网络

```shell
./network.sh up
```

### 创建通道

```shell
./network.sh createChannel
```

### 部署链码

在这之前，先确定golang安装成功且配置goproxy

```shell
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
```

### 与网络交互

> 如果遇到错误，以[官方文档](https://hyperledger-fabric.readthedocs.io/en/release-2.5/test_network.html#interacting-with-the-network)为准

使用`peer`​与网络交互，`peer`​命令可实现：

* 调用已部署的智能合约，
* 更新通道
* 安装和部署新的智能合约

#### 设置基础变量

```shell
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
```

#### 设置 org1 的环境变量

```shell
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
```

#### 初始化账本

```shell
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"InitLedger","Args":[]}'
```

命令成功时输出

```shell
-> INFO 001 Chaincode invoke successful. result: status:200
```

#### 查询账本

```shell
peer chaincode query -C mychannel -n basic -c '{"Args":["GetAllAssets"]}'
```

命令成功时输出

```shell
[
  {"ID": "asset1", "color": "blue", "size": 5, "owner": "Tomoko", "appraisedValue": 300},
  {"ID": "asset2", "color": "red", "size": 5, "owner": "Brad", "appraisedValue": 400},
  {"ID": "asset3", "color": "green", "size": 10, "owner": "Jin Soo", "appraisedValue": 500},
  {"ID": "asset4", "color": "yellow", "size": 10, "owner": "Max", "appraisedValue": 600},
  {"ID": "asset5", "color": "black", "size": 15, "owner": "Adriana", "appraisedValue": 700},
  {"ID": "asset6", "color": "white", "size": 15, "owner": "Michel", "appraisedValue": 800}
]
```

#### 调用链码

```shell
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"TransferAsset","Args":["asset6","Christopher"]}'
```

命令成功时输出

```shell
[chaincodeCmd] chaincodeInvokeOrQuery -> INFO 001 Chaincode invoke successful. result: status:200
```

接下来，我们尝试使用 org2 的身份来查询上述改动是否已经生效

#### 设置 org2 的环境变量

```shell
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051
```

#### 再次查询账本

```shell
peer chaincode query -C mychannel -n basic -c '{"Args":["ReadAsset","asset6"]}'
```

成功时输出

```shell
{"ID":"asset6","color":"white","size":15,"owner":"Christopher","appraisedValue":800}
```

结果显示  `"asset6"`​  转给了 Christopher

至此，测试网络的最简单实践已经全部完成！

### 关闭网络

```shell
./network.sh down
```
