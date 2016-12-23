/**
 * Created by sammy on 16-9-13.
 */

function SummaryService($resource,$q,ENV) {
    return {
        getLastestBlock: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getBlock:{
                        method:"POST"
                    }
                }).getBlock({
                    method: "block_latestBlock",
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        getAvgTime: function(from,to){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    query:{
                        method:"POST"
                    }
                }).query({
                    method: "tx_getTxAvgTimeByBlockNumber",
                    params: [
                        {
                            "from":from,
                            "to":to
                        }
                    ],
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        getTransactionSum:function () {
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getTxSum:{
                        method:"POST"
                    }
                }).getTxSum({
                    method: "tx_getTransactionsCount",
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        getNodeInfo: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getNodes:{
                        method:"POST"
                    }
                }).getNodes({
                    method: "node_getNodes",
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        }
    }
}

function BlockService($resource,$q,ENV) {
    return {
        getAllBlocks: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getBlock:{
                        method:"POST"
                    }
                }).getBlock({
                    method: "block_getBlocks",
                    params: [{}],
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        queryCommitAndBatchTime: function(from ,to){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    query:{
                        method:"POST"
                    }
                }).query({
                    method: "block_queryCommitAndBatchTime",
                    params: [
                        {
                            "from":from,
                            "to":to
                        }
                    ],
                    id: 1
                },function(res){
                    console.log(res);
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        queryEvmAvgTime: function(from, to) {
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    query:{
                        method:"POST"
                    }
                }).query({
                    method: "block_queryEvmAvgTime",
                    params: [
                        {
                            "from":from,
                            "to":to
                        }
                    ],
                    id: 1
                },function(res){
                    console.log(res);
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        }
    }
}

function TransactionService($resource,$q,ENV,UtilsService) {

    var getTxSignHash = function(txData){
        return $q(function(resolve, reject){
            $resource(ENV.API,{},{
                getHash:{
                    method:"POST"
                }
            }).getHash({
                method: "tx_getSignHash",
                params: [
                    {
                        "from":txData.from,
                        "to": txData.to,
                        "value": txData.value,
                        "timestamp": txData.timestamp,
                        "nonce": txData.nonce
                    }
                ],
                id: 1
            },function(res){
                console.log(res);
                if (res.error) {
                    reject(res.error)
                } else {
                    resolve(res.result)
                }

            })
        })
    }

    return {
        getAllTxs: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getTxs:{
                        method:"POST"
                    }
                }).getTxs({
                    method: "tx_getTransactions",
                    params: [{}],
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        SendTransaction: function(privkey,to,value){

            var from = UtilsService.getAddress(privkey);

            var txData = {
                from: from,
                to: to,
                value: value,
                timestamp: (new Date().getTime())*1e6, //to ns
                nonce: parseInt(UtilsService.random_16bits(), 10)
            };

            return $q(function(resolve, reject){

                getTxSignHash(txData).then(function(hash){
                    txData.signature = UtilsService.getSignature(hash, privkey);
                    $resource(ENV.API,{},{
                        sendTx:{
                            method:"POST"
                        }
                    }).sendTx({
                        method: "tx_sendTransaction",
                        params: [
                            {
                                "from":txData.from,
                                "to":txData.to,
                                "value": txData.value,
                                "timestamp": txData.timestamp,
                                "signature": txData.signature,
                                "nonce": txData.nonce
                            }
                        ],
                        id: 1
                    },function(res){
                        console.log(res);
                        if (res.error) {
                            reject(res.error)
                        } else {
                            resolve(res.result)
                        }

                    })
                }, function(err){
                    reject(err)
                })


            })
        }
    }
}

function AccountService($resource,$q,ENV) {
    return {
        getAllAccounts: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getAcc:{
                        method:"POST"
                    }
                }).getAcc({
                    method: "acc_getAccounts",
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        newAccount: function(password){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    newAcc:{
                        method:"POST"
                    }
                }).newAcc({
                    method: "acc_newAccount",
                    params: [password],
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        unlockAccount:function (address,password) {
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    unlockac:{
                        method:"POST"
                    }
                }).unlockac({
                    method: "acc_unlockAccount",
                    params: [
                        {
                            "address":address,
                            "password":password,
                        }
                    ],
                    id: 1
                },function(res){
                    console.log(res);
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }
                })
            })
        }
    }
}

