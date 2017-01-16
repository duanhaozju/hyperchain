package hyperdb

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type IteratorImp struct {
	Pool      *redis.Pool
	ListKey   []string
	ListValue []string
	numLen    int
	cu_Num    int
	start     []byte
	end       []byte
	err       error
}

func (iteratorImp *IteratorImp) Key() []byte {
	return []byte(iteratorImp.ListKey[iteratorImp.cu_Num])
}

func (iteratorImp *IteratorImp) Value() []byte {
	return []byte(iteratorImp.ListValue[iteratorImp.cu_Num])
}

func (iteratorImp *IteratorImp) Seek(key []byte) bool {

	start := make([]byte, len(key))
	copy(start, key)
	iteratorImp.start = start
	iteratorImp.end = append(start, byte(255))
	iteratorImp.numLen = 0
	//先next
	iteratorImp.cu_Num = -1

	con := iteratorImp.Pool.Get()
	defer con.Close()

	value, err := redis.String(con.Do("get", string(key)))
	//如果key本身存在，取到key
	if err == nil {
		iteratorImp.ListKey[0] = string(key)
		iteratorImp.ListValue[0] = value
		iteratorImp.numLen++
	} else if err.Error() != "redigo: nil returned" {
		iteratorImp.err = err
		return false
	}
	resp, err := redis.Strings(con.Do("scan", iteratorImp.start, iteratorImp.end, ssdbIteratorSize-1))
	if len(resp) > 0 && err == nil {
		i := (len(resp)) / 2
		//j=0或1
		j := iteratorImp.numLen
		iteratorImp.numLen += i

		//从零开始算
		k := iteratorImp.numLen - 1
		for ; k >= j; i-- {
			iteratorImp.ListKey[k] = resp[2*i-2]
			iteratorImp.ListValue[k] = resp[2*i-1]
			k--
		}
		return true

	}
	iteratorImp.err = err
	return false
}

func (iteratorImp *IteratorImp) seek(key []byte) bool {

	iteratorImp.numLen = 0
	iteratorImp.cu_Num = 0
	con := iteratorImp.Pool.Get()
	defer con.Close()
	resp, err := redis.Strings(con.Do("scan", key, iteratorImp.end, ssdbIteratorSize-1))
	if len(resp) > 0 && err == nil {
		i := (len(resp)) / 2
		iteratorImp.numLen += i
		//从零开始算
		i = iteratorImp.numLen - 1
		for ; i >= 0; i-- {
			iteratorImp.ListKey[i] = resp[2*i]
			iteratorImp.ListValue[i] = resp[2*i+1]
		}
		return true
	}
	iteratorImp.err = err
	return false
}

func (iteratorImp *IteratorImp) Next() bool {
	//fmt.Println("nuM:")
	//fmt.Println(iteratorImp.cu_Num)
	//fmt.Println(iteratorImp.numLen)
	if iteratorImp.cu_Num < iteratorImp.numLen-1 {
		iteratorImp.cu_Num++
		return true
	}
	if iteratorImp.cu_Num > ssdbIteratorSize-3 {
		return iteratorImp.seek([]byte(iteratorImp.Key()))
	} else {
		return false
	}
}
func (iteratorImp *IteratorImp) Prev() bool {
	fmt.Println(iteratorImp.cu_Num)
	if iteratorImp.cu_Num > 0 {
		iteratorImp.cu_Num--
		return true
	}
	return false
}

//因为不会go起来所以可以关闭连接也不用锁
func (iteratorImp *IteratorImp) Release() {
	iteratorImp.Pool.Close()
}
