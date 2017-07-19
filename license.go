package main

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/p2p/transport"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	LICENSE_PATH = "./LICENSE"
)

func CheckLicense(exit chan bool) {
	// this ensures that license checker always hit in `os thread` to avoid jmuping to other threads
	// since in this approach, working directory will not be affected by other operators.
	runtime.LockOSThread()
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			if expired := isLicenseExpired(); expired {
				notifySystemExit(exit)
				return
			}
		}
	}
}

// isLicenseExpired - check whether license is expired.
func isLicenseExpired() (expired bool) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("invalid license.")
			expired = true
		}
	}()

	dateChecker := func(now, expire time.Time) bool {
		return now.Before(expire)
	}

	privateKey := string("TnrEP|N.*lAgy<Q&@lBPd@J/")
	identificationSuffix := string("Hyperchain")
	license, err := ioutil.ReadFile(LICENSE_PATH)
	if err != nil {
		fmt.Println("no license found.")
		expired = true
		return
	}
	pattern, _ := regexp.Compile("Identification: (.*)")
	identification := pattern.FindString(string(license))[16:]
	ctx, err := transport.TripleDesDec([]byte(privateKey), common.Hex2Bytes(identification))
	if err != nil {
		fmt.Println("invalid license.")
		expired = true
		return
	}
	plainText := string(ctx)
	suffix := plainText[len(plainText)-len(identificationSuffix):]
	if strings.Compare(suffix, identificationSuffix) != 0 {
		fmt.Println("invalid license.")
		expired = true
		return
	}
	timestamp, err := strconv.ParseInt(plainText[:len(plainText)-len(identificationSuffix)], 10, 64)
	if err != nil {
		fmt.Println("invalid license.")
		expired = true
		return
	}

	expiredTime := time.Unix(timestamp, 0)
	if !dateChecker(time.Now(), expiredTime) {
		fmt.Println("license expired.")
		expired = true
		return
	}
	return
}

// notifySystemExit - license expired or not found, shut down system.
func notifySystemExit(exit chan bool) {
	exit <- true
}
