# Hyperchain RESTful接口设计


# Blockchain

##1.1 获取最新区块 
**Method:** `GET` <br/>

**URI:** `/v1/blockchain/blocks/latest`<br/>

**Request:** `localhost:9000/v1/blockchain/blocks/latest`<br/>

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "number": "0x4",
    "hash": "0x9c450b2470e82ae22144b04ba727360705d1432637808d4a9701de114ff617b5",
    "parentHash": "0xaaa29f2d348e8737b6bfa9937cf2d93cbabf9a48d7c7311c84e9b63870fb44d8",
    "writeTime": 1479522703681186330,
    "avgTime": "0x4",
    "txcounts": "0x1",
    "merkleRoot": "0x3396bc716c64854bb3a90f55aebdba45c953f39eda95f0f275ad13e7063d765a",
    "transactions": [
      {
        "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
        "blockNumber": "0x4",
        "blockHash": "0x9c450b2470e82ae22144b04ba727360705d1432637808d4a9701de114ff617b5",
        "txIndex": "0x0",
        "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
        "to": "0x0000000000000000000000000000000000000000",
        "amount": "0x0",
        "timestamp": 1478959217012956575,
        "executeTime": "0x4",
        "invalid": false,
        "invalidMsg": ""
      }
    ]
  }
}
```

## 1.2 获取交易总数量
**Method:** `GET`<br/>

**URI:** `/v1/blockchain/transactions/count`<br/>

**Request:** `localhost:9000/v1/blockchain/transactions/count`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0x2"
}
```


# Transactions
##2.1 获取所有交易
**Method:** `GET`<br/>

**URI:** `/v1/transactions/list`<br/>

**Request:** `localhost:9000/v1/transactions/list?from=1&to=2`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": [
    {
      "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
      "blockNumber": "0x2",
      "blockHash": "0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba",
      "txIndex": "0x0",
      "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
      "to": "0x0000000000000000000000000000000000000000",
      "amount": "0x0",
      "timestamp": 1478959217012956575,
      "executeTime": "0x2283740",
      "invalid": false,
      "invalidMsg": ""
    },
    {
      "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
      "blockNumber": "0x1",
      "blockHash": "0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba",
      "txIndex": "0x0",
      "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
      "to": "0x0000000000000000000000000000000000000000",
      "amount": "0x0",
      "timestamp": 1478959217012956575,
      "executeTime": "0x2283740",
      "invalid": false,
      "invalidMsg": ""
    }
  ]
}
```

##2.2 获取交易（hash）
**Method:** `GET`<br/>

**URI: **`/v1/transactions/:txHash`<br/>

**Request:** `localhost:9000/v1/transactions/0xb6835407fb8d7ff3533eaecf5c32151add9ad58bf17e3179587eb71e8ce8a9b2`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "hash": "0xb6835407fb8d7ff3533eaecf5c32151add9ad58bf17e3179587eb71e8ce8a9b2",
    "blockNumber": "0x3",
    "blockHash": "0xaaa29f2d348e8737b6bfa9937cf2d93cbabf9a48d7c7311c84e9b63870fb44d8",
    "txIndex": "0x0",
    "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
    "to": "0x80958818f0a025273111fba92ed14c3dd483caeb",
    "amount": "0x35",
    "timestamp": 1478959217012956575,
    "executeTime": "0x4",
    "invalid": false,
    "invalidMsg": ""
  }
}
```

## 2.3 获取交易（blkNum）
**Method:** `GET`<br/>

**URI:** `/v1/transactions/query`<br/>

