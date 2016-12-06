#!/usr/bin/env bash
################################ tx 服务 ################################ （说明：value可以是十六进制字符串、八进制字符串、十进制字符串或整数）
# 普通交易 SendTransaction
# 无request —— 用于Hyperchain测试, 时间戳单位是纳秒
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x80958818f0a025273111fba92ed14c3dd483caeb","timestamp":1478959217012956575,"value":53,"signature":"a68c4b492f020c2cc98bd15b64020a569365e1ed0a17cb38e8c472758f01dc2c2c60041e0ea90e6ffe3580d7629138c51f3927263dda23bc001e8e9bc368a0c900"}],"id":1}'
# 有request —— 用于Dashboard测试
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"4762d4f39099727dc58d1cec8d8f80f5c683a054","to":"0x0000000000000000000000000000000000000003","value":1,"timestamp":1478959217012956575,"signature":"test signature","request":10}],"id":1}'

# 根据交易hash查询交易信息 GetTransactionByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByHash","params":["0xe1404deb79b84d4303dd9ec25f53f0833975633eaf3016b10d12084b698479d5"],"id":1}'

# 根据区块hash和索引查询交易信息 GetTransactionByBlockHashAndIndex
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockHashAndIndex","params":["0x9e330e8890df02d22a7ade73b5060db6651658b676dc9b30e54537853e39c81d",0],"id":1}'

# 根据区块number和索引查询交易信息 GetTransactionsByBlockNumberAndIndex
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockNumberAndIndex","params":[1,0],"id":1}'

# 查询指定区块的交易数量
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getBlockTransactionCountByHash","params":["<block hash>"],"id":1}'

# 获取交易收据Receipt GetTransactionReceipt
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionReceipt","params":["<tx hash>"],"id":1}'

# 获取所有交易 GetTransactions
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactions","params":[{"from":5, "to":"latest"}],"id":1}'

# 获取所有失败交易 GetDiscardTransactions
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getDiscardTransactions","params":[],"id":1}'

# 获取签名哈希 GetSignHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getSignHash","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":"0x01"，"timestamp":1478959217012956575}],"id":1}'

# 获取链上的总交易量 GetTransactionsCount
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionsCount","params":[],"id":1}'

# 查询交易的平均处理时间 GetTxAvgTimeByBlockNumber
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTxAvgTimeByBlockNumber","params":[{"from":2,"to":8}],"id":1}'


################################ contract 服务 #################################
#　部署合约 DeployContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"contract_deployContract","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","payload":"0x60606040526000805463ffffffff1916815560ae908190601e90396000f3606060405260e060020a60003504633ad14af381146030578063569c5f6d14605e578063d09de08a146084575b6002565b346002576000805460e060020a60243560043563ffffffff8416010181020463ffffffff199091161790555b005b3460025760005463ffffffff166040805163ffffffff9092168252519081900360200190f35b34600257605c6000805460e060020a63ffffffff821660010181020463ffffffff1990911617905556","timestamp":1478959217012956575,"signature":"28e82952fd3264507de6a47189fb650846628a78324e064cae30931577b6edd90b9a6d1aaa41b9485873707de7cb123cb02521acb21b99174c42da1846cf772d01"}],"id":1}'

# 编译ABI CompileContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"contract_compileContract","params":["contract Accumulator{    uint32 sum = 0;   function increment(){         sum = sum + 1;     }      function getSum() returns(uint32){         return sum;     }   function add(uint32 num1,uint32 num2) {         sum = sum+num1+num2;     } }"],"id":1}'

# 调用合约方法 InvokeContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_invokeContract","params": [{"from": "<caller address>", "to": "<contract address>", "timestamp":1478959217012956575, "payload": "0x3ad14af300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002", "signature":"c2c1149d93f52d586b4ab9d9b634bf9b0221a2f6b78710d58bcbe417482884ca"}],"id": 1}'

# 获取合约code GetCode
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getCode","params": ["<contract address>","<block number>"],"id": 1}'

# 获取账户部署合约数量 GetContractCountByAddr
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getContractCountByAddr","params": ["0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","<block number>"],"id": 1}'

# 获取合约账户Storage GetStorageByAddr
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getStorageByAddr","params": ["<contract address>","latest"],"id": 1}'


################################ block 服务 ####################################
# 得到最新区块 LastestBlock
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_latestBlock","params":[],"id":1}'

# 得到所有区块 GetBlocks
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":1, "to":1}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":1, "to":5}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":5, "to":"latest"}],"id":1}'

# 根据区块hash查询区块信息 GetBlockByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlockByHash","params":["<block hash>"],"id":1}'

# 根据区块number查询区块信息 GetBlockByNumber, number可以是整数、十六进制字符串或者“latest”
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlockByNumber","params":[<block number>],"id":1}'

# 查询区块commit、batch平均时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryCommitAndBatchTime","params":[{"from":"the number of block", "to":"the number of block"}],id: 1}'

# 查询EVM平均处理时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryEvmAvgTime","params":[{"from":"the number of block","to":"the number of block"}],"id": 1}'


################################## account服务 ##################################
# 新建一个账户 NewAccount
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"account_newAccount","params":["123456"],"id": 1}'

# 解锁账户 UnlockAccount
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"account_unlockAccount","params":[{"address":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd", "password":"123"}],"id": 1}'

# 获取所有账户信息 GetAccounts
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"account_getAccounts","params":[],"id": 1}'

# 查询账户余额 GetBalance
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"account_getBalance","params":["0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"],"id": 1}'


################################# node服务 ######################################
# 得到节点信息
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_getNodeHash","params":[],"id": 1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_delNode","params":[{"nodehash":<nodehash>}],"id": 1}'


################################# 批量调用例子 ###################################
curl localhost:8081 --data '[{"jsonrpc":"2.0","method":"block_lastestBlock","params":[],"id":1}, {"jsonrpc":"2.0","method": "node_getNodes","id": 2}]'

