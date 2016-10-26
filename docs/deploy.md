# HyperChain 部署文档

`deploy.md` `Version 1.0.1`


## Quick Guide
Installation And Quick Start

1. clone the master branch code of hyperchain:

`git clone https://git.hyperchain.cn/hyperchain/hyperchain.git`

2. Build the binary 

`cd hyperchain && govendor build -o ./build`

3. Local run hyperchain test

`cd scripts && ./ubuntu.sh`

> Other Linux distributions:

Hyperchain officially support Ubuntu 16.04, if you are using other distribution Linux, you can do follow steps:

1. `cd hyperchain && mkdir build`
2. `govendor build -o build/hyperchain`
3. `cp -rf config build/`
4. `cd build && ./hyperchain -o 1 -l 8001 -t 8081 (if running first node)`
5. `cd build && ./hyperchain -o 2 -l 8002 -t 8082 (if running second node)`
6. `cd build && ./hyperchain -o 3 -l 8003 -t 8083 (if running third node)`
7. `cd build && ./hyperchain -o 4 -l 8004 -t 8084 (if running fourth node)`

    4.Keep those configuration files besides the hyperchain binary file :

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
│   │   ├── b18c8575e3284e79b92100025a31378feb8100d6
│   │   ├── e93b92f1da08f925bdee44e91e7768380ae83307
│   │   └── addresses
│   │                └── address
│   ├── local_peerconfig.json
│   ├── membersrvc.yaml
│   ├── pbft.yaml
│   └── peerconfig.json
└── hyperchain
```

if you want hyerchain node one by one, you can type this command:

```bash
cd build
./hyperchain -o 1 -l 8001 -t 8081 //run this on first node
./hyperchain -o 2 -l 8002 -t 8082 //run this on second node
./hyperchain -o 3 -l 8003 -t 8083 //run this on third node
./hyperchain -o 4 -l 8004 -t 8084 //run this on fourth node
```
Note: if you want run 4 nodes in localhost, modify the `-l` and `-t` options, ensure the ports are different. 

The configuration files are all contains in ‘config’ folder are the read initially while the project is started. Detailed introduction will be propose in next section.
## Configuration

There are some essential configurations which are include by config folder:

```bash 

 ls config
.
├── cert
│   ├── server
│   │   ├── tlsca.cert
│   │   └── tlsca.priv
│   ├── tls_peer.ca
│   ├── tls_peer.cert
│   └── tls_peer.priv
├── genesis.json
├── global.yaml
├── keystore
│   ├── 000f1a7a08ccc48e5d30f80850cf1cf283aa3abd
│   ├── 6201cb0448964ac597faf6fdf1f472edf2a22b89
│   ├── addresses
│   │   └── address
│   ├── b18c8575e3284e79b92100025a31378feb8100d6
│   └── e93b92f1da08f925bdee44e91e7768380ae83307
├── local_peerconfig.json
├── membersrvc.yaml
├── pbft.yaml
└── peerconfig.json
```

- `cert (folder)` This is the certification folder, which is used for transport layer security and permission access control. You do not need modify anything of this folder.

- `genesis.json` This file is predefined account of hyperchain, only if you are know how to generate the correctly.

- `global.yaml` The global config.

- `keystore (folder)`
This is the account private storage folder, DO NOT modify those files, only if you know how to generate the files. 

- `local_peerconfig.json`
This is the local peerconfig which is used for local test target. If you just run hyperchain locally. Do not modify this file. 

- `membersrvc.yaml`
This is the CA server config, *DO NOT* Modify this,only if you want change the output file path.

- `pbft.yaml`
This is PBFT core consensus algorithm, there are two positions need to modify:
    
    1. `“N”` is the node number which is actually participation the consensus process.
    
    2. `f`  is the flag which should suit the equation: “N=3*f+1”  

- `peerconfig.json`
This config file should generate automatically. You can generate this file by the script named “genpeerconfig.py”, that will be propose in next section.  

## Automatic Scripts
Hyperchain provides a dozen of automatic script, you can use this to accelerate the development speed.
The tree structure like below:  
```bash
.
├── command.sh
├── genpeerconfig.py
├── genyaml.py
├── innerserverlist.txt
├── mac.sh
├── serverlist.txt
├── server.py
├── server.sh
├── sub_scripts
│   ├── auto_test
│   │   ├── auto_local_test.js
│   │   ├── auto_server_test.js
│   │   └── primary_test.js
│   └── deploy
│       ├── server_addkey.sh
│       └── server_deploy.sh
├── ubuntu.sh
└── ubuntu_tmux.sh
```

You can use those scripts:

1.`genpeerconfig.py` : generate the peerconfig.json automatic, depends on serverlist.txt and innerserverlist.txt 

2.`genyaml.py` : useless current

3.`mac.sh` : local test for macos

4.`serverlist.txt` : paste the public IP address of server node (IMPORTANT: leave a blank line in the end of file)

5.`server.py` : useless current

6.`server.sh` : this script provide a auto test on the remote server node, depends on serverlist.txt and innerserverlist.txt.

7.`sub_scripts` (support scripts)

8.`sub_scripts/auto_test` (folder) : this folder contains the auto test scripts 

9.`sub_scripts/deploy` (folder) : this folder contains the deploy scripts which used for auto distribute hyperchain binary file on remote nodes.

10.`ubuntu.sh/ubuntu_tmux.sh` : used for local test, the former will open gnome-terminal and the poster will open tmux automatically. 

## Known Issues
- Wired memory leak
- Wired recover connection

Contributor:
- Chen Quan 2016-10-20