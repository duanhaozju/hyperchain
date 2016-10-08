#!/usr/bin/env bash
########## tx 服务 ##########
# 普通交易 SendTransaction
curl localhost:8081 --data '{"method":"tx_sendTransaction","params":[{"from":"0x0000000000000000000000000000000000000002","to":"0x0000000000000000000000000000000000000003","value":"0x9184e72a"}],"id":1}'

#　部署合约 SendTransactionOrContract
curl localhost:8081 --data '{"method":"tx_sendTransactionOrContract","params":[{"from":"0x0000000000000000000000000000000000000002","payload":"contract Accumulator{ uint sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint){ return sum; }}"}],"id":1}'

# 编译ABI ComplieContract
curl localhost:8081 --data '{"method":"tx_complieContract","params":["contract Accumulator{ uint sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint){ return sum; }}"],"id":1}'

# 调用合约方法 SendTransactionOrContract
curl localhost:8081 --data '{"method": "tx_sendTransactionOrContract","params": [{"from": "caller address", "to": "contract address", "payload": "encode data"}],id: 1}'

########## block 服务 ##########
# 得到最新区块
curl localhost:8081 --data '{"method":"block_lastestBlock","params":[],"id":1}'

# 得到所有区块
curl localhost:8081 --data '{"method":"block_getBlocks","params":[],"id":1}'

# 查询交易平均处理时间
curl localhost:8081 --data '{"method": "block_queryExecuteTime","params": [{"from":"the number of block","to":"the number of block"}],"id": 1}'

# 查询区块commit、batch平均时间
curl localhost:8081 --data '{"method":"block_queryCommitAndBatchTime","params":[{"from":"the number of block", "to":"the number of block"}],id: 1}'

# 查询EVM平均处理时间
curl localhost:8081 --data '{"method":"block_queryEvmAvgTime","params":[{"from":"the number of block","to":"the number of block"}],"id": 1}'

########## node服务 ##########
# 得到节点信息
curl localhost:8081 --data '{"method": "node_getNodes","id": 1}'