function ContractService($resource,$q ,$timeout, ENV, UtilsService) {

    var getReceipt = function(txHash){
        return $q(function(resolve, reject){
            $resource(ENV.API,{},{
                get:{
                    method:"POST"
                }
            }).get({
                method: "tx_getTransactionReceipt",
                params: [txHash],
                id: 1
            },function(res){
                console.log(res)
                if (res.error) {
                    reject(res.error)
                } else {
                    resolve(res.result)
                }

            })
        })
    }

    var getDeploySignHash = function(txData){
        return $q(function(resolve, reject){
            $resource(ENV.API,{},{
                getHash:{
                    method:"POST"
                }
            }).getHash({
                method: "tx_getSignHash",
                params: [
                    {
                        "from":txData.from,
                        "payload": txData.payload,
                        "timestamp": txData.timestamp,
                        "nonce": txData.nonce
                    }
                ],
                id: 1
            },function(res){
                console.log(res);
                if (res.error) {
                    reject(res.error)
                } else {
                    resolve(res.result)
                }

            })
        })
    }

    var getInvokeSignHash = function(txData){
        return $q(function(resolve, reject){
            $resource(ENV.API,{},{
                getHash:{
                    method:"POST"
                }
            }).getHash({
                method: "tx_getSignHash",
                params: [
                    {
                        "from":txData.from,
                        "to": txData.to,
                        "payload": txData.payload,
                        "timestamp": txData.timestamp,
                        "nonce": txData.nonce
                    }
                ],
                id: 1
            },function(res){
                console.log(res);
                if (res.error) {
                    reject(res.error)
                } else {
                    resolve(res.result)
                }

            })
        })
    }

    return {
        compileContract: function(contract){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    compile:{
                        method:"POST"
                    }
                }).compile({
                    method: "contract_compileContract",
                    params: [contract],
                    id: 1
                },function(res){
                    if (res.error) {
                        reject(res.error)
                    } else {
                        resolve(res.result)
                    }

                })
            })
        },
        deployContract: function(privkey, sourceCode){
            var from = UtilsService.getAddress(privkey)
            // console.log("from",from)
            var txData = {
                from: from,
                payload: sourceCode,
                timestamp: (new Date().getTime())*1e6, //to ns
                nonce: parseInt(UtilsService.random_16bits(), 10)
            };

            return $q(function(resolve, reject){

                getDeploySignHash(txData).then(function(hash){
                    txData.signature = UtilsService.getSignature(hash, privkey);
                    console.warn(hash)
                    console.warn(txData)
                    $resource(ENV.API,{},{
                        deploy:{
                            method:"POST"
                        }
                    }).deploy({
                        // method: "tx_sendTransactionOrContract",
                        method: "contract_deployContract",
                        params: [
                            {
                                "from": txData.from,
                                "payload": txData.payload,
                                "timestamp": txData.timestamp,
                                "signature": txData.signature,
                                "nonce": txData.nonce
                            }
                        ],
                        id: 1
                    },function(res){
                        console.log(res);
                        if (res.error) {
                            reject(res.error)
                        } else {

                            var flag = false;
                            // var result;

                            var startTime = new Date().getTime();
                            var getResp = function(){
                                console.log(flag);
                                if (!flag) {
                                    if ((new Date().getTime() - startTime) < 8000) {
                                        getReceipt(res.result)
                                            .then(function(data){
                                                console.log(data);

                                                if(data){
                                                    flag = true;
                                                    resolve(data)
                                                }
                                            }, function(error){
                                                console.log(error);
                                                reject(error)
                                            });
                                        $timeout(getResp, 400)
                                    } else {
                                        reject({message: "timeout"})
                                    }
                                }
                            };
                            $timeout(getResp, 400);
                        }

                    })

                },function(err){
                    reject(err)
                })
            })
        },
        invokeContract: function(privkey, to, data) {
            var from = UtilsService.getAddress(privkey);

            var txData = {
                from: from,
                to: to,
                payload: data,
                timestamp: (new Date().getTime())*1e6, //to ns
                nonce: parseInt(UtilsService.random_16bits(), 10)
            };

            return $q(function(resolve, reject){

                getInvokeSignHash(txData).then(function(hash){
                    txData.signature = UtilsService.getSignature(hash, privkey);
                    $resource(ENV.API,{},{
                        invoke:{
                            method:"POST"
                        }
                    }).invoke({
                        method: "contract_invokeContract",
                        params: [
                            {
                                "from": txData.from,
                                "to": txData.to,
                                "payload": txData.payload,
                                "timestamp": txData.timestamp,
                                "signature": txData.signature,
                                "nonce": txData.nonce
                            }
                        ],
                        id: 1
                    },function(res){
                        console.log(res)
                        if (res.error) {
                            reject(res.error)
                        } else {

                            var flag = false;

                            var startTime = new Date().getTime();
                            var getResp = function(){
                                console.log(flag);
                                if (!flag) {
                                    if ((new Date().getTime() - startTime) < 5000) {
                                        getReceipt(res.result)
                                            .then(function(data){
                                                console.log(data);

                                                if(data){
                                                    flag = true;
                                                    resolve(data)
                                                }
                                            }, function(error){
                                                console.log(error);
                                                reject(error)
                                            });
                                        $timeout(getResp, 400)
                                    } else {
                                        reject({message: "timeout"})
                                    }
                                }
                            };
                            $timeout(getResp, 400);
                        }

                    })
                },function(err){
                    reject(err)
                })
            })
        }
        // getContract: function(addr){
        //     return $q(function(resolve, reject){
        //         $resource(ENV.API,{},{
        //             get:{
        //                 method:"POST"
        //             }
        //         }).get({
        //             method: "contract_getContractByAddr",
        //             params: [addr],
        //             id: 1
        //         },function(res){
        //             if (res.error) {
        //                 reject(res.error)
        //             } else {
        //                 resolve(res.result)
        //             }
        //
        //         })
        //     })
        // }
    }
}


angular
    .module('starter')
    .factory('SummaryService', SummaryService)
    .factory('BlockService', BlockService)
    .factory('TransactionService', TransactionService)
    .factory('AccountService', AccountService)
    .factory('ContractService', ContractService)