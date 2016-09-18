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
                    method: "block_lastestBlock",
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
        getAvgTimeAndCount: function(from,to){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    query:{
                        method:"POST"
                    }
                }).query({
                    method: "block_queryExecuteTime",
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

function TransactionService($resource,$q,ENV) {
    return {
        getAllTxs: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getTxs:{
                        method:"POST"
                    }
                }).getTxs({
                    method: "tx_getTransactions",
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
        SendTransaction: function(from,to,value){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    sendTx:{
                        method:"POST"
                    }
                }).sendTx({
                    method: "tx_sendTransaction",
                    params: [
                        {
                            "from":from, 
                            "to":to, 
                            "value": value
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

function AccountService($resource,$q,ENV) {
    return {
        getAllAccounts: function(){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getAcc:{
                        method:"POST"
                    }
                }).getAcc({
                    method: "acot_getAccounts",
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
                    method: "acot_newAccount",
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
        }
    }
}

function ContractService($resource,$q,ENV) {
    return {
        compileContract: function(contract){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getAcc:{
                        method:"POST"
                    }
                }).getAcc({
                    method: "tx_complieContract",
                    params: [contract],
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
        },
        deployContract: function(from, sourceCode){
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getAcc:{
                        method:"POST"
                    }
                }).getAcc({
                    method: "tx_sendTransactionOrContract",
                    params: [
                        {
                            "from": from,
                            "payload": sourceCode
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
        invokeContract: function(from, to, data) {
            return $q(function(resolve, reject){
                $resource(ENV.API,{},{
                    getAcc:{
                        method:"POST"
                    }
                }).getAcc({
                    method: "tx_sendTransactionOrContract",
                    params: [
                        {
                            "from": from,
                            "to": to,
                            "payload": data
                        }
                    ],
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
    }
}

angular
    .module('starter')
    .factory('SummaryService', SummaryService)
    .factory('BlockService', BlockService)
    .factory('TransactionService', TransactionService)
    .factory('AccountService', AccountService)
    .factory('ContractService', ContractService)