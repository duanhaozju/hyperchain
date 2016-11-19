package builtin

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func StressTest(nodeFile string, duration int, tps int, instant int, testType int, ratio float64, randNormalTx int, randContractTx int, randContract int, simulateNum int, code string, methoddata string, silense bool, simulate bool, load bool, estimation int) bool {
	if tps == 0 || instant == 0 {
		output(silense, "invalid tps or instant parameter")
		return false
	}
	_, globalAccounts = read(accountList, accountNumber)
	logger.Notice("accounts", globalAccounts)
	nodes, n, err := loadNodeInfo(nodeFile)
	if err != nil {
		return false
	}
	generateNormalTransaction(load, randNormalTx, simulate)
	generateContractTransaction(load, randContract, randContractTx, code, methoddata, simulate)
	generateSimulateTransaction(load, simulateNum, methoddata)
	start := time.Now().Unix()
	timeBegin := time.Now().UnixNano()
	end := start + int64(duration)
	var interval time.Duration
	tmp := float64(instant) / float64(tps)
	if tmp >= 1 {
		output(silense, "invalid tps or instant parameter")
		return false
	} else if tmp >= 0.001 && tmp < 1 {
		interval = time.Duration(tmp*1000) * time.Millisecond
	} else if tmp < 0.001 && tmp >= 0.001*0.001 {
		interval = time.Duration(tmp*1000*1000) * time.Microsecond
	} else {
		interval = time.Duration(tmp*1000*1000*1000) * time.Nanosecond
	}
	logger.Notice("=============== Stress Test Begin ===============")
	logger.Notice("=============== Arguments =======================")
	logger.Notice("\tTPS: ", tps)
	logger.Notice("\tInstant Concurrency: ", instant)
	logger.Notice("\tDuration: ", duration)
	var wg sync.WaitGroup
	for time.Now().Unix() < end {
		go sendBatch(nodes, n, instant, testType, ratio, wg)
		time.Sleep(interval)
	}
	wg.Wait()
	timeEnd := time.Now().UnixNano()
	logger.Notice("=============== Stress Test Finish ===============")
	logger.Notice("=============== Performance Behavior  =======================")
	time.Sleep(15 * time.Second)
	queryStatistic(timeBegin+int64(estimation)*time.Second.Nanoseconds(), timeEnd-int64(estimation)*time.Second.Nanoseconds(), globalNodes[2])
	return true
}

func sendBatch(nodes []string, n int, instant int, testType int, ratio float64, wg sync.WaitGroup) {
	for i := 0; i < instant/n; i += 1 {
		for j := 0; j < n; j += 1 {
			wg.Add(1)
			go sendRequest(nodes[j], testType, ratio, wg)
			time.Sleep(400 * time.Microsecond)
		}
	}
}
func sendRequest(address string, testType int, ratio float64, wg sync.WaitGroup) {
	client := http.Client{}
	var cmd string
	if testType == 0 {
		if normalTxPool == nil {
			logger.Fatal("empty normal transaction pool")
			return
		}
		cmd = normalTxPool[rand.Intn(len(normalTxPool))]
	} else if testType == 1 {
		if contractTxPool == nil {
			logger.Fatal("empty contract transaction pool")
			return
		}
		cmd = contractTxPool[rand.Intn(len(contractTxPool))]
	} else if testType == 2 {
		if normalTxPool == nil || contractTxPool == nil {
			logger.Fatal("empty contract transaction pool or normal transaction pool")
			return
		}
		if randomChoice(ratio) == 0 {
			// normal
			cmd = normalTxPool[rand.Intn(len(normalTxPool))]
		} else {
			// contract
			cmd = contractTxPool[rand.Intn(len(contractTxPool))]
		}
	} else if testType == 3 {
		// nonghang
		cmd, _ = NewTransaction(genesisPassword, globalAccounts[rand.Intn(len(globalAccounts))], NHcontract, time.Now().UnixNano(), 0, NHmethod2, 1, "", 0, false, true)
	} else {
		// mix transaction with normal contract tx and simulate contract tx
		if contractTxPool == nil || simulateTxPool == nil {
			logger.Fatal("empty contract transaction pool or simulate transaction pool")
			return
		}
		if randomChoice(ratio) == 0 {
			// normal
			cmd = contractTxPool[rand.Intn(len(contractTxPool))]
		} else {
			// contract
			cmd = simulateTxPool[rand.Intn(len(simulateTxPool))]
		}
	}
	var jsonStr = []byte(cmd)
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Error: ", err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Info("Result: ", string(body))
	}
	wg.Done()
}

