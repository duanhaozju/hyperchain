/**
 * Created by sammy on 16-9-29.
 */

function hexStringToNum() {
    return function(str) {
        return parseInt(str, 16)
    }
}

function timestampToDate() {
    return function(timestamp){
        var timestamp_ms = timestamp/1e6
        var date = new Date(timestamp_ms)
        var Y = date.getFullYear() + '-',
            M = (date.getMonth()+1 < 10 ? '0'+(date.getMonth()+1) : date.getMonth()+1) + '-',
            D = date.getDate() + ' ',
            h = date.getHours() + ':',
            m = date.getMinutes() + ':',
            s = date.getSeconds();
        return Y+M+D+h+m+s;
    }
}

/**
 *
 * Pass all functions into module
 */
angular
    .module('starter')
    .filter('hexStringToNum', hexStringToNum)
    .filter('timestampToDate', timestampToDate)

