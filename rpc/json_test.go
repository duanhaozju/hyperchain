package jsonrpc

import (
	"testing"
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"encoding/json"
	"strconv"
	"fmt"
	"net/http"
)

type rwc struct {
	*bufio.ReadWriter
}

func (rwc *rwc) Close() error {
	return nil
}

func TestJsonCodecImpl_ReadRawRequest(t *testing.T) {
	ast := assert.New(t)
	req := bytes.NewBufferString(`{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`)
	reply := bytes.NewBufferString("")

	rwc := &rwc{bufio.NewReadWriter(bufio.NewReader(req), bufio.NewWriter(reply))}

	codec := NewJSONCodec(rwc, nil, nil, nil)

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
}

func TestJsonCodecImpl_GetAuthInfo(t *testing.T) {

	ast := assert.New(t)
	req := bytes.NewBufferString(`{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`)
	reply := bytes.NewBufferString("")

	rwc := &rwc{bufio.NewReadWriter(bufio.NewReader(req), bufio.NewWriter(reply))}

	h := make(http.Header)
	h.Set("Authorization", "authorization")
	h.Set("Method", "method")

	httpRequest := &http.Request{
		Header: h,
	}

	codec := NewJSONCodec(rwc, httpRequest, nil, nil)

	auth, method := codec.GetAuthInfo()
	ast.Equal("authorization", auth,
		fmt.Sprintf("Expected 'authorization', but got %v", auth))
	ast.Equal("method", method,
		fmt.Sprintf("Expected 'method', but got %v", method))
}