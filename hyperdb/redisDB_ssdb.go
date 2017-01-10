
package hyperdb

import "errors"
import "sync"
import "github.com/garyburd/redigo/redis"



type dbDatabaseImpl struct {
	redisDb *RsDatabase
	ssdbDb *SSDatabase
	db_status bool
}

func NewRdSdDb() (Database ,error){

	rsdb,err:=NewRsDatabase()
	if err!=nil&&rsdb!=nil{
		log.Noticef("NewRsDatabase(%v) fail. err is %v. \n",grpcPort,err.Error())
		return nil,err
	}


	ssdb, err:= NewSSDatabase()
	if err != nil&&ssdb!=nil {
		log.Noticef("NewNewSSDatabase(%v) fail. err is %v. \n", grpcPort, err.Error())
		return nil, err
	}

	return &dbDatabaseImpl{redisDb:rsdb,ssdbDb:ssdb,db_status:true},nil
}

func (db *dbDatabaseImpl)Put(key []byte, value []byte) error{

	if err:=db.check(); err!=nil{
		return err
	}

	err:=db.redisDb.Put(key,value)
	if err!=nil{
		return err
	}
	go db.ssdbDb.Put(key,value)
	return nil
}


func (db *dbDatabaseImpl)Get(key []byte) ([]byte, error){
	if err:=db.check(); err!=nil{
		return nil,err
	}

	data,err:=db.redisDb.Get(key)

	if len(data)!=0&&err==nil{
		return data,err
	}
	data,err=db.ssdbDb.Get(key)
	if len(data)!=0&&err==nil{
		return data,err
	}
	return nil,err
}

func (db *dbDatabaseImpl)Delete(key []byte) error{

	if err:=db.check(); err!=nil{
		return err
	}
	err:=db.redisDb.Delete(key)

	if err==nil{
		err=db.ssdbDb.Delete(key)
	}
	return err
}

func (db *dbDatabaseImpl)NewIterator(prefix []byte) (Iterator){
	return db.ssdbDb.NewIterator(prefix)
}
//关闭数据库是不安全的，因为有可能有线程在写数据库，如果做到安全要加锁
//此处仅设置状态关闭
func(db *dbDatabaseImpl)Close(){
	if err:=db.check(); err==nil{
		db.db_status=false
	}

}

func (db *dbDatabaseImpl)check()error{
	if db.db_status==false{
		log.Notice("DB has been closed")
		return errors.New("DB has been closed")
	}
	return nil
}

type DB_Batch struct {
	mutex sync.Mutex
	redisDb *RsDatabase
	ssdbDb *SSDatabase
	batch_status bool
	batch_map map[string] []byte
}

func (db *dbDatabaseImpl)NewBatch() Batch{
	if err:=db.check(); err!=nil{
		log.Notice("Bad operation:try to create a new batch with closed db")
		return nil
	}
	return &DB_Batch{
		redisDb:db.redisDb,
		ssdbDb:db.ssdbDb,
		batch_status:true,
		batch_map:make(map[string][]byte),
	}
}

func (batch *DB_Batch) Put(key, value []byte) error{
	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}

	value1:=make([]byte,len(value))
	copy(value1,value)
	batch.mutex.Lock()
	batch.batch_map[string(key)]=value1
	batch.mutex.Unlock()

	return nil
}

func (batch *DB_Batch)Delete(key []byte) error{
	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	batch.mutex.Lock()
	delete(batch.batch_map,string(key))
	batch.mutex.Unlock()

	return nil
}

func (batch *DB_Batch)Write() error{
	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}

	err:=batch.RdWrite(batch.redisDb.rdPool)

	if err==nil{
		msp2:=batch.batch_map
		batch.batch_map = make(map[string][]byte)
		go batch.SdWrite(batch.ssdbDb.rdPool,msp2)
	}
	return err
}

func (batch *DB_Batch)RdWrite(db *redis.Pool) error{
	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	num:=0;
	for {

		list := make([]string, 0, 20)

		for k, v := range batch.batch_map {
			list = append(list, string(k), string(v))
		}
		con :=db.Get()
		_, err:= con.Do("mset", list)
		con.Close()
		if err == nil {
			return nil
		}
		num++
		if IfLogStatus(){
			writeLog(`Redis con.Do("mset", list)`,num,err)
		}


		if err.Error()!="ERR Connection timed out"{
			return err
		}

		if num>=redisMaxConnectTimes{
			log.Error("Redis Write Batch  to DB failed and it may cause unexpected effects")
			return err
		}
	}

}

func (batch *DB_Batch)SdWrite(db *redis.Pool,map2 map[string] []byte) error{

	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	num:=0;
	for {
		list := make([]string, 0, 20)

		for k, v := range map2 {
			list = append(list, string(k), string(v))
		}
		con :=db.Get()
		_, err:= con.Do("mset", list)
		con.Close()
		if err == nil {
			return nil
		}
		num++
		if IfLogStatus(){
			writeLog(`SSDB con.Do("mset", list)`,num,err)
		}


		if err.Error()!="ERR Connection timed out"{
			return err
		}

		if num>=redisMaxConnectTimes{
			log.Error("SSDB Write Batch  to DB failed and it may cause unexpected effects")
			return err
		}

	}

}
