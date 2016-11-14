package main

import (
	"flag"
	"strings"
	"hyperchain/utils/builtin"
)

var option = flag.String("o", "", "Two option:(1)account: create account  (2)transaction: create normal transaction or contract transaction")
var silense = flag.Bool("s", false, "use to mask output")

// account releated
var password = flag.String("p", "", "use to specify account password")

// transaction related
var from = flag.String("from", "", "use to specify transaction sender address")
var to = flag.String("to", "", "use to specify transaction receiver address")
var amount = flag.String("v", "", "use to specify transaction transfer amount")
var payload = flag.String("l", "", "use to specify transaction payload")
var t = flag.Int("t", 0, "use to specify transaction type to generate, 0 represent normal transaction, 1 represent contract transaction")
var ip = flag.String("ip", "localhost", "use to specify the ip address")

// stress test related
var nodeFile = flag.String("nodefile", "./nodes.json", "use to specify node address configuration file path, default is ./nodes.json")
var duration = flag.Int("duration", 0, "use to specify the duration of the stress test")
var tps = flag.Int("tps", 0, "use to specify the tps during the stress test")
var instant = flag.Int("instant", 0, "use to specify the instantaneous transaction number")
var testType = flag.Int("type", 0, "use to specify the stress test type, 0 represent normal transaction test, 1 represent contract transaction test, 2 represent normal and contract mix test")
var ratio = flag.Float64("ratio", 0.5, "use to specify the normal transaction's proportion, 1 - ratio represent the contract transaction's proportion")


func main() {
	flag.Parse()
	if *option == "" {
		flag.PrintDefaults()
	} else if strings.ToLower(*option) == "account" {
		builtin.NewAccount(*password, *silense)
	} else if strings.ToLower(*option) == "transaction" {
		builtin.NewTransaction(*password, *from, *to, *amount, *payload, *t, *ip,*silense)
	} else if strings.ToLower(*option) == "test" {
		builtin.StressTest(*nodeFile, *duration, *tps, *instant, *testType, *ratio, *silense)
	}
}

