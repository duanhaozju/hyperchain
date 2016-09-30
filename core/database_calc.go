package core

import (
	"hyperchain/hyperdb"
	"time"

	"strconv"
	"os"
	"io"
	"path/filepath"
)

// CalcResponseCount calculate response count of a block for given blockNumber
// millTime is Millisecond
func CalcResponseCount(blockNumber uint64, millTime int64) (int64,float64){
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	blockHash, err := GetBlockHash(db, blockNumber)
	block, err := GetBlock(db, blockHash)
	var count int64 = 0
	for _, trans := range block.Transactions {
		if block.WriteTime - trans.TimeStamp <= millTime * int64(time.Millisecond) {
			count ++
		}
	}
	percent := float64(count)/100

	//fmt.Println("block number is",block.Number)
	//if block.Transactions!=nil{
	//	fmt.Println("batch time is ",(block.Timestamp-block.Transactions[0].TimeStamp)/int64(time.Millisecond))
	//}

	/*fmt.Println("commit time is ",(block.CommitTime-block.Timestamp)/int64(time.Millisecond))
	fmt.Println("write time is ",(block.WriteTime-block.CommitTime)/ int64(time.Millisecond))
	fmt.Println("evm time is ",(block.EvmTime-block.WriteTime)/ int64(time.Millisecond))*/
	//fmt.Println("evm time is ",(block.EvmTime-block.WriteTime)/ int64(time.Millisecond))
	return count,percent
}
//CalcCommitAVGTime calculates block average commit time
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcCommitBatchAVGTime(from,to uint64) (int64,int64) {
	if from > to {
		log.Error("from less than to")
		return -1,-1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var commit int64 = 0
	var batch int64 = 0
	for i := from; i <= to; i ++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1,-1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1,-1
		}
		commit += block.CommitTime - block.Timestamp
		if block.Transactions!=nil{
			batch += block.Timestamp - block.Transactions[0].TimeStamp
		}

	}
	num := int64(to-from+1)
	return commit/(num)/int64(time.Millisecond),batch/(num)/int64(time.Millisecond)

}
func CalTransactionSum()  uint64{
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum uint64 = 0
	height := GetHeightOfChain()
	for i:=uint64(1);i<=height;i++{
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return uint64(0)
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return uint64(0)
		}
		tmp := uint64(len(block.Transactions))
		//log.Info("block tx number is:",tmp)
		sum+=tmp
	}
	return sum
}

// CalcResponseAVGTime calculate response avg time of blocks
// whose blockNumber from 'from' to 'to', include 'from' and 'to'
// return : avg Millisecond
func CalcResponseAVGTime(from, to uint64) int64 {
	if from > to {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	var length int64 = 0
	for i := from; i <= to; i ++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1
		}
		for _, trans := range block.Transactions {
			sum += block.WriteTime - trans.TimeStamp
		}
		length += int64(len(block.Transactions))
	}

	if length == 0 {
		return 0
	} else {
		return sum / (length * int64(time.Millisecond))
	}


}

func CalcBlockAVGTime(from,to uint64) int64 {
	if from > to {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	for i := from; i <= to; i ++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1
		}
		sum+=(block.WriteTime-block.Timestamp)/ int64(time.Millisecond)
	}
	num := int64(to-from+1)
	if num == 0 {
		return 0
	} else {
		return sum / (num)
	}

}

func CalcEvmAVGTime(from, to uint64) int64 {
	if from > to {
		log.Error("from less than to")
		return -1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	var sum int64 = 0
	for i := from; i <= to; i ++ {
		blockHash, err := GetBlockHash(db, i)
		if err != nil {
			log.Error(err)
			return -1
		}
		block, err := GetBlock(db, blockHash)
		if err != nil {
			log.Error(err)
			return -1
		}
		sum+=(block.EvmTime-block.WriteTime)/ int64(time.Millisecond)
	}
	num := int64(to-from+1)
	if num == 0 {
		return 0
	} else {
		return sum / (num)
	}


}
func CalBlockGPS()error{
	db,err:= hyperdb.GetLDBDatabase()
	if err!=nil{
		log.Fatal(err)
	}
	height := GetHeightOfChain()
	genesis,_ := GetBlockByNumber(db,uint64(1))
	startTime := genesis.WriteTime
	startSec := time.Unix(startTime/int64(time.Second),0).Second()
	latest,_ := GetBlockByNumber(db,height)
	endTime := latest.WriteTime
	content :=[]string{}
	s := "start time: "+time.Unix(0,startTime).Format("2006-01-02 15:04:05")+" end time: "+time.Unix(0,endTime).Format("2006-01-02 15:04:05")+"\n"

	count :=0
	flag := true
	var avg float64 = 0
	for i:=uint64(1);i<=height;i++{
		block,_ := GetBlockByNumber(db,i)
		if block.WriteTime-startTime>int64(20*time.Second){
			break
		}
		if block.WriteTime-startTime>int64(10*time.Second){
			avg+=1
		}
	}
	avg = avg/10
	log.Info(avg,"----------------1----------------")
	s = s+"total blocks: "+strconv.FormatUint(height,10)+"      blocks per second: "+strconv.FormatFloat(avg,'f',2,32)+"\n"
	content=append(content,s)
	for i:=uint64(1);i<=height;i++{
		block,_ := GetBlockByNumber(db,i)
		//println(time.Unix(block.WriteTime / int64(time.Second), 0).Format("2006-01-02 15:04:05"),"********",block.Number)
		endSec := time.Unix(block.WriteTime/int64(time.Second),0).Second()
		if block.WriteTime>=startTime&&endSec-startSec==0{
			count++
			if i==height{
				current:= time.Unix(0,startTime).Format("2006-01-02 15:04:05")
				s = current+":"+strconv.Itoa(count)+" blocks generated"+"\n"
				content=append(content,s)
			}
			continue
		}
		current:= time.Unix(0,startTime).Format("2006-01-02 15:04:05")
		s = current+":"+strconv.Itoa(count)+" blocks generated"+"\n"
		content=append(content,s)
		flag = false
		if flag==false{
			startTime = block.WriteTime
			startSec = time.Unix(startTime/int64(time.Second),0).Second()
			count = 1
			flag = true
			if i==height{
				s = current+":"+strconv.Itoa(count)+" blocks generated"+"\n"
				content=append(content,s)
			}
		}

	}
	path :="/tmp/hyperchain/cache/statis/block_time_statis"
	return storeData(path,content)
}
func storeData(file string,content []string) error {
	index := time.Now().Unix()
	file = file + "_"+strconv.FormatInt(index,10)
	dir := filepath.Dir(file)
	_, err := os.Stat(dir)
	if !(err == nil || os.IsExist(err)){//file exists
		err = os.MkdirAll(dir, 0700)
		if err !=nil{
				return err
		}
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,0600)
	if err != nil {
		return err
	}
	for _,d:=range(content){
		n, err := f.Write([]byte(d))
		if err == nil && n < len([]byte(d)) {
			err = io.ErrShortWrite
		}

	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}