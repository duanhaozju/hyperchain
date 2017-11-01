#!/usr/bin/env bash
################################ tx service ################################
# SendTransaction
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"tx_sendTransaction","params":[{"from":"17d806c92fa941b4b7a8ffffc58fa2f297a3bffc","to":"000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","timestamp":1481767697713000000,"nonce":6482432088169214,"value":1,"signature":"0x6dc99241793600eeeed72578e8de490894ec509bbc6c85753caf3176342c9cb85aa102f50424d508cc23d0c544763020f8136a742726a2b91bf55ff9ca45837e01"}],"id":1}'

# GetTransactionByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByHash","params":["0x975a7c1df0eafeec28624eca66ea3cfb3f5f3cba942564b5fd80da58c9ec79a7"],"id":1}'

# GetTransactionByBlockHashAndIndex
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockHashAndIndex","params":["0x9e330e8890df02d22a7ade73b5060db6651658b676dc9b30e54537853e39c81d",0],"id":1}'

# GetTransactionsByBlockNumberAndIndex
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionByBlockNumberAndIndex","params":[1,0],"id":1}'

# GetBlockTransactionCountByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getBlockTransactionCountByHash","params":["<block hash>"],"id":1}'

# GetTransactionReceipt
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionReceipt","params":["<tx hash>"],"id":1}'

# GetTransactions
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactions","params":[{"from":5, "to":"latest"}],"id":1}'

# GetDiscardTransactions
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getDiscardTransactions","params":[],"id":1}'

# GetSignHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getSignHash","params":[{"from":"17d806c92fa941b4b7a8ffffc58fa2f297a3bffc","nonce":1775845467490815,"payload":"0x606060405234610000575b6101e1806100186000396000f3606060405260e060020a60003504636fd7cc16811461002957806381053a7014610082575b610000565b346100005760408051606081810190925261006091600491606491839060039083908390808284375093955061018f945050505050565b6040518082606080838184600060046018f15090500191505060405180910390f35b346100005761010a600480803590602001908201803590602001908080602002602001604051908101604052809392919081815260200183836020028082843750506040805187358901803560208181028481018201909552818452989a9989019892975090820195509350839250850190849080828437509496506101bc95505050505050565b6040518080602001806020018381038352858181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018381038252848181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f15090500194505050505060405180910390f35b6060604051908101604052806003905b600081526020019060019003908161019f5750829150505b919050565b60408051602081810183526000918290528251908101909252905281815b925092905056","timestamp":1481767468349000000}],"id":1}'

# GetTransactionsCount
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionsCount","params":[],"id":1}'

# GetTxAvgTimeByBlockNumber
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTxAvgTimeByBlockNumber","params":[{"from":2,"to":8}],"id":1}'

# GetTransactionsByTime
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"tx_getTransactionsByTime","params":[{"startTime":1, "endTime":1581776001230590326}],"id":1}'


################################ contract service #################################
#　DeployContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"contract_deployContract","params":[{"from":"17d806c92fa941b4b7a8ffffc58fa2f297a3bffc","nonce":1775845467490815,"payload":"0x606060405234610000575b6101e1806100186000396000f3606060405260e060020a60003504636fd7cc16811461002957806381053a7014610082575b610000565b346100005760408051606081810190925261006091600491606491839060039083908390808284375093955061018f945050505050565b6040518082606080838184600060046018f15090500191505060405180910390f35b346100005761010a600480803590602001908201803590602001908080602002602001604051908101604052809392919081815260200183836020028082843750506040805187358901803560208181028481018201909552818452989a9989019892975090820195509350839250850190849080828437509496506101bc95505050505050565b6040518080602001806020018381038352858181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018381038252848181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f15090500194505050505060405180910390f35b6060604051908101604052806003905b600081526020019060019003908161019f5750829150505b919050565b60408051602081810183526000918290528251908101909252905281815b925092905056","timestamp":1481767468349000000,"signature":"0x3fd9d4bfd7ffae745218e49941cbbb353af649d13590414e7ad333214efa1a1f28cfbc71b66fc4d5642bf14cb328e151a854bf9dc659cdf1b556e344156dc77201"}],"id":1}'

# CompileContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"contract_compileContract","params":["contract Accumulator{    uint32 sum = 0;   function increment(){         sum = sum + 1;     }      function getSum() returns(uint32){         return sum;     }   function add(uint32 num1,uint32 num2) {         sum = sum+num1+num2;     } }"],"id":1}'

# InvokeContract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_invokeContract","params": [{"from": "17d806c92fa941b4b7a8ffffc58fa2f297a3bffc", "to": "0x3a3cae27d1b9fa931458b5b2a5247c5d67c75d61", "timestamp":1481767474717000000, "nonce": 8054165127693853,"payload": "0x6fd7cc16000000000000000000000000000000000000000000000000000000000000007b00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "signature":"0x19c0655d05b9c24f5567846528b81a25c48458a05f69f05cf8d6c46894b9f12a02af471031ba11f155e41adf42fca639b67fb7148ddec90e7628ec8af60c872c00"}],"id": 1}'

