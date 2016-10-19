#HyperChain 部署文档

`deploy.md` `Version 1.0`


## Quick Guide
这是快速启动`hyperchain`的步骤：
1. 请先将数据库文件进行清理：
```bash
rm -rf /tmp/hyperchain/
```
2. 将编译好的二进制文件，配置文件以及证书文件`cp`到你想要部署的目录下：
```bash
.
├── cert
│   ├── server
│   │   ├── tlsca.cert
│   │   └── tlsca.priv
│   ├── tls_peer.ca
│   ├── tls_peer.cert
│   └── tls_peer.priv
├── genesis.json
├── hyperchain （binary execute file）
├── keystore
├── logs
├── membersrvc.yaml
├── pbft
│   └── config.yaml
└── peerconfig.json
```
3. 请修改peerconfig.json 文件：
```json
{
  //系统包括的节点数
  "MAXPEERS":"4",
  //系统共识信息传输端口
  "port1":"8001",
  "port2":"8001",
  "port3":"8001",
  "port4":"8001",
  //系统共识信息传输ip
  "node1":"10.105.75.5", // 将此4行修改为你需要连接的服务器ip地址
  "node2":"10.105.89.219",
  "node3":"10.105.85.32",
  "node4":"10.105.80.128"
  }
```
3. 在相应的服务器上依次运行如下命令：
第一个节点：
`./hyperchain -o 1 -l 8081 -p ./peerconfig.json -f ./pbft/ -g ./genesis.json`
第二个节点：
`./hyperchain -o 2 -l 8081 -p ./peerconfig.json -f ./pbft/ -g ./genesis.json`
第三个节点：
`./hyperchain -o 3 -l 8081 -p ./peerconfig.json -f ./pbft/ -g ./genesis.json`
第四个节点：
`./hyperchain -o 4 -l 8081 -p ./peerconfig.json -f ./pbft/ -g ./genesis.json`
4. 完成
当你看到所有节点的输出日志中有：
```bash
┌────────────────────────────┐
│  All NODES WERE CONNECTED  │
└────────────────────────────┘
```
就说明所有的节点已经连接完毕，你可以通过`jsonrpc`服务调用`hyperchain`提供的接口了。


## 二进制文件
`hyperchain`的二进制文件在ubuntu 16.04 (amd_64)系统中采用`go 1.6.3` 进行编译，无法在`macOS`,`win32/64`
系统中运行。
如果您需要在上述平台中运行`hyperchain`，请使用`docker`。


## 启动命令
在当前版本中，基本的参数包括：
```bash
./hyperchain -h
Options:
   // 显示帮助信息
  -h, --help                                  display help information
  //当前需要启动的节点id
  -o, --nodeid[=1]                            current node ID
  //grpc 数据交换端口
  -l, --localport[=8001]                      gRPC data transport port
  // peerconfig.json 配置文件的路径，请指定完整路径
  -p, --peerconfig[=./peerconfig.json]        peerconfig.json path
  // pbft 共识算法配置文件*文件夹*的路径，请确保该文件夹下有config.yaml文件
  -f, --pbftconfig[=./consensus/pbft/]        pbft config file folder path,ensure this is a validate dir path
  // genesis.json 配置文件的路径，请指定完整路径
  -g, --genesisconfig[=./core/genesis.json]   genesis.json config file path
```
- -h,--help : 显示帮助信息
- -o,--nodeid : 指定当前需要启动的节点的id (要求整数|默认值为 `1`)
- -l (L),--localport : hypechain系统共识数据交换所需的端口 （请指定合法端口|默认`8001`）
- -p,--peerconfig : `peerconfig.json`文件所在的完整路径
- -f,--pbftconfig : `pbft`算法相关的配置文件，请指定目录路径，要求该目录下包含`config.yaml`文件
- -g,--genesisconfig : `genesis.json`配置文件所在路径，请指定peerconfig.json 的完整路径

*启动命令示例：*
`./hyperchain -o 1 -l 8081 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json`

> 说明：
> 当前版本 `hyperchain` 还不支持以`Deamon`方式运行，

## 目录结构
```
.
├── cert
│   ├── server
│   │   ├── tlsca.cert
│   │   └── tlsca.priv
│   ├── tls_peer.ca
│   ├── tls_peer.cert
│   └── tls_peer.priv
├── genesis.json
├── hyperchain （binary execute file）
├── keystore
├── logs
├── membersrvc.yaml
├── pbft
│   └── config.yaml
└── peerconfig.json
```

请确保上述文件结构完整，并且通过命令：
`./hyperchain -o 1 -l 8081 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json`
运行 `hyperchain` 的第一个节点

## 配置文件说明
### peerconfig.json
在本版本中，节点信息需要完全包含在`peerconfig.json`中，系统需要依赖该文件完成全连接。

*以四节点系统为例*
`peerconfig.json:`
```json
{
  //系统包括的节点数
  "MAXPEERS":"4",
  //系统共识信息传输端口
  "port1":"8001",
  "port2":"8001",
  "port3":"8001",
  "port4":"8001",
  //系统共识信息传输ip
  "node1":"10.105.75.5",
  "node2":"10.105.89.219",
  "node3":"10.105.85.32",
  "node4":"10.105.80.128"
  }
```

