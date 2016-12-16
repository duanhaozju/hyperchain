package hyperdb

import "github.com/seefan/gossdb"

import "github.com/syndtr/goleveldb/leveldb/util"

type IteratorImp struct{
	ssdb_pool *gossdb.Connectors
	listkey []string
	listvalue []string
	numlen int
	cu_num int

}

func (self *IteratorImp) Key() []byte{
	return []byte(self.listkey[self.cu_num])
}

func (self *IteratorImp) Value() []byte{
	return []byte(self.listvalue[self.cu_num])
}

func (self *IteratorImp) Valid() bool{
	return false
}


func (self *IteratorImp) First() bool {
	if len(self.listkey)==0{
		return false
	}
	self.cu_num=0;
	return true
}

func (self *IteratorImp) Last() bool {
	if len(self.listkey)==0{
		return false
	}
	self.cu_num=len(self.listkey)-1;
	return true
}


func (self *IteratorImp) Seek(key []byte) bool {
	self.numlen=0;
	self.cu_num=0;

	con,err:=self.ssdb_pool.NewClient()
	if err!=nil{
		log.Error("get new connection from ssdb_pool fair")
		return false
	}
	defer con.Close()

	value,err:=con.Get(string(key))
	if err==nil{
		self.listkey[0]=string(key)
		self.listvalue[0]=value.String()
		self.numlen++;
	}
	resp,err:=con.Do("scan",key,"",49)
	if len(resp) > 0 && resp[0] == "ok"{
	//resp[0]=ok [1]=key1 [2]=value1 ........
		i:=(len(resp)-1)/2
		self.numlen+=i
		for ;i>0;i--{
			self.listkey[i]=resp[2*i-1]
			self.listvalue[i]=resp[2*i]
		}
		return true
	}
	return false
}

func (self *IteratorImp) seek(key []byte) bool {
	self.numlen=0;
	self.cu_num=0;

	con,err:=self.ssdb_pool.NewClient()
	if err!=nil{
		log.Error("get new connection from ssdb_pool fair")
		return false
	}
	defer con.Close()


	resp,err:=con.Do("scan",key,"",49)
	if len(resp) > 0 && resp[0] == "ok"{
		//resp[0]=ok [1]=key1 [2]=value1 ........
		i:=(len(resp)-1)/2
		self.numlen+=i
		for ;i>0;i--{
			self.listkey[i-1]=resp[2*i-1]
			self.listvalue[i-1]=resp[2*i]
		}
		return true
	}
	return false
}
func (self *IteratorImp) Next() bool{
	if self.cu_num<self.numlen{
		self.cu_num++
		return true
	}
	return self.seek([]byte(self.listkey[self.cu_num]))
}

func (self *IteratorImp) Prev() bool{
	if self.cu_num>0{
		self.cu_num--
		return true
	}
	return false
}


func (self *IteratorImp) Error() error  { return nil }

func(self *IteratorImp) SetReleaser(releaser util.Releaser){

}

func (self *IteratorImp) Release(){}
