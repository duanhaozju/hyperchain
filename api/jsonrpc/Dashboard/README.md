# Hyperchain Test Framework
This is a test framework to contract with signature.

## Prepare
如果不修改 utilsService.js 和　utilsService_source.js 代码，可以不执行下面的步骤。
>cd jsonrpc/Dashboard

>npm install 

if it is slowly, please use:
>npm install --registry=https://registry.npm.taobao.org

说明：
1. 部署合约和调用合约不再是传入以前那种写法from，而是传入一个私钥字符串（未加密的私钥）。