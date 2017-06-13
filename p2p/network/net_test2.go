package network

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseRoute(t *testing.T) {
	hostname1,port1,err1 := ParseRoute("127.0.0.1:8081")
	assert.Equal(t,err1,nil)
	assert.Equal(t,"127.0.0.1",hostname1)
	assert.Equal(t,port1,8081)

	hostname2,port2,err2 := ParseRoute("172.16.10.10:8081")
	assert.Equal(t,err2,nil)
	assert.Equal(t,"172.16.10.10",hostname2)
	assert.Equal(t,port2,8081)

	hostname3,port3,err3 := ParseRoute("127.0.1.2:8000021223181")
	assert.Equal(t,err3,nil)
	assert.Equal(t,"127.0.1.2",hostname3)
	assert.Equal(t,port3,8000021223181)

	hostname4,port4,err4 := ParseRoute("node3err")
	assert.Equal(t,err4,errInvalidRoute)
	assert.Equal(t,"",hostname4)
	assert.Equal(t,port4,0)

	hostname5,port5,err5 := ParseRoute("node3err:asd")
	assert.NotNil(t,err5)
	assert.Equal(t,hostname5,"")
	assert.Equal(t,port5,0)

}
