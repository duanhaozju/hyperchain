package main

import (
	"flag"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/p2p/hts/secimpl"
	"io/ioutil"
	"strconv"
	"time"
)

var year = flag.Int("y", 1, "License generation date's year")
var month = flag.Int("m", 1, "License generation date's month")
var day = flag.Int("d", 1, "License generation date's day")
var path = flag.String("p", "./", "Output Path")

func GenerateLicense() {
	// IMPORTANT this private
	flag.Parse()
	privateKey := string("TnrEP|N.*lAgy<Q&@lBPd@J/")
	idSuffix := string("Hyperchain")
	date := time.Date(*year, time.Month(*month), *day, 0, 0, 0, 0, time.UTC)
	timestamp := date.Unix()
	id, err := secimpl.TripleDesEncrypt8([]byte(strconv.FormatInt(timestamp, 10)+idSuffix), []byte(privateKey))
	if err != nil {
		fmt.Println("####################  Generate License Failed ###################")
		fmt.Println(err)
		return
	}
	licenseTemplate, err := ioutil.ReadFile("LICENSE-TEMP")
	if err != nil {
		fmt.Println("####################  Generate License Failed ###################")
		fmt.Println(err)
		return
	}
	ctx := "\n\t12. Identification: " + string(common.Bytes2Hex(id))
	ioutil.WriteFile(*path+"LICENSE", []byte(string(licenseTemplate)+ctx), 0644)
	fmt.Println("####################  Generate License Success ###################")
	fmt.Println("   Arguments:")
	fmt.Println("   \tExpired Date:", date)
	fmt.Println("   Output Path:")
	fmt.Println("   \t", *path+"LICENSE")
}
func main() {
	GenerateLicense()
}
