/**
 * Created by sammy on 16-9-29.
 */

function hexStringToNum() {
    return function(str) {
        return parseInt(str, 16)
    }
}

/**
 *
 * Pass all functions into module
 */
angular
    .module('starter')
    .filter('hexStringToNum', hexStringToNum)

