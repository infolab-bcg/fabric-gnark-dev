# Fabric chaincode verify gnark proof

## 单元测试

1. clone this repo

2. cd to gnarkverify 

```bash
cd fabric-gnark-dev/chaincode-go/gnarkverify
```

3. run unit test

```bash
go test -v 
```
## 链上测试

1. 运行fabric-samples test-network

参考[fabric快速开始](./doc/fabric_quickstart.md)，确保全部流程跑通。

2. clone 本仓库到fabric-samples根目录

```bash
cd fabric-samples
```

```bash
git clone https://github.com/infolab-bcg/fabric-gnark-dev.git
```

3. 启动网络

```bash
cd fabric-gnark-dev/verify-on-chain
```
```bash
./run.sh up
```

4. 部署链码

```bash
./run.sh deploy
```

5. 查询链码

```bash
./run.sh query
```

6. 生成proof

```bash
./run.sh generate
```

7. 调用链码验证proof

```bash
./run.sh verify
```

