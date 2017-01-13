package main

import (
	"flag"
	"hyperchain/utils/builtin"
	"strings"
)

var option = flag.String("o", "", "Two option:(1)account: create account  (2)transaction: create normal transaction or contract transaction")
var silense = flag.Bool("s", false, "use to mask output")

// account releated
var password = flag.String("p", "", "use to specify account password")

// transaction related
var from = flag.String("from", "", "use to specify transaction sender address")
var to = flag.String("to", "", "use to specify transaction receiver address")
var timestamp = flag.Int64("timestamp", 0, "use to specify transaction's timestamp, use the current system time if no value been passed")
var amount = flag.Int64("v", -1, "use to specify transaction transfer amount")
var payload = flag.String("l", "", "use to specify transaction payload")
var t = flag.Int("t", 0, "use to specify transaction type to generate, 0 represent normal transaction, 1 represent contract transaction")
var ip = flag.String("ip", "localhost", "use to specify the ip address")
var port = flag.Int("port", 8081, "use to specify the server port")

// stress test related
var nodeFile = flag.String("nodefile", "./nodes.json", "use to specify node address configuration file path")
var duration = flag.Int("duration", 0, "use to specify the duration of the stress test")
var tps = flag.Int("tps", 0, "use to specify the tps during the stress test")
var instant = flag.Int("instant", 0, "use to specify the instantaneous transaction number")
var testType = flag.Int("type", 0, "use to specify the stress test type, 0 represent normal transaction test, 1 represent contract transaction test, 2 represent normal and contract mix test")
var ratio = flag.Float64("ratio", 0.5, "use to specify the normal transaction's proportion, 1 - ratio represent the contract transaction's proportion")
var code = flag.String("code", "", "use to specify the contract bin code during the stress test")
var methoddata = flag.String("invoke_payload", "", "use to specify the contract invocation payload during the stress test")
var normalTxNum = flag.Int("rand_normal_tx", 0, "use to specify the number of random normal transaction which used in stress test")
var contractTxNum = flag.Int("rand_contract_tx", 0, "use to specify the number of random contract transaction which used in stress test")
var contractNum = flag.Int("rand_contract", 0, "use to specify the number of random contract which used in stress test")
var load = flag.Bool("load", false, "use the generated transaction saved in the file")
var estimation = flag.Int("e", 0, "use to specify the statistic estimation")
var accountNum = flag.Int("acct_num", 0, "use to specify the account num")
var simulate = flag.Bool("simulate", false, "use to specify the simulate")

func main() {
	flag.Parse()
	if *option == "" {
		flag.PrintDefaults()
	} else if strings.ToLower(*option) == "account" {
		builtin.NewAccount(*password, *silense)
	} else if strings.ToLower(*option) == "transaction" {
		builtin.NewTransaction(*password, *from, *to, *timestamp, *amount, *payload, *t, *ip, *port, *silense, *simulate)
	} else if strings.ToLower(*option) == "execute_transaction" {
		builtin.ExecuteTransaction(*password, *from, *to, *timestamp, *amount, *payload, *t, *ip, *port, *silense, *simulate)
	} else if strings.ToLower(*option) == "nh" {
		builtin.NHBANK(*tps, *instant, *duration, *estimation, *simulate)
	} else if strings.ToLower(*option) == "accounts" {
		builtin.NewAccounts(*accountNum, "123", false)
	}
}
