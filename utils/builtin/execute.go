package builtin

import (
	"regexp"
	"fmt"
	"time"
)

func ExecuteTransaction(password string, from string, to string, timestamp int64, amount int64, payload string, t int, ip string, port int, simulate bool, silense bool) (string, bool) {
	begin := time.Now()
	command, success := NewTransaction(password, from, to, timestamp, amount, payload, t, ip, port, simulate, silense)
	logger.Notice("Generate a transaction elapsed", time.Since(begin))
	var execRes string
	if success == false {
		logger.Error("create transaction failed")
		return "", false
	} else {
		pattern, _ := regexp.Compile(".*'(.*?)'")
		ret := pattern.FindStringSubmatch(command)
		if ret == nil || len(ret) < 2{
			logger.Error("create transaction failed")
			return "", false
		}
		cmd := ret[1]
		server := fmt.Sprintf("http://%s:%d", ip, port)
		resp, err := communication(cmd, server)
		if err != nil {
			logger.Error("execute transaction failed")
			return "", false
		} else {
			logger.Notice("Result:", string(resp))
			pattern, _ := regexp.Compile(".*result...(.*)\"")
			ret := pattern.FindStringSubmatch(string(resp))
			if ret == nil || len(ret) < 2 {
				logger.Error("DEBUG1")
				logger.Error("execute transaction failed")
				return "", false
			}
			txHash := ret[1]
			cmd := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"tx_getTransactionReceipt\",\"params\":[\"%s\"],\"id\":1}", txHash)
			time.Sleep(2 * time.Second)
			resp, err := communication(cmd, server)
			if err != nil {
				logger.Errorf("Error: ", err)
				return "", false
			} else {
				pattern, err := regexp.Compile(".*result.*")
				if err != nil {
					logger.Fatal(err)
					return "", false
				} else {
					ret := pattern.FindStringSubmatch(string(resp))
					if ret == nil || len(ret) < 1 {
						// invalid transaction
						cmd := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"tx_getTransactionByHash\",\"params\":[\"%s\"],\"id\":1}", txHash)
						resp, err := communication(cmd, server)
						if err != nil {
							logger.Errorf("Error: ", err)
							return "", false
						} else {
							ret := pattern.FindStringSubmatch(string(resp))
							if ret == nil || len(ret) < 1 {
								logger.Error("execute transaction failed")
								return "", false
							} else {
								output(silense,"Log:",ret[0])
								execRes = ret[0]
							}
						}
					} else {
						output(silense,"Receipt:",ret[0])
						execRes = ret[0]
					}
				}
			}

		}
	}
	return execRes, true
}