**Request:** `localhost:9000/v1/transactions/query?blkNum=1&index=0`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
    "blockNumber": "0x1",
    "blockHash": "0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba",
    "txIndex": "0x0",
    "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
    "to": "0x0000000000000000000000000000000000000000",
    "amount": "0x0",
    "timestamp": 1478959217012956575,
    "executeTime": "0x2283740",
    "invalid": false,
    "invalidMsg": ""
  }
}
```

##2.4 获取交易（blkhash）
**Method:** `GET`<br/>

**URI:** `/v1/transactions/query`<br/>

**Request:** `localhost:9000/v1/transactions/query?blkHash=0xaaa29f2d348e8737b6bfa9937cf2d93cbabf9a48d7c7311c84e9b63870fb44d8&index=0`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "hash": "0xb6835407fb8d7ff3533eaecf5c32151add9ad58bf17e3179587eb71e8ce8a9b2",
    "blockNumber": "0x3",
    "blockHash": "0xaaa29f2d348e8737b6bfa9937cf2d93cbabf9a48d7c7311c84e9b63870fb44d8",
    "txIndex": "0x0",
    "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
    "to": "0x80958818f0a025273111fba92ed14c3dd483caeb",
    "amount": "0x35",
    "timestamp": 1478959217012956575,
    "executeTime": "0x4",
    "invalid": false,
    "invalidMsg": ""
  }
}
```

##2.5 发送交易
**Method:** `POST`<br/>

**URI:** `/v1/transactions/send`<br/>

**Request:** `localhost:9000/v1/transactions/send`<br/>
```json
{
	"from":"000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
	"to":"0x80958818f0a025273111fba92ed14c3dd483caeb",
	"timestamp":1478959217012956575,
	"value":53,
	"signature":"a68c4b492f020c2cc98bd15b64020a569365e1ed0a17cb38e8c472758f01dc2c2c60041e0ea90e6ffe3580d7629138c51f3927263dda23bc001e8e9bc368a0c900"
}
```

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0xb6835407fb8d7ff3533eaecf5c32151add9ad58bf17e3179587eb71e8ce8a9b2"
}
```

## 2.6 获取交易回执
**Method:** `GET`<br/>

**URI:** `/v1/transactions/:txHash/receipt`<br/>

**Request:** `localhost:9000/v1/transactions/0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8/receipt`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "txHash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
    "postState": "",
    "contractAddress": "0x3379474040a904478dcaec148398a0ebbc44a186",
    "ret": "0x606060405260e060020a60003504633ad14af381146030578063569c5f6d14605e578063d09de08a146084575b6002565b346002576000805460e060020a60243560043563ffffffff8416010181020463ffffffff199091161790555b005b3460025760005463ffffffff166040805163ffffffff9092168252519081900360200190f35b34600257605c6000805460e060020a63ffffffff821660010181020463ffffffff1990911617905556"
  }
}
```

##2.7 获取区块交易数量
**Method:** `GET`<br/>

**URI:** `/v1/blocks/:blkHash/transactions/count`<br/>

**Request:** `localhost:9000/v1/blocks/0x5f0d026ded386f6a797a95791a6e47f7ed384728b3e2febf7e12397fe3adbba9/transactions/count`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0x2"
}
```

## 2.8 获取签名哈希 
**Method:** `POST`<br/>

**URI:** `/v1/transactions/get-hash-for-sign`<br/>

**Request:** `localhost:9000/v1/transactions/get-hash-for-sign`<br/>
```json
{
	"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
	"to":"0x0000000000000000000000000000000000000003",
	"value":"0x01",
	"timestamp":1478959217012956575
}
```

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0xbd7c8f92833739c3bad6851d2164da0540179c569db37d616e4058a44471ac58"
}
```

## 2.9 获取交易平均 处理时间
**Method:** `POST`<br/>

**URI:** `/v1/transactions/average-time`<br/>

**Request:** `localhost:9000/v1/transactions/average-time?from=1&to=1`<br/>

