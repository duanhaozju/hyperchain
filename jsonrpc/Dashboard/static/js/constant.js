/**
 * Created by sammy on 16-9-15.
 */


angular
    .module('starter')
    .constant('ENV',{
        "API": "http://localhost:8081",
        "PATTERN": [
            {name: "pattern1", value: "contract Accumulator{ uint sum = 0; function increment(){ sum = sum + 1; } function getSum() returns(uint){ return sum; }}"},
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
            "function transfer(address addr1,address addr2,uint amount) returns (bool){" +
            "if(accounts[addr1] >= amount){" +
            "accounts[addr1] = accounts[addr1] - amount;" +
            "accounts[addr2] = accounts[addr2] + amount;" +
            "return true;" +
            "}" +
            "return false;" +
            "}" +
            "function getAccountBalance(address addr) returns(uint){" +
            "return accounts[addr];" +
            "}}"},
            {name: "pattern3", value: "contract InfoPlatform{" +
            "struct User{" +
            "string  name;   // the name of the user" +
            "uint    age;    // the age of the user" +
            "string  id;     // the id of the user" +
            "}" +
            "mapping(address => User) public users;" +
            "function setInformation(address addr,string name,uint age,string id){" +
            "User u = users[addr];" +
            "u.name = name;" +
            "u.age = age;" +
            "u.id = id;" +
            "}" +
            "function getInformation(address addr) returns (string,uint,string){" +
            "User u = users[addr];" +
            "return (u.name,u.age,u.id);" +
            "}}"
            }]
    })