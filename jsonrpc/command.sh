#!/usr/bin/env bash
# tx 服务
# 普通交易
curl localhost:8084 --data '{"method":"tx_sendTransaction","params":[{"from":"0x0000000000000000000000000000000000000002","to":"0x0000000000000000000000000000000000000003","value":"0x9184e72a"}],"id":1}'

#　合约交易
curl localhost:8084 --data '{"method":"tx_sendTransactionOrContract","params":[{"from":"0x0000000000000000000000000000000000000002","payload":"contract Accumulator{ uint sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint){ return sum; }}"}],"id":1}'

# 编译ABI ComplieContract
curl localhost:8084 --data '{"method":"tx_complieContract","params":["contract Accumulator{ uint sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint){ return sum; }}"],"id":1}'

# block 服务
# 得到最新区块
curl localhost:8084 --data '{"method":"block_lastestBlock","params":[],"id":1}'

# 得到所有区块
curl localhost:8084 --data '{"method":"block_getBlocks","params":[],"id":1}'

# 所有交易的平均处理时间
curl localhost:8084 --data '{"method":"block_queryExcuteTime","params":[],"id":1}'

# node服务
#
