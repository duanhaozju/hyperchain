package builtin

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	logger         *logging.Logger // package-level logger
	encryption     = crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash     = crypto.NewKeccak256Hash("keccak256")
	normalTxPool   []string
	contractTxPool []string
	genesisAccount []string
	contract       []string
	globalNodes    []string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	logger = logging.MustGetLogger("builtin")
	initGenesis()
}

const (
	keystore           = "./keystore"
	transferUpperLimit = 100
	transferLowerLimit = 0
	defaultGas         = 10000
	defaultGasPrice    = 10000
	timestampRange     = 10000000000
	genesisPassword    = "123"

	normalTransactionNumber   = 10
	contractTransactionNumber = 10
	contractNumber            = 10
	payload                   = "0x60606040526000805463ffffffff1916815560ae908190601e90396000f3606060405260e060020a60003504633ad14af381146030578063569c5f6d14605e578063d09de08a146084575b6002565b346002576000805460e060020a60243560043563ffffffff8416010181020463ffffffff199091161790555b005b3460025760005463ffffffff166040805163ffffffff9092168252519081900360200190f35b34600257605c6000805460e060020a63ffffffff821660010181020463ffffffff1990911617905556"
	methodid                  = "0x569c5f6d"
)

func NewAccount(password string, silense bool) (string, bool) {
	if password == "" {
		output(silense, "Please enter your password")
		return "", false
	}
	am := accounts.NewAccountManager(keystore, encryption)
	account, err := am.NewAccount(password)
	if err != nil {
		output(silense, "Create Account Failed! Detail error message: ", err.Error())
		return "", false
	} else {
		output(silense, "================================ Create Account ========================================")
		output(silense, "Create new account! Your Account address is:", account.Address.Hex())
		output(silense, "Your Account password is:", password)
		return account.Address.Hex(), true
	}
}

func NewTransaction(password string, from string, to string, timestamp int64, amount int64, payload string, t int, ip string, port int, silense bool) (string, bool) {
	if password == "" {
		output(silense, "Please enter your password")
		if silense == false {
			flag.PrintDefaults()
		}
		return "", false
	}
	var _from string
	var _to string
	var _timestamp int64
	var _amount int64
	var success bool
	if from != "" {
		_from = from
	} else {
		_from, success = NewAccount(password, silense)
		if success == false {
			output(silense, "Create account failed!")
			return "", false
		}
	}
	if timestamp <= 0 {
		_timestamp = time.Now().UnixNano() - timestampRange
	} else {
		_timestamp = timestamp
	}

	if amount < 0 {
		_amount = generateTransferAmount()
	} else {
		_amount = amount
	}
	if to == "" {
		_to = generateAddress()
	} else {
		_to = to
	}
	am := accounts.NewAccountManager(keystore, encryption)
	if t == 0 {
		txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), _amount, nil)
		value, _ := proto.Marshal(txValue)
		tx := types.NewTransaction(common.HexToAddress(_from).Bytes(), common.HexToAddress(_to).Bytes(), value, _timestamp)
		signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), password)
		if err != nil {
			output(silense, "Create Transaction failed!, detail error message: ", err)
			return "", false
		}
		output(silense, "Create Normal Transaction Success!")
		output(silense, "Arugments:")
		output(silense, "\tfrom:", _from)
		output(silense, "\tto:", _to)
		output(silense, "\ttimestamp:", _timestamp)
		output(silense, "\tvalue:", _amount)
		output(silense, "\tsignature:", common.Bytes2Hex(signature))
		output(silense, "JSONRPC COMMAND:")
		command := fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"tx_sendTransaction\",\"params\":[{\"from\":\"%s\",\"to\":\"%s\",\"timestamp\":%d,\"value\":%d,\"signature\":\"%s\"}],\"id\":1}'", ip, port,_from, _to, _timestamp, _amount, common.Bytes2Hex(signature))
		output(silense, "\t", command)
		return command, true
	} else {
		_payload := common.StringToHex(payload)
		txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), 0, common.FromHex(payload))
		value, _ := proto.Marshal(txValue)
		var tx *types.Transaction
		if to == "" {
			tx = types.NewTransaction(common.HexToAddress(_from).Bytes(), nil, value, _timestamp)
		} else {
			tx = types.NewTransaction(common.HexToAddress(_from).Bytes(), common.HexToAddress(to).Bytes(), value, _timestamp)
		}
		signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), password)
		if err != nil {
			output(silense, "Create Transaction failed!, detail error message: ", err)
			return "", false
		}
		output(silense, "Create Contract Transaction Success!")
		output(silense, "ARGUMENTS:")
		output(silense, "\tfrom:", _from)
		output(silense, "\tto:", to)
		output(silense, "\ttimestamp:", _timestamp)
		output(silense, "\tpayload:", _payload)
		output(silense, "\tsignature:", common.Bytes2Hex(signature))
		output(silense, "JSONRPC COMMAND:")
		if to == "" {
			command := fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_deployContract\",\"params\":[{\"from\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\"}],\"id\":1}'", ip, port,_from, _timestamp, _payload, common.Bytes2Hex(signature))
			output(silense, "\t", command)
			return command, true
		} else {
			command := fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_invokeContract\",\"params\":[{\"from\":\"%s\", \"to\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\"}],\"id\":1}'", ip, port,_from, to, _timestamp, _payload, common.Bytes2Hex(signature))
			output(silense, "\t", command)
			return command, true
		}
	}
}

