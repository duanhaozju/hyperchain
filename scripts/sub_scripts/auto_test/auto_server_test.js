/**
 * this file is for test multi http request by once click
 * the request must be async, so use the nodeJS
 * author: ChenQuan
 * create date: 2016-09-04
 * description: this is a pure http request file, depends on noting, and you can use this just type `node auto_test.js`
 *
 */

 var http  = require('http');
 var config = require('../../../config/peerconfig.json');
 var genesis = require('../../../config/genesis.json');
 var address = genesis.genesis.alloc;
 var addresses = Object.keys(address);
 var nodes = config.nodes;

function testRequest(opt){
var options = {
    host: opt.url,
    port: opt.port,
    //path: '/trans',
    method: 'POST',
    headers: {
          'Content-Type': 'application/json'
    }
};
//var post_data = JSON.stringify({"from":opt.from,"to":opt.to,"value":'1'});
var post_data = JSON.stringify({
    "method": "tx_sendTransaction",
    // "method": "tx_sendTransactionOrContract",
    "params": [
        {
            "from":opt.from,
            "to":opt.to,
            "value": '1',
            "payload" : "0x8da9b772",
        }
    ],
    "id": 1
});
console.log(options);
// Set up the request
    var post_req = http.request(options, function(res) {
        res.setEncoding('utf8');
        res.on('data', function (chunk) {
            console.log('Response: OK');
        });
       });
    post_req.on('error',function(err){
        console.log(err);
    });
    post_req.write(post_data);
    post_req.end();
}

//http.request(options, callback).end();
for (var node in nodes) {
     testRequest({
        'url':nodes[node].external_address,
        'port':nodes[node].rpc_port,
        'from':addresses[0],
        'to':addresses[2]
        })
}
