/**
 * Created by sammy on 16-9-17.
 */

// var SF = require("w3.js");
// var coder = require('w3.js/solidity/coder.js');
// var coder = require('hpc-web3/lib/solidity/coder');

var SolidityFunction = require('hpc-web3/lib/web3/function');
var coder = require('hpc-web3/lib/solidity/coder');
var ethereumUtil = require("ethereumjs-util");
const secp256k1 = require('secp256k1');

angular
    .module('starter')
    .factory('UtilsService', function($q) {

        return {
            encode: function(abimethod, params) {
                console.log(abimethod);

                var p = [];

                for (var i = 0; i <  abimethod.inputs.length; i++) {
                    for (var k in params) {
                        if (abimethod.inputs[i].name == k) {
                            var patt = new RegExp(/int(\d*)\[(\d*)\]/g)
                            var isNumberArray = patt.test(abimethod.inputs[i].type)

                            if (isNumberArray) {
                                if (params[k].indexOf(",") != -1) {
                                    var n_params = params[k].split(",")
                                    p.push(n_params)
                                } else {
                                    var n_params_arr = []
                                    n_params_arr.push(params[k])
                                    p.push(n_params_arr)
                                }
                            } else {
                                p.push(params[k])
                            }
                            break;
                        }
                    }
                }

                console.log(params)

                var types = abimethod.inputs.map(function(json){
                    return json.type
                })
                console.warn("============ the types sequence is ================")
                for (var i = 0;i < types.length;i++) {
                    console.log(types[i])
                }
                console.warn("============ the values sequence is ================")
                for (var i = 0;i < p.length;i++) {
                    console.log(p[i])
                }

                return $q(function(resolve, reject){
                    // var sf = new SF(abimethod);
                    // var data  = sf.getData.apply(sf, p);
                    // resolve(data)
                    var fun = new SolidityFunction('', abimethod, '');
                    var payload = fun.toPayload(p).data;
                    resolve(payload)
                });
            },
            unpackOutput: function(abi, ret) {
                console.log(abi)
                console.log(ret)

                return $q(function(resolve, reject) {
                    // var sf = new SF(abi)
                    // var upackOut = sf.unpackOutput(ret)

                    var sf = new SolidityFunction('', abi, '')
                    var upackOut = sf.unpackOutput(ret)

                    resolve(upackOut)
                })
            },
            encodeConstructorParams: function(abi, params){
                return $q(function(resolve, reject){
                    console.log(abi)
                    var p = [];

                    // for (var k in params) {
                    //     p.push(params[k])
                    // }

                    console.log(params)

                    data = abi.filter(function (json) {
                            return json.type === 'constructor' && json.inputs.length === Object.keys(params).length;
                        }).map(function (json) {
                            return json.inputs.map(function (input) {
                                for (var k in params) {
                                    if (input.name == k) {
                                        var patt = new RegExp(/int(\d*)\[(\d*)\]/g)
                                        var isNumberArray = patt.test(input.type)

                                        if (isNumberArray) {
                                            if (params[k].indexOf(",") != -1) {
                                                var n_params = params[k].split(",")
                                                p.push(n_params)
                                            } else {
                                                var n_params_arr = []
                                                n_params_arr.push(params[k])
                                                p.push(n_params_arr)
                                            }
                                        } else {
                                            p.push(params[k])
                                        }
                                        break;
                                    }
                                }
                                return input.type;
                            });
                        }).map(function (types) {
                            console.warn("============ the types sequence is ================")
                            for (var i = 0;i < types.length;i++) {
                                console.log(types[i])
                            }
                            console.warn("============ the values sequence is ================")
                            for (var i = 0;i < p.length;i++) {
                                console.log(p[i])
                            }
                            return coder.encodeParams(types, p);
                        })[0] || '';
                    resolve(data)
                })
            },
            hex_to_ascii: function(str1){
                var hex  = str1.toString();
                var str = '';
                for (var n = 0; n < hex.length; n += 2) {
                    str += String.fromCharCode(parseInt(hex.substr(n, 2), 16));
                }
                return str;
            },
            ascii_to_hex: function(pStr) {
                var tempstr = '';
                for (a = 0; a < pStr.length; a = a + 1) {
                    tempstr = tempstr + pStr.charCodeAt(a).toString(16);
                }
                return tempstr;
            },
            getAddress: function(privkey) {
                var privateKey = new Buffer(privkey, 'hex');
                return ethereumUtil.privateToAddress(privateKey).toString('hex');
            },
            getSignature: function(signHash, privkey) {
                var hashBuffer = new Buffer(signHash.substr(2,signHash.length),'hex');
                var privateKey = new Buffer(privkey, 'hex');
                var signature = secp256k1.sign(hashBuffer,privateKey);
                var v = ethereumUtil.intToHex(signature.recovery);
                return '0x'+signature.signature.toString('hex')+v.substr(2,v.length);
            },
            random_16bits: function(){
                var num = Math.random().toString();
                if (num.substr(num.length - 16, 1) === '0') {
                    return this.random_16bits();
                }
                return num.substring(num.length - 16);
            }
        }
    });