package jsonrpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
	"io/ioutil"
	"github.com/hyperchain/hyperchain/crypto/primitives"
	nsmock "github.com/hyperchain/hyperchain/namespace/mocks"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/admittance"
	"crypto/ecdsa"
	"encoding/asn1"
	"crypto/rand"
	"math/big"
	"crypto/sha256"
	"github.com/hyperchain/hyperchain/hyperdb"
	"os"
)

func initCodec(request string) ServerCodec {
	req := bytes.NewBufferString(request)
	reply := bytes.NewBufferString("")

	rwc := &rwc{bufio.NewReadWriter(bufio.NewReader(req), bufio.NewWriter(reply))}

	h := make(http.Header)
	h.Set("Authorization", "authorization")
	h.Set("Method", "method")

	httpRequest := &http.Request{
		Header: h,
	}

	codec := NewJSONCodec(rwc, httpRequest, nil, nil)
	return codec
}

type rwc struct {
	*bufio.ReadWriter
}

func (rwc *rwc) Close() error {
	return nil
}

func TestJsonCodecImpl_ReadRawRequest(t *testing.T) {
	ast := assert.New(t)

	codec := initCodec(`{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`)
	defer codec.Close()

	requests, isBatch, err := codec.ReadRawRequest(OptionMethodInvocation)
	if err != nil {
		t.Fatalf("%v", err)
	}

	ast.False(isBatch, "This request is not a batch request")
	ast.Equal("global", requests[0].Namespace,
		fmt.Sprintf("Expected namespace 'global', but got %s", requests[0].Namespace))
	ast.Equal("test", requests[0].Service,
		fmt.Sprintf("Expected service 'test', but got %s", requests[0].Service))
	ast.Equal("hello", requests[0].Method,
		fmt.Sprintf("Expected method 'hello', but got %s", requests[0].Method))
	ast.False(requests[0].IsPubSub)

	if rawId, ok := requests[0].Id.(*json.RawMessage); ok {
		id, e := strconv.ParseInt(string(*rawId), 0, 64)
		if e != nil {
			t.Fatalf("%v", e)
		}
		if id != 1 {
			t.Fatalf("Expected id 1 but got %d", id)
		}
	} else {
		t.Fatalf("invalid request, expected *json.RawMesage got %T", requests[0].Id)
	}

	// test batch request
	codec = initCodec(`[{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}]`)
	_, isBatch, _ = codec.ReadRawRequest(OptionMethodInvocation)
	ast.True(isBatch, "This request is a batch request")

	// test invalid method name
	codec = initCodec(`{"jsonrpc": "2.0", "method": "invalid", "params": [], "id": 1, "namespace":"global"}`)
	_, _, err = codec.ReadRawRequest(OptionMethodInvocation)
	ast.NotNil(err)

	writeErr := codec.Write("response data")
	ast.Nil(writeErr)
}

func TestJsonCodecImpl_ReadRawRequest2(t *testing.T) {
	ast := assert.New(t)

	// send a subscribe request without event name
	codec := initCodec(`{"jsonrpc": "2.0", "method": "test_subscribe", "params": [], "id": 1, "namespace":"global"}`)
	defer codec.Close()

	_, _, err := codec.ReadRawRequest(OptionMethodInvocation|OptionSubscriptions)
	ast.NotNil(err, "should happen error, because there is no specified event name in params")

	// send a subscribe request with event name
	codec = initCodec(`{"jsonrpc": "2.0", "method": "test_subscribe", "params": ["block"], "id": 1, "namespace":"global"}`)
	requests, isBatch, err := codec.ReadRawRequest(OptionMethodInvocation|OptionSubscriptions)
	if err != nil {
		t.Fatalf("ReadRawRequest error: %v", err)
	}

	ast.False(isBatch, "This request is not a batch request")
	ast.Equal("test", requests[0].Service,
		fmt.Sprintf("Expected service 'test', but got %s", requests[0].Service))
	ast.Equal("block", requests[0].Method,
		fmt.Sprintf("Expected method 'block', but got %s", requests[0].Method))
	ast.True(requests[0].IsPubSub)

	// send a unsubscribe request without params
	codec = initCodec(`{"jsonrpc": "2.0", "method": "test_unsubscribe", "params": [], "id": 1, "namespace":"global"}`)
	requests, isBatch, err = codec.ReadRawRequest(OptionMethodInvocation|OptionSubscriptions)
	if err != nil {
		t.Fatalf("ReadRawRequest error: %v", err)
	}
	ast.False(isBatch, "This request is not a batch request")
	ast.Equal("test_unsubscribe", requests[0].Method,
		fmt.Sprintf("Expected method 'block', but got %s", requests[0].Method))
	ast.True(requests[0].IsPubSub)
}

func TestJsonCodecImpl_GetAuthInfo(t *testing.T) {

	ast := assert.New(t)

	codec := initCodec(`{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`)
	defer codec.Close()

	auth, method := codec.GetAuthInfo()
	ast.Equal("authorization", auth,
		fmt.Sprintf("Expected 'authorization', but got %v", auth))
	ast.Equal("method", method,
		fmt.Sprintf("Expected 'method', but got %v", method))
}

