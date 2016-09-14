/**
 * Created by sammy on 16-9-13.
 */

function SummaryService($resource,$q) {
    return {
        getLastestBlock: function(){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
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
        getAvgTimeAndCount: function(){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
                    getBlock:{
                        method:"POST"
                    }
                }).getBlock({
                    method: "block_queryExcuteTime",
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

function BlockService($resource,$q) {
    return {
        getAllBlocks: function(){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
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
        }
    }
}

function TransactionService($resource,$q) {
    return {
        getAllTxs: function(){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
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
                $resource("http://localhost:8084",{},{
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
        },
        QueryCommitAndBatchTime: function(from ,to){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
                    sendTx:{
                        method:"POST"
                    }
                }).sendTx({
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
        }
    }
}

function AccountService($resource,$q) {
    return {
        getAllAccounts: function(){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
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
        }
    }
}

function ContractService($resource,$q) {
    return {
        compileContract: function(contract){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
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