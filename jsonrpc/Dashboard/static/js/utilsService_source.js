/**
 * Created by sammy on 16-9-17.
 */

var SF = require("w3.js");
var coder = require('w3.js/solidity/coder.js');

angular
    .module('starter')
    .factory('UtilsService', function($q) {

        return {
            encode: function(abimethod, params) {
                console.log(abimethod);

                var p = [];

                // for (var key in params) {
                //     p.push(params[key])
                // }
                // 排序
                for (var i = 0; i <  abimethod.inputs.length; i++) {
                    for (var k in params) {
                        if (abimethod.inputs[i].name == k) {
                            p.push(params[k])
                            break;
                        }
                    }
                }

                console.log(params)
                console.log(p)
                return $q(function(resolve, reject){
                    var sf = new SF(abimethod);
                    var data  = sf.getData.apply(sf, p);
                    // var data  = sf.getData.apply(null, p);
                    resolve(data)
                });
            },
            unpackOutput: function(abi, ret) {
                console.log(abi)
                console.log(ret)
                return $q(function(resolve, reject) {
                    var sf = new SF(abi)
                    resolve(sf.unpackOutput(ret))
                })
            },
            encodeConstructorParams: function(abi, params){
                return $q(function(resolve, reject){

                    var p = [];

                    for (var i = 0; i <  abi.inputs.length; i++) {
                        for (var k in params) {
                            if (abi.inputs[i].name == k) {
                                p.push(params[k])
                                break;
                            }
                        }
                    }

                    console.log(abi)
                    console.log(params)
                    console.log(p)
                    data = abi.filter(function (json) {
                            return json.type === 'constructor' && json.inputs.length === params.length;
                        }).map(function (json) {
                            return json.inputs.map(function (input) {
                                return input.type;
                            });
                        }).map(function (types) {
                            return coder.encodeParams(types, params);
                        })[0] || '';
                    resolve(data)
                })
            }
        }
    });