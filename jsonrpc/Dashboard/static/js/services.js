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
                    getBlock:{
                        method:"POST"
                    }
                }).getBlock({
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
        }
    }
}

function AccountService($resource,$q) {
    return {
        getAllAccounts: function(){
            return $q(function(resolve, reject){
                $resource("http://localhost:8084",{},{
                    getBlock:{
                        method:"POST"
                    }
                }).getBlock({
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

angular
    .module('starter')
    .factory('SummaryService', SummaryService)
    .factory('BlockService', BlockService)
    .factory('TransactionService', TransactionService)
    .factory('AccountService', AccountService)