func loadNodeInfo(nodeFile string) ([]string, int, error) {
	var nodes []string
	n := 0
	file, err := os.Open(nodeFile)
	if err != nil {
		logger.Fatal(err)
		return nil, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pattern, _ := regexp.Compile("http[s]?...([0-9]{2,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})|localhost.[0-9]{4}")
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		match := pattern.Match([]byte(scanner.Text()))
		if match == false {
			continue
		}
		nodes = append(nodes, scanner.Text())
		n += 1
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
		return nil, 0, err
	}
	globalNodes = nodes
	logger.Notice(nodes, n)
	return nodes, n, nil
}

func generateNormalTransaction(load bool, n int, simulate bool) {
	logger.Notice("================ Generate Normal Transactions ================")
	var cnt int
	var tmp []string
	var ret []string
	if n == 0 {
		n = normalTransactionNumber
	}
	if load {
		cnt, ret = read(normalTxStore, n)
		if cnt > 0 {
			normalTxPool = append(normalTxPool, ret...)
		}
	}
	for i := 0; i < n-cnt; i += 1 {
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		receiver := generateAddress()
		amount := generateTransferAmount()
		command, success := NewTransaction(genesisPassword, sender, receiver, 0, amount, "", 0, "localhost", 8081, simulate ,true)

		if success == false {
			logger.Error("create normal transaction failed")
		} else {
			pattern, _ := regexp.Compile(".*'(.*?)'")
			ret := pattern.FindStringSubmatch(command)
			if ret == nil || len(ret) < 2 {
				return
			}
			normalTxPool = append(normalTxPool, ret[1])
			tmp = append(tmp, ret[1])
		}
	}
	if len(tmp) > 0 {
		write(normalTxStore, tmp)
	}
	logger.Notice(normalTxPool)
	logger.Noticef("================ %d Normal Transactions Generated ================", len(normalTxPool))
}

func generateContractTransaction(load bool, n, m int, code, methoddata string, simulate bool) {
	var cnt int
	var ret []string
	var cnt2 int
	var tmp []string
	if n == 0 {
		n = contractNumber
	}
	if m == 0 {
		m = contractTransactionNumber
	}
	if load {
		cnt, ret = read(contractTxStore, m)
		contractTxPool = append(contractTxPool, ret...)

		cnt2, ret = read(contractStore, n)
		contract = append(contract, ret...)
	}
	generateContract(n-cnt2, code)
	logger.Notice("================ Generate Contract Transactions ================")
	for i := 0; i < m-cnt; i += 1 {
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		if contract == nil {
			logger.Fatal("empty contract pool")
			return
		}
		receiver := contract[rand.Intn(len(contract))]
		if methoddata == "" {
			if code != "" {
				logger.Fatal("Please specify related invocation payload")
				continue
			} else {
				methoddata = methodid
			}
		}
		command, success := NewTransaction(genesisPassword, sender, receiver, 0, 0, methoddata, 1, "localhost", 8081, simulate, true)
		if success == false {
			logger.Error("create contract transaction failed")
		} else {
			pattern, _ := regexp.Compile(".*'(.*?)'")
			ret := pattern.FindStringSubmatch(command)
			if ret == nil || len(ret) < 2 {
				return
			}
			contractTxPool = append(contractTxPool, ret[1])
			tmp = append(tmp, ret[1])
		}
	}
	write(contractTxStore, tmp)
	logger.Notice(contractTxPool)
	logger.Noticef("================ %d Contract Transactions Generated ================", len(contractTxPool))
}

