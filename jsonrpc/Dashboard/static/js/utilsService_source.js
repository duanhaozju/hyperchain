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
                for (var i = 0; i <  abimethod.inputs.length; i++) {
                    for (var k in params) {
                        if (abimethod.inputs[i].name == k) {
                            p.push(params[k])
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
                    console.log(abi)
                    var p = [];

                    for (var k in params) {
                        p.push(params[k])
                    }

                    // for (var i = 0; i <  abi.inputs.length; i++) {
                    //     for (var k in params) {
                    //         if (abi.inputs[i].name == k) {
                    //             p.push(params[k])
                    //             break;
                    //         }
                    //     }
                    // }

                    console.log(params)

                    data = abi.filter(function (json) {
                            return json.type === 'constructor' && json.inputs.length === p.length;
                        }).map(function (json) {
                            return json.inputs.map(function (input) {
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
            }
        }
    });