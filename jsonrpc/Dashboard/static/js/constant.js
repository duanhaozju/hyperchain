/**
 * Created by sammy on 16-9-15.
 */


angular
    .module('starter')
    .constant('ENV',{
        "API": "http://localhost:8081",
        "PATTERN": [
            {name: "pattern0", value: "contract Accumulator{ uint32 sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint32){ return sum; } function add2(uint32 num1,uint32 num2) returns(uint32,uint32){ sum = sum + num1; sum = sum + num2; return sum1,sum2; }"},
            {name: "pattern1", value: 'contract Accumulator{    uint32 sum = 0;   function increment(){         sum = sum + 1;     }      function getSum() returns(uint32){         return sum;     }   function add(uint32 num1,uint32 num2) {         sum = sum+num1+num2;     } }'},
            {name: "pattern3", value: 'contract weiwei{ struct NameList{ string name; address[] AddressList; } mapping(uint256 => NameList) nameListMap; uint256 length = 0; string[] stringList; string temp_name_list = ""; function weiwei(){} function setNameList(string temp_name,address temp_addr){ for(uint256 i=0;i<length;i++){ if(stringsEqual(nameListMap[i].name,temp_name)){ nameListMap[i].AddressList.push(temp_addr); return; } } nameListMap[length].name = temp_name; nameListMap[length].AddressList.push(temp_addr); length++; } function getNameListsByName(string temp_name) returns(address[]){ for(uint256 i=0;i<length;i++){ if(stringsEqual(nameListMap[i].name,temp_name)){ return nameListMap[i].AddressList; } } } function getNameLists() returns(string){ setNameList("hehe",msg.sender); temp_name_list = ""; for(uint256 i=0;i<length;i++){ temp_name_list = strConcat(temp_name_list,",",nameListMap[i].name); } return temp_name_list; } function stringsEqual(string storage _a, string memory _b) internal returns (bool) { bytes storage a = bytes(_a); bytes memory b = bytes(_b); if (a.length != b.length) return false; for (uint i = 0; i < a.length; i ++) if (a[i] != b[i]) return false; return true; } function strConcat(string _a, string _b, string _c) internal returns (string){ bytes memory _ba = bytes(_a); bytes memory _bb = bytes(_b); bytes memory _bc = bytes(_c); string memory abcde = new string(_ba.length + _bb.length + _bc.length); bytes memory babcde = bytes(abcde); uint k = 0; for (uint i = 0; i < _ba.length; i++) babcde[k++] = _ba[i]; for (i = 0; i < _bb.length; i++) babcde[k++] = _bb[i]; for (i = 0; i < _bc.length; i++) babcde[k++] = _bc[i]; return string(babcde); } }'},
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