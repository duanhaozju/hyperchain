# HYPERCHAIN

>HyperChain Version 1.0

## CHANGE LOG:

- version 1.0 Release 2016/10/19

## STABLE BRANCH

`master`


## LATEST BRANCH

`EVM`

## LATEST TAG

`Release_1.0_20161019`

## Product Description

Hyperchain is a foundation consortium blockchain platform that meets the needs of industrial applications. Hyperchain integrates high-performance and reliable consensus algorithm and is compatible with the open source community's smart contract development language and execution environment. 

Hyperchain strengthens the key features such as accounting authorization mechanism and transaction data encryption, while providing a powerful visual Web management console for efficient management of blockchain nodes, ledgers, transactions and smart contracts.
The Hyperchain consortium blockchain service platform provides enterprise-level blockchain networking solutions for enterprise, government, and industry alliance. Support enterprises based on the existing cloud platform in rapid deployment, expansion configuration and management blockchain network, also the blockchain network running status of real-time visual monitoring. The standard specification is fully compatible with the existing blockchain open source community of smart contract mode and application programming interface, full supporting of the blockchain network common functions. Also importantly, we improved of the consensus mechanism, access mechanisms and security mechanisms at the bottom. Hyperchain can be used by businesses and organizations as a reliable foundation for block chain industry application solutions. 

## Features of Hyperchain

- Consensus on block coherence based on RBFT(Robust Byzantine Fault Tolerance)
- Limit the entry of accounting node base on certificate authority.
- Encryption of transmission data base on elliptic curve diffie-hellman.
- Digital signature on transaction based on elliptic curves.
- Validate the signature of transaction.
- Query balance of the account.
- Query block information.
- Deploy smart contract.
- Run smart contract.
- Query information of smart contract.
- Query information of peers in the blockchain..
- Query transaction information.
- Save balance, transaction, receipt by Merkle Patricia Tree.

## Installation And Quick Start

clone the master branch code of hyperchain:

`git clone https://git.hyperchain.cn/hyperchain/hyperchain.git`

`cd hyperchain`

`govendor build -o ./build`

with the necessary config files:

```bash
ls build
--------------
.
├── config
│   ├── cert
│   │   ├── server
│   │   │   ├── tlsca.cert
│   │   │   └── tlsca.priv
│   │   ├── tls_peer.ca
│   │   ├── tls_peer.cert
│   │   └── tls_peer.priv
│   ├── genesis.json
│   ├── global.yaml
│   ├── keystore
│   │   ├── 000f1a7a08ccc48e5d30f80850cf1cf283aa3abd
│   │   ├── 6201cb0448964ac597faf6fdf1f472edf2a22b89
│   │   ├── addresses
│   │   │   └── address
│   │   ├── b18c8575e3284e79b92100025a31378feb8100d6
│   │   └── e93b92f1da08f925bdee44e91e7768380ae83307
│   ├── local_peerconfig.json
│   ├── membersrvc.yaml
│   ├── pbft.yaml
│   └── peerconfig.json
└── hyperchain

```
you can run the local test just type:

`cd script && ./ubuntu.sh`

if you want hyerchain node one by one, you can type this command:

```bash
cd build
./hyperchain -o 1 -l 8001 -t 8081 //run this on first node
./hyperchain -o 2 -l 8001 -t 8081 //run this on second node
./hyperchain -o 3 -l 8001 -t 8081 //run this on third node
./hyperchain -o 4 -l 8001 -t 8081 //run this on fourth node
```
Note: if you want run 4 nodes in localhost, modify the `-l` and `-t` options, ensure the ports are different. 


## Configuration

### `config/global.yaml`

This configuration file contains the mainly config of hyperchain,generally, DO NOT need to modify this file.

IMPORTANT : if you just want run Hyperchain locally, just type:

    `cd scripts && ./ubuntu.sh`
if you need run hyperchain at remote server,please ensure the key: `global.configs.peers` matchs value:
`config/peerconfig.json` not the `config/local_peerconfig.json`

```yaml
# node config

global:

##########################
# data storage config #
##########################
    account:
        keystoredir: "./build/keystore"
        keynodesdir: "./build/keynodes"

    logs:
        # dump the log file or not
        dumpfile: false
        logsdir: "./build/logs"
        # loglevel contains those levels:
        #    High   CRITICAL
        #           ERROR
        #           WARNING
        #           NOTICE
        #           INFO
        #    Low    DEBUG

        loglevel: "NOTICE"
    database:
        dir: "./build/database/"
###########################
# configs
############################
    configs:
        peers: "config/local_peerconfig.json"
        genesis: "config/genesis.json"
        membersrvc: "config/membersrvc.yaml"
        pbft: "config/pbft.yaml"
```

### `configs/peerconfig.json`

Hyperchain version 1.0 not support the add or delete nodes, so need to add all nodes infos into the `peerconfig.json` 
actually. you can use the auto generate script named `scripts/genpeerconfig.py` to generate the peerconfig.json.
if you need do this, add the node ip info into `scripts/serverlist.txt` and `scripts/innerserverlist.txt`.
put the external ip address into `scripts/serverlist.txt`, and put the internal ip address into `scripts/innerserverlist.txt`.

then generate `peerconfig.json` just type:

`python genpeerconfig.py`

> peerconfig.json example

```json
{
    "nodes": [
        {
            "external_address": "182.254.220.52",
            "rpc_port": 8081,
            "port": 8001,
            "id": 1,
            "address": "10.105.8.145"
        },
        {
            "external_address": "115.159.26.15",
            "rpc_port": 8081,
            "port": 8001,
            "id": 2,
            "address": "10.105.2.6"
        },
        {
            "external_address": "115.159.212.225",
            "rpc_port": 8081,
            "port": 8001,
            "id": 3,
            "address": "10.105.10.74"
        },
        {
            "external_address": "115.159.213.69",
            "rpc_port": 8081,
            "port": 8001,
            "id": 4,
            "address": "10.105.36.220"
        }
    ],
    "maxpeernode": 4
}
```

Notes:
    the `maxpeernode` key's number depends on the `serverlist.txt` lines, this number means the hyperchain max support nodes number.
    if you want expand the maxnumber, you should add the node ip into the `serverlist.txt`, not edit this file dirctily.
    meanwhile, you should modify the `config/pbft.yaml`

## `config/pbft.yaml`

if you want to expand the max node numbers, please modify the `“N”` to your target number, and the `f` should be meet the equation: `N=3*f+1`. DONOT modify other content of this file, if you dont know the means clearly.
```yaml
general:

    # Operational mode: currently only batch ( this value is case-insensitive)
    mode: batch

    # Maximum number of validators/replicas we expect in the network
    # Keep the "N" in quotes, or it will be interpreted as "false".
    "N": 4

    # Number of byzantine nodes we will tolerate
    f: 1

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

...
```

## Known Issues

## Documentation
Hyperchain documentation is available on the Hyperchain Documentation website https://git.hyperchain.cn/hyperchain/hyperchain.Updated documentation will be available on that page when any documents are updated post-release.
Hyperchain has the following product documentations:

- White Paper
- Product Introduction Document
- Application Programming Interface Document
- Performance Test Report


## Troubleshooting and Getting Help

Contacting Technical Support:

Before you contact our technical support staff, have the following information available.

- Your name, title, company name, phone number, and email address

- Operating system and version number

- Product name and release version

- Problem description

Hours: 9:00 AM to 5:00 PM PST (Monday-Friday, except Holidays)

Phone: 0571-81180102, 0571-81180103

Email: support@hyperchain.cn