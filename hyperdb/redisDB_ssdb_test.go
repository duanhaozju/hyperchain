package hyperdb

import (
	"testing"

	"strconv"
)

func TestWrite(t *testing.T) {
	db, err := NewRdSdDb("8001", 2)
	if err != nil {
		t.Error("NewRdSdDb fail")
	}

	batch := db.NewBatch()
	for {
		for i:=0;i<200;i++{
			batch.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		}

		batch.Write()
	}

}