> 说明：系统需要依赖该配置文件实现节点之间的项目连接，请确保ip列表中的节点能够直接连接。
> 此处的端口为gRPC的信息交换端口，让防火墙允许该TCP协议通过该端口进行连接。

### genesis.json
在本版本中，该配置文件主要包括了创世块的账户信息，需要确保该文件中的账户地址信息和余额信息有效。
`genesis.json`

```json
{
  "test1": {
    "Timestamp": 122222,
    "alloc": {
      "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd": 1000000000,
      "e93b92f1da08f925bdee44e91e7768380ae83307": 1000000000,
      "6201cb0448964ac597faf6fdf1f472edf2a22b89": 1000000000,
      "b18c8575e3284e79b92100025a31378feb8100d6": 1000000000
    },
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "BlockHash": "0x00012",
    "Number": 0
  }
}
```
> 说明：本文件在默认情况之下不允许修改，请直接放在相应目录即可
> 高级：alloc值信息请您知道账户生成方式之后再进行修改。

### pbft/config.yaml

本文件用于共识算法的配置：
```yaml
	################################################################################
	#
	#   PBFT PROPERTIES
	#
	#   - List all algorithm-specific properties here.
	#   - Nest keys where appropriate, and sort alphabetically for easier parsing.
	#
	################################################################################
general:

    # Operational mode: currently only batch ( this value is case-insensitive)
    mode: batch

    # Maximum number of validators/replicas we expect in the network
    # Keep the "N" in quotes, or it will be interpreted as "false".
    "N": 4 // 如果增加节点，请修改此处的n 需要确保 n = 3f+1， 请保留 ‘N’上的双引号

    # Number of byzantine nodes we will tolerate
    f: 1 //此处f 需要满足与N的关系 N=3*f+1

    # Checkpoint period is the maximum number of pbft requests that must be
    # re-processed in a view change. A smaller checkpoint period will decrease
    # the amount of time required to recover from an error, but will decrease
    # overall throughput in normal case operation.
    K: 10

    # Affects the receive log size which is K * logmultiplier
    # The primary will only send sequence numbers which fall within K * logmultiplier/2 of
    # its high watermark, so this cannot be set to less than 2
    # For high volume/high latency environments, a higher log size may increase throughput
    logmultiplier: 4

    # How many requests should the primary send per pre-prepare when in "batch" mode
    batchsize: 100

    # Whether the replica should act as a byzantine one; useful for debugging on testnets
    byzantine: false

    # After how many checkpoint periods the primary gets cycled automatically.  Set to 0 to disable.
    viewchangeperiod: 0

# Timeouts
timeout:

    # Send a pre-prepare if there are pending requests, batchsize isn't reached yet,
    # and this much time has elapsed since the current batch was formed
    batch: 1s

    # How long may a request take between reception and execution, must be greater than the batch timeout
    request: 3s

    # How long may a validate Txs process will take by outter
    validate: 1s

    # How long may a view change take
    viewchange: 4s

    # How long to wait for a view change quorum before resending (the same) view change
    resendviewchange: 10s

    # Interval to send "keep-alive" null requests.  Set to 0 to disable. If enabled, must be greater than request timeout
    nullrequest: 4s

    # How long to wait for N-f responses after send negotiate view
    negoView: 6s

	################################################################################
	#
	#   SECTION: EXECUTOR
	#
	#   - This section applies to the distinct executor service
	#
	################################################################################
executor:

    # The queue size for execution requests, ordering proceeds and queues execution
    # requests.  This value should always exceed the pbft log size
    queuesize: 30

```

### membersrvc.yaml

该文件为CA证书证书读取与生成文件，其中的几个路径需要进行说明：

`membersrvc.yaml:`
```yaml
server:
        # current version of the CA
        version: "0.1"

        # limits the number of operating system threads used by the CA
        # set to negative to use the system default setting
        gomaxprocs: -1

		# 本路径用于生成CA相应的配置文件的根目录，请确保该目录具有可写权限
		rootpath: "/tmp/hyperchain/production"

		# 该路径相对于上述的rootpath 主要用于生成相应的ca证书文件，如果不存在该文件夹，将会自动创建
        cadir: ".membersrvc"

		# 本目录将会自动加载到下列文件路径之前，请确保您需要这个路径，否则请将其设置为./
        caserverdir: "./"

        # port the CA services are listening on
		# CA servers 正在监听的端口

        port: ":50051"
        paddr: localhost:50051

        # TLS certificate and key file paths
		# TLS 证书和秘钥存储的路径，再运行过程中，将会自动生成相应的证书文件，主要应用于传输层安全的CA认证
        tls:
            cert:
                file: "cert/server/tlsca.cert"
            key:
                file: "cert/server/tlsca.priv"
node:
        tls:
            cert:
                file: "cert/tls_peer.cert"
            key:
                file: "cert/tls_peer.priv"
            cap:
                file: "cert/tls_peer.ca"
		# 此处用于辨识CA服务，可以自定义
        serverhostoverride: "hyperchain"

	#　请确保在本行之下的配置文件内容不被修改！
```

> 说明：`tls`证书需要外部提供，本系统提供的仅仅是内部签出的证书，请不要再生产系统中进行使用。


## 备注
ChangeLog:
- 添加了基本内容，涵盖启动方式，配置文件的配置说明，二进制文件说明等

Contributor:
- Chen Quan 2016-10-18