/**
 * Created by sammy on 16-9-15.
 */


angular
    .module('starter')
    .constant('ENV',{
        "API": "http://114.55.64.132:8081",
        "PATTERN": [
            {name: "pattern0", value: "contract HelloWorld{    string hello= \"hello world\";    function getHello() returns(string) {    return hello;    }}"},
            {name: "pattern1", value: 'contract Accumulator{    uint32 sum = 0;   function increment(){         sum = sum + 1;     }      function getSum() returns(uint32){         return sum;     }   function add(uint32 num1,uint32 num2) {         sum = sum+num1+num2;     } }'},
            {name: "pattern2", value: "contract SimulateBank{" +
            "address owner;" +
            "mapping(address => uint) public accounts;" +
            "function SimulateBank(){" +
            "owner = msg.sender;" +
            "}" +
            "function issue(address addr,uint number) returns (bool){" +
            "if(msg.sender==owner){" +
            "accounts[addr] = accounts[addr] + number;" +
            "return true;" +
            "}" +
            "return false;" +
            "}" +
            "function transfer(address addr1,address addr2,uint32 amount) returns (bool){" +
            "if(accounts[addr1] >= amount){" +
            "accounts[addr1] = accounts[addr1] - amount;" +
            "accounts[addr2] = accounts[addr2] + amount;" +
            "return true;" +
            "}" +
            "return false;" +
            "}" +
            "function getAccountBalance(address addr) returns(uint){" +
            "return accounts[addr];" +
            "}}"}
            // {name: "pattern3", value: "contract mortal {" +
            // "/* Define variable owner of the type address*/" +
            // "address owner;" +
            // "/* this function is executed at initialization and sets the owner of the contract */" +
            // "function mortal() {" +
            // "owner = msg.sender;" +
            // "}" +
            // "/* Function to recover the funds on the contract */" +
            // "function kill() {" +
            // "if (msg.sender == owner)" +
            // "selfdestruct(owner);" +
            // "}" +
            // "}" +
            // "contract greeter is mortal {" +
            // "/* define variable greeting of the type string */" +
            // "string greeting;" +
            // "uint32 sum;" +
            // "/* this runs when the contract is executed */" +
            // "function greeter(string _greeting) public {" +
            // "greeting = _greeting;" +
            // "}" +
            // "/* main function */" +
            // "function greet() constant returns (string) {" +
            // "return greeting;" +
            // "}" +
            // "function add(uint32 num1,uint32 num2) {" +
            // "sum = sum+num1+num2;" +
            // "}" +
            // "}"
            // }
            ],
        "CONTRACT": [],
        "FROM": "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
        "STORAGE": "contracts"
    });