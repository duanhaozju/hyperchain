package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestSeekAndAppend(t *testing.T) {
	filePath := "./test/resources/global.toml"
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
