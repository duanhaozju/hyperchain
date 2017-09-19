package guomi

import "testing"

// 测试
func Test_Hash(t *testing.T) {
	h := New()
	h.Write([]byte("abc"))
	hashData := h.Sum(nil)
	t.Logf("%x \n", hashData)
	t.Log(hashData)
}
