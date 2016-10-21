#!/usr/bin/env bash
########### tx 服务 ########## （说明：value可以是十六进制字符串、八进制字符串、十进制字符串或整数）
# 普通交易 SendTransaction
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":"0x01"}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":1,"signature":"test signature"}],"id":1}'

# 测试：传送私钥 SendTransactionTest
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransactionTest","params":[{"from":"0xdac4c37ca97e3c025d7b3580c4a4a4adc096eebf","to":"0x0000000000000000000000000000000000000003","value":1,"privKey":"c2c1149d93f52d586b4ab9d9b634bf9b0221a2f6b78710d58bcbe417482884ca"}],"id":1}'

# 根据交易hash查询交易信息 GetTransactionByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByHash","params":["0xe17cdfcbb5e6825a41a823598ba2acaf1ae48b45513aa1ba4e2844bc75ea9e81"],"id":1}'

# 根据区块hash和索引查询交易信息 GetTransactionByBlockHashAndIndex
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockHashAndIndex","params":["0x87dba1c547fbdfec844e56c9ac2f412f2df989ad65a63ecc00fcb6b319f66adf",0],"id":1}'

# 根据区块number和索引查询交易信息 GetTransactionsByBlockNumberAndIndex
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockNumberAndIndex","params":[1,0],"id":1}'

# 查询指定区块的交易数量
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getBlockTransactionCountByHash","params":["<block hash>"],"id":1}'

# 获取交易收据Receipt GetTransactionReceipt
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionReceipt","params":["<tx hash>"],"id":1}'

# 获取所有交易 GetTransactions
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactions","params":[],"id":1}'

########## contract 服务 ##########
#　部署合约 DeployContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"contract_deployContract","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","payload":"0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101cd806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003f578063569c5f6d146100675780638da9b7721461007c578063d09de08a146100ea575b610002565b34610002576000805460043563ffffffff8216016024350163ffffffff19919091161790555b005b346100025761010e60005463ffffffff165b90565b3461000257604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101289490928301828280156101c15780601f10610196576101008083540402835291602001916101c1565b34610002576100656000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101885780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b8154815290600101906020018083116101a457829003601f168201915b5050505050905061007956","privKey":"c2c1149d93f52d586b4ab9d9b634bf9b0221a2f6b78710d58bcbe417482884ca"}],"id":1}'

# 编译ABI CompileContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"contract_compileContract","params":["contract Accumulator{    uint32 sum = 0;   function increment(){         sum = sum + 1;     }      function getSum() returns(uint32){         return sum;     }   function add(uint32 num1,uint32 num2) {         sum = sum+num1+num2;     } }"],"id":1}'

# 调用合约方法 InvokeContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_invokeContract","params": [{"from": "<caller address>", "to": "<contract address>", "payload": "<encode data>"}],"id": 1}'

# 获取合约code GetCode
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getCode","params": ["<contract address>"],"id": 1}'

# 获取账户部署合约数量 GetContractCountByAddr
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getContractCountByAddr","params": ["0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"],"id": 1}'

########## block 服务 ##########
# 得到最新区块 LastestBlock
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_lastestBlock","params":[],"id":1}'

# 得到所有区块 GetBlocks
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":1, "to":1}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":1, "to":5}],"id":1}'

# 根据区块hash查询区块信息 GetBlockByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlockByHash","params":["<block hash>"],"id":1}'

# 根据区块number查询区块信息 GetBlockByNumber, number可以是整数、十六进制字符串或者“latest”
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlockByNumber","params":[<block number>],"id":1}'

# 查询交易平均处理时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "block_queryExecuteTime","params": [{"from":"the number of block","to":"the number of block"}],"id": 1}'

# 查询区块commit、batch平均时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryCommitAndBatchTime","params":[{"from":"the number of block", "to":"the number of block"}],id: 1}'

# 查询EVM平均处理时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryEvmAvgTime","params":[{"from":"the number of block","to":"the number of block"}],"id": 1}'

########## account服务 ##########
# 新建一个账户 NewAccount
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"acc_newAccount","params":["123456"],"id": 1}'

# 解锁账户 UnlockAccount
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"acc_unlockAccount","params":[{"address":0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd, "password":"123"}],"id": 1}'

# 获取所有账户信息 GetAccounts
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"acc_getAccounts","params":[],"id": 1}'

# 查询账户余额 GetBalance
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"acc_getBalance","params":["0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"],"id": 1}'

########## node服务 ##########
# 得到节点信息
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_getNodes","id": 1}'

######### 批量调用例子 #########
curl localhost:8081 --data '[{"jsonrpc":"2.0","method":"block_lastestBlock","params":[],"id":1}, {"jsonrpc":"2.0","method": "node_getNodes","id": 2}]'