func generateAddress() string {
	var letters = []byte("abcdef0123456789")
	b := make([]byte, 40)
	for i := 0; i < 40; i += 1 {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "0x" + string(b)
}

func generateTransferAmount() int64 {
	return int64(rand.Intn(transferUpperLimit-transferLowerLimit) + transferLowerLimit)
}

func output(silense bool, msg ...interface{}) {
	if !silense {
		fmt.Println(msg...)
	}
}

func StressTest(nodeFile string, duration int, tps int, instant int, testType int, ratio float64, randNormalTx int, randContractTx int, randContract int, code string, methoddata string ,silense bool) bool {
	if tps == 0 || instant == 0 {
		output(silense, "invalid tps or instant parameter")
		return false
	}
	nodes, n, err := loadNodeInfo(nodeFile)
	if err != nil {
		return false
	}
	generateNormalTransaction(randNormalTx)
	generateContractTransaction(randContract, randContractTx, code, methoddata)
	start := time.Now().Unix()
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
	} else {
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
		if scanner.Text() == ""{
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

func generateNormalTransaction(n int) {
	logger.Notice("================ Generate Normal Transactions ================")
	if n == 0 {
		n = normalTransactionNumber
	}
	for i := 0; i < n; i += 1 {
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		receiver := generateAddress()
		amount := generateTransferAmount()
		command, success := NewTransaction(genesisPassword, sender, receiver, 0, amount, "", 0, "localhost", 8081, true)

		if success == false {
			logger.Error("create normal transaction failed")
		} else {
			pattern, _ := regexp.Compile(".*'(.*?)'")
			cmd := pattern.FindStringSubmatch(command)[1]
			normalTxPool = append(normalTxPool, cmd)
		}
	}
	logger.Notice(normalTxPool)
	logger.Noticef("================ %d Normal Transactions Generated ================", len(normalTxPool))
}

func generateContractTransaction(n, m int, code, methoddata string) {
	generateContract(n, code)
	logger.Notice("================ Generate Contract Transactions ================")
	if m == 0 {
		m = contractTransactionNumber
	}
	for i := 0; i < m; i += 1 {
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		if contract == nil {
			logger.Fatal("empty contract pool")
			return
		}
		receiver := contract[rand.Intn(len(contract))]
		if methoddata == "" {
			if code != ""{
				logger.Fatal("Please specify related invocation payload")
				continue
			} else {
				methoddata = methodid
			}
		}
		command, success := NewTransaction(genesisPassword, sender, receiver, 0, 0, methoddata, 1, "localhost", 8081, true)
		if success == false {
			logger.Error("create contract transaction failed")
		} else {
			pattern, _ := regexp.Compile(".*'(.*?)'")
			cmd := pattern.FindStringSubmatch(command)[1]
			contractTxPool = append(contractTxPool, cmd)
		}
	}
	logger.Notice(contractTxPool)
	logger.Noticef("================ %d Contract Transactions Generated ================", len(contractTxPool))
}

func generateContract(n int, code string) {
	logger.Notice("================ Generate Contract ================")
	if n == 0 {
		n = contractNumber
	}
	for i := 0; i < n; i += 1 {
		sender := genesisAccount[rand.Intn(len(genesisAccount))]
		if code == "" {
			code = payload
		}
		command, success := NewTransaction(genesisPassword, sender, "", 0, 0, code, 1, "localhost", 8081, true)
		if success {
			pattern, _ := regexp.Compile(".*'(.*)'")
			cmd := pattern.FindStringSubmatch(command)[1]
			resp, err := communication(cmd, globalNodes[0])
			if err != nil {
				logger.Errorf("Error: ", err)
			} else {
				pattern, _ := regexp.Compile(".*result...(.*)\"")
				txHash := pattern.FindStringSubmatch(string(resp))[1]
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
						contractAddr := pattern.FindStringSubmatch(string(resp))[1]
						contract = append(contract, contractAddr)
					}
				}
			}
		} else {
			logger.Fatal("Create deploy transaction failed")
		}
	}
	logger.Notice(contract)
	logger.Noticef("================ %d Contracts Generated ================", len(contract))
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
