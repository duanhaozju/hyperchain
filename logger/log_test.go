package logger

import (
	"testing"
)

func TestOpenLogFile(t *testing.T) {
	NewLogger("8002")

	O.Println("hahahahah")

}