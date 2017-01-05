package hyperdb


import (
"github.com/garyburd/redigo/redis"
"fmt"
)
//迭代每次取的大小
var Size=10
type IteratorImp struct{
	Pool *redis.Pool
	Listkey []string
	Listvalue []string
	numlen int
	cu_num int
	start []byte
	end []byte
	err error
}

func (self *IteratorImp) Key() []byte{
	return []byte(self.Listkey[self.cu_num])
}

func (self *IteratorImp) Value() []byte{
	return []byte(self.Listvalue[self.cu_num])
}



func (self *IteratorImp) Seek(key []byte) bool {

	start:=make([]byte,len(key))
	copy(start,key)
	self.start=start
	self.end=append(start,byte(255))
	self.numlen=0;
	//先next
	self.cu_num=-1;

	con:=self.Pool.Get()
	defer con.Close()

	value,err:=redis.String(con.Do("get",string(key)))
	//如果key本身存在，取到key
	if err==nil{
		self.Listkey[0]=string(key)
		self.Listvalue[0]=value
		self.numlen++;
	}else if err.Error()!="redigo: nil returned"{
		self.err=err
		return false
	}
	resp,err:=redis.Strings(con.Do("scan",self.start,self.end,Size-1))
	if len(resp)>0&&err==nil {
		i:=(len(resp))/2
		//j=0或1
		j:=self.numlen
		self.numlen+=i

		//从零开始算
		k:=self.numlen-1
		for ;k>=j;i--{
			self.Listkey[k]=resp[2*i-2]
			self.Listvalue[k]=resp[2*i-1]
			k--
		}
		return true

	}
	self.err=err
	return false
}

func (self *IteratorImp) seek(key []byte) bool {

	self.numlen=0;
	self.cu_num=0;
	con:=self.Pool.Get()
	defer con.Close()
	resp,err:=redis.Strings(con.Do("scan",key,self.end,Size-1))
	if len(resp)>0&&err==nil {
		i:=(len(resp))/2
		self.numlen+=i
		//从零开始算
		i=self.numlen-1
		for ;i>=0;i--{
			self.Listkey[i]=resp[2*i]
			self.Listvalue[i]=resp[2*i+1]
		}
		return true
	}
	self.err=err
	return false
}


func (self *IteratorImp) Next() bool {
	fmt.Println("nuM:")
	fmt.Println(self.cu_num)
	fmt.Println(self.numlen)
	if self.cu_num < self.numlen-1  {
		self.cu_num++
		return true
	}
	if self.cu_num > Size-3 {
		return self.seek([]byte(self.Key()))
	}else{
		return false
	}
}
func (self *IteratorImp) Prev() bool{
	fmt.Println(self.cu_num)
	if self.cu_num>0{
		self.cu_num--
		return true
	}
	return false
}

//因为不会go起来所以可以关闭连接也不用锁
func (self *IteratorImp) Release(){
	self.Pool.Close()
}
