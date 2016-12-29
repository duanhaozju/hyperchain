package hyperdb

//import "errors"
//import "sync"
//import "github.com/garyburd/redigo/redis"
//import "os"
//import "fmt"
//import "time"
//import "strconv"

//type dbDatabaseImpl struct {
//	rsdb_status int
//	ssdb_status int
//	ledb_status int
//	redis_db *RsDatabase
//	ssdb_db *SSDatabase
//	level_db   *LDBDatabase
//	db_status bool
//}
//
//func NewDatabase(DBPath string ,portDBPath string,Db_type int) (Database ,error){
//	var rsdb *RsDatabase
//	var ssdb *SSDatabase
//	var leveldb *LDBDatabase
//	var rsdb_status ,ssdb_status ,ledb_status int
//	var err error
//	if Db_type==110{
//		rsdb_status=1
//		ssdb_status=1
//	}else if Db_type==001{
//		ledb_status=1
//	}else{
//		return nil,errors.New("incorrect Db_type")
//	}
//
//	if rsdb_status==1{
//		rsdb,err=NewRsDatabase(portDBPath)
//		if err!=nil{
//			log.Noticef("NewRsDatabase(%v) fail. err is %v. \n",portDBPath,err.Error())
//			return nil,err
//		}
//	}
//
//	if ssdb_status==1 {
//		ssdb, err = NewSSDatabase(portDBPath)
//		if err != nil {
//			log.Noticef("NewNewSSDatabase(%v) fail. err is %v. \n", portDBPath, err.Error())
//			return nil, err
//		}
//	}
//	if ledb_status==1 {
//		leveldb, err = NewLDBDataBase(DBPath)
//		if err != nil {
//			log.Noticef("NewLDBDataBase(%v) fail. err is %v. \n", DBPath, err.Error())
//			return nil, err
//		}
//	}
//	return &dbDatabaseImpl{rsdb_status:rsdb_status,ssdb_status:ssdb_status,ledb_status:ledb_status,redis_db:rsdb,ssdb_db:ssdb,level_db:leveldb,db_status:true},nil
//}
//
//func (db *dbDatabaseImpl)Put(key []byte, value []byte) error{
//
//	var err error
//	if err=db.check(); err!=nil{
//		return err
//	}
//	if db.rsdb_status==1{
//	err=db.redis_db.Put(key,value)
//	}
//
//	if db.ssdb_status==1{
//		go db.ssdb_db.Put(key,value)
//	}
//
//	if db.rsdb_status==1&&db.ledb_status==1{
//		go db.level_db.Put(key,value)
//	}else if db.ledb_status==1{
//		err=db.level_db.Put(key,value)
//	}
//
//	return err
//}
//
////先从redis找，如果没有再从leveldb或者ssdb查找
//func (db *dbDatabaseImpl)Get(key []byte) ([]byte, error){
//	var err error
//	var data []byte
//	if err=db.check(); err!=nil{
//		return nil,err
//	}
//	if db.rsdb_status==1{
//		data,err=db.redis_db.Get(key)
//	}
//
//	if len(data)==0&&db.ssdb_status==1{
//		data,err=db.ssdb_db.Get(key)
//	}
//
//	if len(data)==0&&db.ledb_status==1{
//		data,err=db.level_db.Get(key)
//	}
//
//	return data,err
//}
//
//func (db *dbDatabaseImpl)Delete(key []byte) error{
//	var err error
//	if err=db.check(); err!=nil{
//		return err
//	}
//	if db.rsdb_status==1{
//		err=db.redis_db.Delete(key)
//	}
//
//	if db.ssdb_status==1{
//		if db.rsdb_status==1{
//			go db.ssdb_db.Delete(key)
//		}
//		err=db.ssdb_db.Delete(key)
//	}
//
//	if db.ledb_status==1{
//		if db.rsdb_status==1{
//			go db.level_db.Delete(key)
//		}
//		err=db.level_db.Delete(key)
//	}
//
//	return err
//}
//
//func (db *dbDatabaseImpl)NewIteratorWithPrefix(prefix string) (Iterator error){
//	return &IteratorImp{}
//}
////关闭数据库是不安全的，因为有可能有线程在写数据库，如果做到安全要加锁
////此处仅设置状态关闭
//func(db *dbDatabaseImpl)Close(){
//	if err:=db.check(); err==nil{
//		db.db_status=false
//	}
//
//}
//
//func (db *dbDatabaseImpl)check()error{
//	if db.db_status==false{
//		log.Notice("DB has been closed")
//		return errors.New("DB has been closed")
//	}
//	return nil
//}
//
//type DB_Batch struct {
//	leveldb_batch Batch
//	mutex sync.Mutex
//	rsdb_status int
//	ssdb_status int
//	ledb_status int
//	redis_db *RsDatabase
//	ssdb_db *SSDatabase
//	level_db   *LDBDatabase
//	batch_status bool
//	batch_map map[string] []byte
//}
//
//func (db *dbDatabaseImpl)NewBatch() Batch{
//	if err:=db.check(); err!=nil{
//		log.Notice("Bad operation:try to create a new batch with closed db")
//		return nil
//	}
//	return &DB_Batch{
//		leveldb_batch:db.level_db.NewBatch(),
//		rsdb_status:db.rsdb_status,
//		redis_db:db.redis_db,
//		ssdb_status:db.ssdb_status,
//		ssdb_db:db.ssdb_db,
//		ledb_status:db.ledb_status,
//		level_db:db.level_db,
//		batch_status:true,
//		batch_map:make(map[string][]byte),
//	}
//}
//
//func (batch *DB_Batch) Put(key, value []byte) error{
//	if batch.batch_status==false{
//		log.Notice("batch has been closed")
//		return errors.New("batch has been closed")
//	}
//	if batch.rsdb_status==1||batch.ssdb_status==1{
//		value1:=make([]byte,len(value))
//		copy(value1,value)
//		batch.mutex.Lock()
//		batch.batch_map[string(key)]=value1
//		batch.mutex.Unlock()
//	}
//	if batch.ledb_status==1{
//		batch.leveldb_batch.Put(key,value)
//	}
//	return nil
//}
//
//func (batch *DB_Batch)Delete(key []byte) error{
//	if batch.batch_status==false{
//		log.Notice("batch has been closed")
//		return errors.New("batch has been closed")
//	}
//	batch.mutex.Lock()
//	delete(batch.batch_map,string(key))
//	batch.mutex.Unlock()
//
//	return nil
//
//}
//
//func (batch *DB_Batch)Write() error{
//	if batch.batch_status==false{
//		log.Notice("batch has been closed")
//		return errors.New("batch has been closed")
//	}
//	var err error
//	if batch.rsdb_status==1{
//		err=batch.write(batch.redis_db.rd_pool)
//	}
//	if batch.ledb_status==1{
//		if batch.rsdb_status==1{
//			go batch.leveldb_batch.Write()
//		}
//		err=batch.leveldb_batch.Write()
//	}
//	if batch.ssdb_status==1{
//		if batch.rsdb_status==1{
//			go batch.write(batch.ssdb_db.rd_pool)
//		}
//		err=batch.write(batch.redis_db.rd_pool)
//	}
//	return err
//}
//
//func (batch *DB_Batch)write(db *redis.Pool) error{
//	if batch.batch_status==false{
//		log.Notice("batch has been closed")
//		return errors.New("batch has been closed")
//	}
//	num:=0;
//	var err error
//
//	for {
//		list := make([]string, 0, 20)
//		con :=db.Get()
//
//		for k, v := range batch.batch_map {
//			list = append(list, string(k), string(v))
//		}
//		_, err:= con.Do("mset", list)
//		con.Close()
//		if err == nil {
//			batch.batch_map = make(map[string][]byte)
//			break
//		} else {
//			num++
//			f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY|os.O_CREATE, 0644)
//			if err1 != nil {
//				fmt.Println("1.txt file create failed. err: " + err.Error())
//			} else {
//				n, _ := f.Seek(0, os.SEEK_END)
//				currentTime := time.Now().Local()
//				newFormat := currentTime.Format("2006-01-02 15:04:05.000")
//				str := portDBPath + newFormat + `con.Do("mset",list) :` + err.Error() +" num:"+strconv.Itoa(num)+"\n"
//				_, err1 = f.WriteAt([]byte(str), n)
//				f.Close()
//			}
//		}
//		if err.Error()!="ERR Connection timed out"||num>=3{
//			break
//		}
//	}
//
//	return err
//}
//
