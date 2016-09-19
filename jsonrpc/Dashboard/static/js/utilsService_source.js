/**
 * Created by sammy on 16-9-17.
 */

var SF = require("w3.js");

angular
    .module('starter')
    .factory('UtilsService', function($q) {
        return {
            encode: function(abimethod, params) {
                console.log(abimethod);
                console.log(params);

                var p = [];

                for (var key in params) {
                    p.push(params[key])
                }

                return $q(function(resolve, reject){
                    var sf = new SF(abimethod);
                    var data  = sf.getData.apply(sf, p);
                    // var data  = sf.getData.apply(null, p);
                    resolve(data)
                });


            }
        }
    });