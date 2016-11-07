package certification

import (
	"io/ioutil"
)

func ReadCertFile(path string) (cert string) {
	filecontent, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
	}
	return string(filecontent)
}
