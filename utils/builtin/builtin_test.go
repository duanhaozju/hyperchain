package builtin

import (
	"testing"
)

func TestReadAndWrit(t *testing.T) {
	var data []string
	for i := 0; i < 10; i += 1 {
		data = append(data, "123")
	}
	write("./test", data)
	cnt, ctx := read("./test", 15)
	t.Log(cnt, ctx)

}
