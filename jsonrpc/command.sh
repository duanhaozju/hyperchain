#!/usr/bin/env bash
########## tx 服务 ########## （说明：value可以是十六进制字符串、八进制字符串、十进制字符串或整数）
# 普通交易 SendTransaction
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":"0x01"}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":1}],"id":1}'

#　部署合约 SendTransactionOrContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_sendTransactionOrContract","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","payload":"0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101cd806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003f578063569c5f6d146100675780638da9b7721461007c578063d09de08a146100ea575b610002565b34610002576000805460043563ffffffff8216016024350163ffffffff19919091161790555b005b346100025761010e60005463ffffffff165b90565b3461000257604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101289490928301828280156101c15780601f10610196576101008083540402835291602001916101c1565b34610002576100656000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101885780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b8154815290600101906020018083116101a457829003601f168201915b5050505050905061007956"}],"id":1}'

# 编译ABI ComplieContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_complieContract","params":["contract Accumulator{ uint sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint){ return sum; }}"],"id":1}'

# 调用合约方法 SendTransactionOrContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "tx_sendTransactionOrContract","params": [{"from": "caller address", "to": "contract address", "payload": "encode data"}],id: 1}'

# 根据交易hash查询交易信息 GetTransactionByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByHash","params":["the transaction hash"],"id":1}'

curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockHashAndIndex","params":["0xf4d42f0d4e276b306f3a6f7c04e06ef52ee9471684408b24cd4096b8402f7721",0],"id":1}'

########## block 服务 ##########
# 得到最新区块
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_lastestBlock","params":[],"id":1}'

# 得到所有区块
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[],"id":1}'

# 查询交易平均处理时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "block_queryExecuteTime","params": [{"from":"the number of block","to":"the number of block"}],"id": 1}'

# 查询区块commit、batch平均时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryCommitAndBatchTime","params":[{"from":"the number of block", "to":"the number of block"}],id: 1}'

# 查询EVM平均处理时间
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryEvmAvgTime","params":[{"from":"the number of block","to":"the number of block"}],"id": 1}'

########## node服务 ##########
# 得到节点信息
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_getNodes","id": 1}'

######### 批量调用例子 #########
curl localhost:8081 --data '[{"jsonrpc":"2.0","method":"block_lastestBlock","params":[],"id":1}, {"jsonrpc":"2.0","method": "node_getNodes","id": 2}]'