**Response:** <br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0x6"
}
```

# Blocks
##3.1 获取所有区块
**Method:** `GET`<br/>

**URI:** `/v1/blocks/list`<br/>

**Request:** `localhost:9000/v1/blocks/list?from=1&to=2`<br/>

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": [
    {
      "number": "0x2",
      "hash": "0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba",
      "parentHash": "0x45c448fbb51d95a5e2da7f5e139aecfe872a9ed0638efb9d6fb5d491490956d4",
      "writeTime": 1479520254280610845,
      "avgTime": "0x2283740",
      "txcounts": "0x1",
      "merkleRoot": "0x0759561de972bfba8637b52d67337bceee481769cdd1d29b682c04f920818de2",
      "transactions": [
        {
          "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
          "blockNumber": "0x2",
          "blockHash": "0x9c450b2470e82ae22144b04ba727360705d1432637808d4a9701de114ff617b5",
          "txIndex": "0x0",
          "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
          "to": "0x0000000000000000000000000000000000000000",
          "amount": "0x0",
          "timestamp": 1478959217012956575,
          "executeTime": "0x4",
          "invalid": false,
          "invalidMsg": ""
        }
      ]
    },
    {
      "number": "0x1",
      "hash": "0xcc1b90088289373737d19e99b24e3ded317f199893314764058ece60fad811f9",
      "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "writeTime": 1479521644289378812,
      "avgTime": "0x23dbe08",
      "txcounts": "0x1",
      "merkleRoot": "0xec25fd7eefd777a0e40734dd229ab442c89f1585d5b5145e59bbc727be4b20c1",
      "transactions": [
        {
          "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
          "blockNumber": "0x1",
          "blockHash": "0x9c450b2470e82ae22144b04ba727360705d1432637808d4a9701de114ff617b5",
          "txIndex": "0x0",
          "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
          "to": "0x0000000000000000000000000000000000000000",
          "amount": "0x0",
          "timestamp": 1478959217012956575,
          "executeTime": "0x4",
          "invalid": false,
          "invalidMsg": ""
        }
      ]
    }
  ]
}
```

##3.2 获取区块（hash）
**Method:** `GET`<br/>

**URI:** `/v1/blocks/query`<br/>

**Request:** `localhost:9000/v1/blocks/query?blkHash=0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba`<br/>

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "number": "0x2",
    "hash": "0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba",
    "parentHash": "0x45c448fbb51d95a5e2da7f5e139aecfe872a9ed0638efb9d6fb5d491490956d4",
    "writeTime": 1479520254280610845,
    "avgTime": "0x2283740",
    "txcounts": "0x1",
    "merkleRoot": "0x0759561de972bfba8637b52d67337bceee481769cdd1d29b682c04f920818de2",
    "transactions": [
      {
        "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
        "blockNumber": "0x2",
        "blockHash": "0x9c450b2470e82ae22144b04ba727360705d1432637808d4a9701de114ff617b5",
        "txIndex": "0x0",
        "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
        "to": "0x0000000000000000000000000000000000000000",
        "amount": "0x0",
        "timestamp": 1478959217012956575,
        "executeTime": "0x4",
        "invalid": false,
        "invalidMsg": ""
      }
    ]
  }
}
```

##3.3 获取区块（number）
**Method:** `GET`<br/>

**URI:** `/v1/blocks/query`<br/>

**Request:** `localhost:9000/v1/blocks/query?blkNum=2`<br/>

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "number": "0x2",
    "hash": "0x6443238a7384940188b8c7da57850c2af1230e63be2d4b82d9f9fcb4c08325ba",
    "parentHash": "0x45c448fbb51d95a5e2da7f5e139aecfe872a9ed0638efb9d6fb5d491490956d4",
    "writeTime": 1479520254280610845,
    "avgTime": "0x2283740",
    "txcounts": "0x1",
    "merkleRoot": "0x0759561de972bfba8637b52d67337bceee481769cdd1d29b682c04f920818de2",
    "transactions": [
      {
        "hash": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8",
        "blockNumber": "0x2",
        "blockHash": "0x9c450b2470e82ae22144b04ba727360705d1432637808d4a9701de114ff617b5",
        "txIndex": "0x0",
        "from": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
        "to": "0x0000000000000000000000000000000000000000",
        "amount": "0x0",
        "timestamp": 1478959217012956575,
        "executeTime": "0x4",
        "invalid": false,
        "invalidMsg": ""
      }
    ]
  }
  }
```
# Contracts
##4.1 编译合约 
**Method:** `POST`<br/>