# MaintainContract: update contract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_maintainContract","params": [{"from": "17d806c92fa941b4b7a8ffffc58fa2f297a3bffc", "to": "0x3a3cae27d1b9fa931458b5b2a5247c5d67c75d61", "timestamp":1481767474717000000, "nonce": 8054165127693853,"payload": "0x6fd7cc16000000000000000000000000000000000000000000000000000000000000007b00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "signature":"0x19c0655d05b9c24f5567846528b81a25c48458a05f69f05cf8d6c46894b9f12a02af471031ba11f155e41adf42fca639b67fb7148ddec90e7628ec8af60c872c00", "opcode": 1}],"id": 1}'

# MaintainContract: freeeze contract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_maintainContract","params": [{"from": "17d806c92fa941b4b7a8ffffc58fa2f297a3bffc", "to": "0x3a3cae27d1b9fa931458b5b2a5247c5d67c75d61", "timestamp":1481767474717000000, "nonce": 8054165127693853, "signature":"0x19c0655d05b9c24f5567846528b81a25c48458a05f69f05cf8d6c46894b9f12a02af471031ba11f155e41adf42fca639b67fb7148ddec90e7628ec8af60c872c00", "opcode": 2}],"id": 1}'

# MaintainContract: unfreeeze contract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_maintainContract","params": [{"from": "17d806c92fa941b4b7a8ffffc58fa2f297a3bffc", "to": "0x3a3cae27d1b9fa931458b5b2a5247c5d67c75d61", "timestamp":1481767474717000000, "nonce": 8054165127693853, "signature":"0x19c0655d05b9c24f5567846528b81a25c48458a05f69f05cf8d6c46894b9f12a02af471031ba11f155e41adf42fca639b67fb7148ddec90e7628ec8af60c872c00", "opcode": 3}],"id": 1}'

# GetStatus: get contract status
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getStatus","params": ["<contract address>"],"id": 1}'

# GetDeployedList
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getDeployedList","params": ["<address>"],"id": 1}'

# GetCreator: get creator of contract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getCreator","params": ["<contract address>"],"id": 1}'

# GetCreateTime: get create time of contract
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getCreateTime","params": ["<contract address>"],"id": 1}'

# GetArchive: show archieve data
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getArchive","params": ["<contract address>", "20170329"],"id": 1}'

# GetCode
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getCode","params": ["<contract address>","<block number>"],"id": 1}'

# GetContractCountByAddr: get the number of deployed contract by account
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_getContractCountByAddr","params": ["0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","<block number>"],"id": 1}'


curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_encryptoMessage","params": [{"balance":100, "amount":10, "hmBalance":"123456"}],"id": 1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "contract_checkHmValue","params": [{"rawValue": [1,2], "encryValue": ["123", "456"]}],"id": 1}'

################################ block service ####################################
# LastestBlock
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_latestBlock","namespace":"global","params":[],"id":1}'

# GetBlocks
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":1, "to":1}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":1, "to":5}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":5, "to":"latest"}],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocks","params":[{"from":5, "to":"latest", "isPlain": true}],"id":1}'

# GetBlocksByTime
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlocksByTime","params":[{"startTime":1, "endTime":1581777983699073419}],"id":1}'

# GetBlockByHash
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getBlockByHash","params":["<block hash>", false],"id":1}'

# GetBlockByNumber
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"block_getBlockByNumber","params":[<block number>, false],"id":1}'
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"block_getBlockByNumber","params":["latest", false],"id":1}'

# QueryCommitAndBatchTime
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryCommitAndBatchTime","params":[{"from":"the number of block", "to":"the number of block"}],id: 1}'

# QueryEvmAvgTime
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_queryEvmAvgTime","params":[{"from":"the number of block","to":"the number of block"}],"id": 1}'

# GetGenesisBlock
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"block_getGenesisBlock","params":[],"id": 1}'

################################## account service ##################################
# GetAccounts
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"account_getAccounts","params":[],"id": 1}'

# GetBalance
curl localhost:8081 --data '{"jsonrpc":"2.0","method":"account_getBalance","params":["0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"],"id": 1}'


################################# node service ######################################
# GetNodes: get all node information
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_getNodes", "namespace": "global", "id": 1}'

# GetNodeHash: get node hash
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_getNodeHash","id": 1}'

# DeleteVP
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_deleteVP","params":[{"nodehash":"<the node hash of the deleting node>"}],"id": 1}'

# DeleteNVP
curl localhost:8081 --data '{"jsonrpc":"2.0","method": "node_deleteNVP","params":[{"nodehash":"<the node hash of the deleting node>"}],"id": 1}'


################################# batch request example ###################################
curl localhost:8081 --data '[{"jsonrpc":"2.0","method":"block_lastestBlock","params":[],"id":1}, {"jsonrpc":"2.0","method": "node_getNodes","id": 2}]'


################################# archive service ######################################
# Snapshot: make a snapshot
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"archive_snapshot","params":[0],"id":1}'

# ReadSnapshot: query snapshot information
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"archive_readSnapshot,"params":["<filterId>", isVerbose],"id":1}'

# ListSnapshot: get snapshot list
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"archive_listSnapshot","params":[],"id":1}'

# DeleteSnapshot: delete snapshot
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"archive_deleteSnapshot","params":["<filterId>"],"id":1}'

# CheckSnapshot
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"archive_checkSnapshot","params":["<filterId>"],"id":1}'

# Archive: archive data
curl localhost:8081 --data '{"jsonrpc":"2.0","namespace":"global","method":"archive_archive","params":["<filterId>"],"id":1}'


