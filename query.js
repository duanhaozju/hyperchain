/**
 * Created by fox on 16-9-18.
 */
/**
 * this file is for test multi http request by once click
 * the request must be async, so use the nodeJS
 * author: ChenQuan
 * create date: 2016-09-04
 * description: this is a pure http request file, depends on noting, and you can use this just type `node auto_test.js`
 *
 */

var http  = require('http');

function testRequest(){
    var options = {
        host: "114.55.64.132",
        port: "8081",
        path: '/query',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    };
    var post_data = JSON.stringify({"from":"186","to":"248"});
    console.log(options);
// Set up the request
    var post_req = http.request(options, function(res) {
        res.setEncoding('utf8');
        res.on('data', function (chunk) {
            // console.log('Response: OK');
            console.log(chunk)
        });
    });
    post_req.on('error',function(err){
        console.log(err);
    });
    post_req.write(post_data);
    post_req.end();
}
testRequest()

//http.request(options, callback).end();

