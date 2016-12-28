package manager

import (
	"time"
	"io/ioutil"
	"regexp"
	"strings"
	"errors"
	"hyperchain/p2p/transport"
	"hyperchain/common"
	"strconv"
)

// checkLicense check whether current date time is larger than expired time in commercial license
func checkLicense(licensePath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Invalid License Cause a Panic")
		}
	}()
	dateChecker := func(now, expire time.Time) bool {
		return now.Before(expire)
	}

	privateKey := string("TnrEP|N.*lAgy<Q&@lBPd@J/")
	identificationSuffix := string("Hyperchain")

	license, err := ioutil.ReadFile(licensePath)
	if err != nil {
		err = errors.New("No License Found")
		return
	}
	pattern, _ := regexp.Compile("Identification: (.*)")
	identification := pattern.FindString(string(license))[16:]

	ctx, err := transport.TripleDesDecrypt(common.Hex2Bytes(identification), []byte(privateKey))
	if err != nil {
		err = errors.New("Invalid License")
		return
	}
	plainText := string(ctx)
	suffix := plainText[len(plainText)-len(identificationSuffix):]
	if strings.Compare(suffix, identificationSuffix) != 0 {
		err = errors.New("Invalid Identification")
		return
	}
	timestamp, err := strconv.ParseInt(plainText[:len(plainText)-len(identificationSuffix)], 10, 64)
	if err != nil {
		err = errors.New("Invalid License Timestamp")
		return
	}
	expiredTime := time.Unix(timestamp, 0)
	currentTime := time.Now()
	if validation := dateChecker(currentTime, expiredTime); !validation {
		err = errors.New("License Expired")
		return
	}
	return
}
func (self *ProtocolManager) checkExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			if err := checkLicense(self.licensePath); err != nil {
				log.Error("license expired. If you need to request to continue using the hyperchain service, please contact us via email: admin@hyperchain.cn")
				self.expired <- true
			}
		}
	}
}
