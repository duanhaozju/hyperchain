package main

import (
	"github.com/op/go-logging"
	"net/http"
	"io/ioutil"
	"bytes"
	"time"
)

var logger *logging.Logger // package-level logger
var address = [4]string{
	"http://123.206.218.227:8081",
	"http://123.206.218.201:8081",
	"http://123.206.216.149:8081",
	"http://123.206.216.163:8081",
}

func init() {
	logger = logging.MustGetLogger("performance")
}

func single(s string) {
	client := http.Client{}

	var jsonprep string = `{"jsonrpc":"2.0","method":"tx_sendTransaction","params":[{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":"0x01","timestamp":1477459062327000000}],"id":1}`

	var jsonStr = []byte(jsonprep)

	req, err := http.NewRequest("POST", s, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Error: ", err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Info("Result: ", string(body))
	}
}

func batch() {
	for i:=1; i<=125; i++ {
		go single(address[0])
		go single(address[1])
		go single(address[2])
		go single(address[3])
		time.Sleep(400 * time.Microsecond)
	}
}

func main() {
	start := time.Now().Unix()
	end := start + 500
	for i:=start; i<end; i=time.Now().Unix() {
		go batch()
		time.Sleep(50 * time.Millisecond)
	}
}