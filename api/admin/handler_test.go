package jsonrpc

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	username := "hpc"
	password := "hyperchain"
	reqJson := []byte("")
	urlStr := "http://127.0.0.1:8081/login"
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(reqJson))
	if err != nil {
		t.Error(err)
	}

	// mock: login with a blank Authorization area.
	req.Header.Set("Authorization", "")
	usr, pwd, err := admin.basicAuth(req)
	ast.EqualError(err, ErrNotLogin.Error(),
		"Authorization failed, should return ErrNotLogin error")

	// mock: login using an Authorization area with incorrect format.
	req.Header.Set("Authorization", "Basic-aHBjOmh5cGVyY2hhaW4=")
	usr, pwd, err = admin.basicAuth(req)
	ast.EqualError(err, ErrNotLogin.Error(),
		"Authorization failed, should return ErrNotLogin error")

	// mock: login using an Authorization area with incorrect data.
	req.Header.Set("Authorization", "Basic mock_data")
	usr, pwd, err = admin.basicAuth(req)
	ast.EqualError(err, ErrDecodeErr.Error(),
		"Decode failed, should return ErrDecodeErr error")

	// mock: login with complete basic auth Authorization.
	// set basic auth header
	req.SetBasicAuth(username, password)
	usr, pwd, err = admin.basicAuth(req)
	ast.Equal(username, usr, "Username should be hpc")
	ast.Equal(password, pwd, "Password should be hyperchain")
}