func TestJsonCodecImpl_CheckHttpHeaders(t *testing.T) {
	log = common.GetLogger("", "jsonrpc")

	req := bytes.NewBufferString(`{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`)
	reply := bytes.NewBufferString("")

	rwc := &rwc{bufio.NewReadWriter(bufio.NewReader(req), bufio.NewWriter(reply))}

	// read in tcert private key
	keyb, err := ioutil.ReadFile("test/certs/tcert.priv")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	priKey, err := primitives.ParseKey(keyb)
	if err != nil {
		t.Fatalf("ParseKey error: %v", err)
	}

	// sign the data
	data := "hyperchain"
	signature, err := ECDSASign(priKey, []byte(data))
	if err != nil {
		t.Fatalf("ECDSASign error: %v", err)
	}

	// read in tcert
	tcert, err := ioutil.ReadFile("test/certs/tcert.cert")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	// create mock object
	cfg := common.NewConfig("test/namespace.toml")
	err = hyperdb.InitDatabase(cfg, "global")
	if err != nil {
		t.Fatalf("InitDatabase error: %v", err)
	}
	camsgr, err := admittance.NewCAManager(cfg)
	if err != nil {
		t.Fatalf("NewCAManager error: %v", err)
	}
	mockNSMgr := &nsmock.MockNSMgr{}
	mockNs    := &nsmock.MockNS{}

	mockNSMgr.On("GetNamespaceByName", "").Return(mockNs)
	mockNs.On("GetCAManager").Return(camsgr)

	h := make(http.Header)
	h.Set("tcert", common.ToHex(tcert))
	h.Set("signature", common.ToHex(signature))
	h.Set("msg", common.TransportEncode(data))

	httpRequest := &http.Request{
		Header: h,
	}

	codec := NewJSONCodec(rwc, httpRequest, mockNSMgr, nil)
	defer func() {
		codec.Close()

		// delete test data
		if common.FileExist("namespaces") {
			t.Logf("hello %v", "workld")
			err := os.RemoveAll("namespaces")
			if err != nil {
				t.Fatalf("delet dir error: %v", err)
			}
		}
	}()
	if err := codec.CheckHttpHeaders("", "hello"); err != nil {
		t.Fatalf("CheckHttpHeaders error: %v", err)
	}
}

func TestJsonCodecImpl_CreateResponse(t *testing.T) {
	ast := assert.New(t)

	codec := initCodec(`{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`)
	defer codec.Close()

	resp := codec.CreateResponse(1, common.DEFAULT_NAMESPACE, "response data")
	if response, ok := resp.(*JSONResponse); ok {
		ast.Equal(1, response.Id)
		ast.Equal(common.DEFAULT_NAMESPACE, response.Namespace)
		ast.Equal(0, response.Code)
		ast.Equal("SUCCESS", response.Message)
		ast.Equal("response data", response.Result)
	} else {
		t.Fatalf("cannot convert type interface{} to *JSONResponse")
	}

	invalidParamErr := &common.InvalidParamsError{Message:"invalid params"}

	errResp := codec.CreateErrorResponse(1, common.DEFAULT_NAMESPACE, invalidParamErr)
	if response, ok := errResp.(*JSONResponse); ok {
		ast.Equal(1, response.Id)
		ast.Equal(common.DEFAULT_NAMESPACE, response.Namespace)
		ast.Equal(invalidParamErr.Code(), response.Code)
		ast.Equal(invalidParamErr.Error(), response.Message)
		ast.Nil(response.Result)
	} else {
		t.Fatalf("cannot convert type interface{} to *JSONResponse")
	}

	errRespWithInfo := codec.CreateErrorResponseWithInfo(1, common.DEFAULT_NAMESPACE, invalidParamErr, "others info")
	if response, ok := errRespWithInfo.(*JSONResponse); ok {
		ast.Equal(1, response.Id)
		ast.Equal(common.DEFAULT_NAMESPACE, response.Namespace)
		ast.Equal(invalidParamErr.Code(), response.Code)
		ast.Equal(invalidParamErr.Error(), response.Message)
		ast.Nil(response.Result)
		ast.Equal("others info", response.Info)
	} else {
		t.Fatalf("cannot convert type interface{} to *JSONResponse")
	}

	notification := codec.CreateNotification(common.ID("123456"), "test", "block", common.DEFAULT_NAMESPACE, "event data")
	if response, ok := notification.(*jsonNotification); ok {
		ast.Equal(common.DEFAULT_NAMESPACE, response.Namespace)

		ast.Equal("123456", response.Result.Subscription)
		ast.Equal("block", response.Result.Event)
		ast.Equal("event data", response.Result.Data)

	} else {
		t.Fatalf("cannot convert type interface{} to *jsonNotification")
	}
}

// ECDSASignature represents an ECDSA signature
type ECDSASignature struct {
	R, S *big.Int
}

func ECDSASign(signKey interface{}, msg []byte) ([]byte, error) {
	temp := signKey.(*ecdsa.PrivateKey)

	h := sha256.New()
	digest := make([]byte, 32)
	h.Write(msg)
	h.Sum(digest[:0])

	r, s, err := ecdsa.Sign(rand.Reader, temp, digest)
	if err != nil {
		return nil, err
	}

	raw, err := asn1.Marshal(ECDSASignature{r, s})
	if err != nil {
		return nil, err
	}

	return raw, nil
}

