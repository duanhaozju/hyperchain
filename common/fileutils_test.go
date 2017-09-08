package common

import (
	"testing"
	"os"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"fmt"
	"strings"
)

func TestSeekAndAppend(t *testing.T) {
	filePath := "./testhelper/resources/global.toml"
	err := SeekAndAppend("[namespace.start]", filePath, fmt.Sprintf("    %s", "ns1 = true"))
	if err != nil {
		t.Error(err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, strings.Contains(string(data), "ns1 = true"))
}
