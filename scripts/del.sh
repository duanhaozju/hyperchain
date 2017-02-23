hash=`curl 127.0.0.1:8083 --data '{"jsonrpc":"2.0","method": "node_getNodeHash","id": 1}' | jq .result`
echo $hash


curl 127.0.0.1:8081 --data "{\"jsonrpc\":\"2.0\",\"method\":\"node_delNode\",\"params\":[{\"nodehash\":${hash}}],\"id\":1}" 
curl 127.0.0.1:8082 --data "{\"jsonrpc\":\"2.0\",\"method\":\"node_delNode\",\"params\":[{\"nodehash\":${hash}}],\"id\": 1}" 
curl 127.0.0.1:8083 --data "{\"jsonrpc\":\"2.0\",\"method\":\"node_delNode\",\"params\":[{\"nodehash\":${hash}}],\"id\": 1}"
curl 127.0.0.1:8084 --data "{\"jsonrpc\":\"2.0\",\"method\":\"node_delNode\",\"params\":[{\"nodehash\":${hash}}],\"id\": 1}"
curl 127.0.0.1:8085 --data "{\"jsonrpc\":\"2.0\",\"method\":\"node_delNode\",\"params\":[{\"nodehash\":${hash}}],\"id\": 1}"