func generateContract(n int, code string) {
	var tmp []string
	logger.Notice("================ Generate Contract ================")
	for i := 0; i < n; i += 1 {
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		if code == "" {
			code = payload
		}
		command, success := NewTransaction(genesisPassword, sender, "", 0, 0, code, 1, "localhost", 8081, false, true)
		if success {
			pattern, _ := regexp.Compile(".*'(.*)'")
			ret := pattern.FindStringSubmatch(command)
			if ret == nil || len(ret) < 2 {
				return
			}
			cmd := ret[1]
			resp, err := communication(cmd, globalNodes[0])
			if err != nil {
				logger.Errorf("Error: ", err)
			} else {
				pattern, _ := regexp.Compile(".*result...(.*)\"")
				ret := pattern.FindStringSubmatch(string(resp))
				if ret == nil || len(ret) < 2 {
					return
				}
				txHash := ret[1]
				// get receipt
				cmd := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"tx_getTransactionReceipt\",\"params\":[\"%s\"],\"id\":1}", txHash)
				time.Sleep(2 * time.Second)
				resp, err := communication(cmd, globalNodes[0])
				if err != nil {
					logger.Errorf("Error: ", err)
				} else {
					pattern, err := regexp.Compile(".*contractAddress...(.*?)\"")
					if err != nil {
						logger.Fatal(err)
					} else {
						ret := pattern.FindStringSubmatch(string(resp))
						if ret == nil || len(ret) < 2 {
							return
						}
						contract = append(contract, ret[1])
						tmp = append(tmp, ret[1])
					}
				}
			}
		} else {
			logger.Fatal("Create deploy transaction failed")
		}
	}
	write(contractStore, tmp)
	logger.Notice(contract)
	logger.Noticef("================ %d Contracts Generated ================", len(contract))
}

func generateSimulateTransaction(load bool, n int, methoddata string) {
	var cnt int
	var ret []string
	var tmp []string
	if n == 0 {
		n = simulateNumber
	}
	if load {
		cnt, ret = read(simulateStore, n)
		simulateTxPool = append(simulateTxPool, ret...)
	}
	for i := 0; i < n - cnt; i += 1 {
		// TODO
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		if contract == nil {
			logger.Fatal("empty contract pool")
			return
		}
		receiver := contract[rand.Intn(len(contract))]
		if methoddata == "" {
			methoddata = methodid
		}
		command, success := NewTransaction(genesisPassword, sender, receiver, 0, 0, methoddata, 1, "localhost", 8081, true, true)
		if success == false {
			logger.Error("create contract transaction failed")
		} else {
			pattern, _ := regexp.Compile(".*'(.*?)'")
			ret := pattern.FindStringSubmatch(command)
			if ret == nil || len(ret) < 2 {
				return
			}
			simulateTxPool = append(simulateTxPool, ret[1])
			tmp = append(tmp, ret[1])
		}
	}
	write(simulateStore, tmp)
	logger.Notice(simulateTxPool)
	logger.Noticef("================ %d Simulate Transactions Generated ================", len(simulateTxPool))
}
func initGenesis() {
	genesisAccount = append(genesisAccount, "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	genesisAccount = append(genesisAccount, "0xe93b92f1da08f925bdee44e91e7768380ae83307")
	genesisAccount = append(genesisAccount, "0x6201cb0448964ac597faf6fdf1f472edf2a22b89")
	genesisAccount = append(genesisAccount, "0xb18c8575e3284e79b92100025a31378feb8100d6")
}


func communication(cmd string, server string) ([]byte, error) {
	client := http.Client{}
	var jsonStr = []byte(cmd)
	req, err := http.NewRequest("POST", server, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Error: ", err)
		return nil, err
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return body, nil
	}
}

func randomChoice(ratio float64) int {
	var tmp []int
	for i := 0; i < int(ratio*10); i += 1 {
		tmp = append(tmp, 0)
	}
	for i := int(ratio * 10); i < 10; i += 1 {
		tmp = append(tmp, 1)
	}
	return tmp[rand.Intn(len(tmp))]
}

func read(path string, n int) (int, []string) {
	file, err := os.Open(path)
	if err != nil {
		return 0, nil
	}
	defer file.Close()
	var ret []string
	var count int = 0
	// TODO read from file randomly
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && count < n {
		ret = append(ret, scanner.Text())
		count += 1
	}
	if err := scanner.Err(); err != nil {
		return 0, nil
	}
	return len(ret), ret
}

func write(path string, data []string) error {
	var file *os.File
	var err error
	idx := strings.LastIndex(path, "/")
	if idx != -1 {
		subpath := path[:idx]
		err := os.MkdirAll(subpath, 0777)
		if err != nil {
			return err
		}
	}
	if _, e := os.Stat(path); os.IsNotExist(e) {
		file, err = os.Create(path)
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		file, err = os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
	}
	for _, entry := range data {
		_, err = file.WriteString(entry + "\n")
	}
	file.Close()
	return nil
}

func queryStatistic(begin, end int64, server string) {
	// Get TPS
	cmd := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"block_queryTPS\",\"params\":[{\"from\":%d,\"to\":%d}],\"id\":1}", begin, end)
	resp, err := communication(cmd, server)
	if err != nil {
		logger.Error("Get Statistic Data failed")
	} else {
		logger.Notice(string(resp))
	}
	// Get Latency
}