**URI:** `/v1/contracts/compile`<br/>

**Request:** `localhost:9000/v1/contracts/compile`<br/>
```json
{
	"source":"contract HelloWorld{    string hello= \"hello world\";    function getHello() returns(string) {    return hello;    }}"
}
```

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": {
    "abi": [
      "[{\"constant\":false,\"inputs\":[],\"name\":\"getHello\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"}]"
    ],
    "bin": [
      "0x60a0604052600b6060527f68656c6c6f20776f726c64000000000000000000000000000000000000000000608052600080548180527f68656c6c6f20776f726c64000000000000000000000000000000000000000016825560b0907f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563602060026001841615610100026000190190931692909204601f01919091048101905b8082111560c05760008155600101609e565b505061012f806100c46000396000f35b509056606060405260e060020a60003504638da9b772811461001e575b610002565b3461000257604080516020808201835260008083528054845160026001831615610100026000190190921691909104601f810184900484028201840190955284815261008c9490928301828280156101255780601f106100fa57610100808354040283529160200191610125565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156100ec5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b81548152906001019060200180831161010857829003601f168201915b505050505090509056"
    ],
    "types": [
      "HelloWorld"
    ]
  }
}
```

##4.2 部署合约
**Method:** `POST`<br/>

**URI:** `/v1/contracts/deploy`<br/>

**Request:** `localhost:9000/v1/contracts/deploy`<br/>
```json
{
	"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
	"payload":"0x60606040526000805463ffffffff1916815560ae908190601e90396000f3606060405260e060020a60003504633ad14af381146030578063569c5f6d14605e578063d09de08a146084575b6002565b346002576000805460e060020a60243560043563ffffffff8416010181020463ffffffff199091161790555b005b3460025760005463ffffffff166040805163ffffffff9092168252519081900360200190f35b34600257605c6000805460e060020a63ffffffff821660010181020463ffffffff1990911617905556",
	"timestamp":1478959217012956575,
	"signature":"28e82952fd3264507de6a47189fb650846628a78324e064cae30931577b6edd90b9a6d1aaa41b9485873707de7cb123cb02521acb21b99174c42da1846cf772d01"
}
```

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0x163493ca333b478c77f31441ff8767afcfbbdfe3c0287eb981351d8bf661fcf8"
}
```

##4.3 调用合约 
**Method:** `POST`<br/>

**URI:** `/v1/contracts/invoke`<br/>

**Request:** `localhost:9000/v1/contracts/invoke`<br/>

**Response:**<br/>

##4.4 获取合约编码 
**Method:** `GET`<br/>

**URI:** `/v1/contracts/query`<br/>

**Request:** `localhost:9000/v1/contracts/query?address=0xb8dc305352edef315f6a7844948c864717e93c84&blkNum=5`<br/>

**Response:**<br/>

```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0x606060405260e060020a60003504633ad14af381146030578063569c5f6d14605e578063d09de08a146084575b6002565b346002576000805460e060020a60243560043563ffffffff8416010181020463ffffffff199091161790555b005b3460025760005463ffffffff166040805163ffffffff9092168252519081900360200190f35b34600257605c6000805460e060020a63ffffffff821660010181020463ffffffff1990911617905556"
}
```

# Accounts

## 5.1 获取账户已部署合约数量
**Method:** `GET`<br/>

**URI:** `/v1/accounts/:address/contracts/count`<br/>

**Request:** `localhost:9000/v1/accounts/0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd/contracts/count?blkNum=1`<br/>

**Response:**<br/>
```json
{
  "code": 200,
  "message": "SUCCESS",
  "result": "0x1"
}
```