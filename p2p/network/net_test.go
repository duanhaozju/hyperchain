package network

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseRoute(t *testing.T) {
	hostname1,port1,err1 := ParseRoute("node1:8081")
	assert.Equal(t,err1,nil)
	assert.Equal(t,hostname1,"node1")
	assert.Equal(t,port1,8081)

	hostname2,port2,err2 := ParseRoute("node1asdaq23qasdasdjaksdfhjkahsdfjkh:8081")
	assert.Equal(t,err2,nil)
	assert.Equal(t,hostname2,"node1asdaq23qasdasdjaksdfhjkahsdfjkh")
	assert.Equal(t,port2,8081)

	hostname3,port3,err3 := ParseRoute("node3:8000021223181")
	assert.Equal(t,err3,nil)
	assert.Equal(t,hostname3,"node3")
	assert.Equal(t,port3,8000021223181)

	hostname4,port4,err4 := ParseRoute("node3err")
	assert.Equal(t,err4,errInvalidRoute)
	assert.Equal(t,hostname4,"")
	assert.Equal(t,port4,0)

	hostname5,port5,err5 := ParseRoute("node3err:asd")
	assert.NotNil(t,err5)
	assert.Equal(t,hostname5,"")
	assert.Equal(t,port5,0)

}
