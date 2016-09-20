/**
 * this file is for test multi http request by once click
 * the request must be async, so use the nodeJS
 * author: ChenQuan
 * create date: 2016-09-04
 * description: this is a pure http request file, depends on noting, and you can use this just type `node auto_test.js`
 *
 */

 var http  = require('http');
 var config = require('./p2p/local_peerconfig.json');
 var genesis = require('./genesis.json');
 var address = genesis.test1.alloc;
 var addresses = Object.keys(address);
 var params = {form:"",to:"",value:1};
 var hosts_url = [];
 var hosts_port = [];
 var MAXNODES = parseInt(config['MAXPEERS']);
 for (var i=1;i<=MAXNODES;i++){
    hosts_url.push(config['external_node'+i]);
    hosts_port.push(config['external_port'+i]);
 }

console.log(hosts_url);
console.log(hosts_port);

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
    "params": [
        {
            "from":opt.from,
            "to":opt.to,
            "value": '1'
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
for(var j=0;j<MAXNODES;j++){
    if (j %2 ==0){
     testRequest({
        'url':hosts_url[j],
        'port':hosts_port[j],
        'from':addresses[0],
        'to':addresses[2]
        })
    }else{
     testRequest({
        'url':hosts_url[j],
        'port':hosts_port[j],
        'from':addresses[1],
        'to':addresses[3]
        })
    }
}
