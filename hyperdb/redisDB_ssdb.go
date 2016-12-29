
package hyperdb

import "errors"
import "sync"
import "github.com/garyburd/redigo/redis"
import "os"
import "fmt"
import "time"
import "strconv"

type dbDatabaseImpl struct {
	redis_db *RsDatabase
	ssdb_db *SSDatabase
	db_status bool
}

func NewRdSdDb(portDBPath string,ssdbnum int) (Database ,error){
	var rsdb *RsDatabase
	var ssdb *SSDatabase

	rsdb,err:=NewRsDatabase(portDBPath)
	if err!=nil{
		log.Noticef("NewRsDatabase(%v) fail. err is %v. \n",portDBPath,err.Error())
		return nil,err
	}


	ssdb, err = NewSSDatabase(portDBPath,ssdbnum)
	if err != nil {
		log.Noticef("NewNewSSDatabase(%v) fail. err is %v. \n", portDBPath, err.Error())
		return nil, err
	}

	return &dbDatabaseImpl{redis_db:rsdb,ssdb_db:ssdb,db_status:true},nil
}

func (db *dbDatabaseImpl)Put(key []byte, value []byte) error{

	if err:=db.check(); err!=nil{
		return err
	}

	err:=db.redis_db.Put(key,value)
	if err!=nil{
		return err
	}
	go db.ssdb_db.Put(key,value)
	return nil
}


func (db *dbDatabaseImpl)Get(key []byte) ([]byte, error){
	if err:=db.check(); err!=nil{
		return nil,err
	}

	data,err:=db.redis_db.Get(key)

	if len(data)!=0&&err==nil{
		return data,err
	}
	data,err=db.ssdb_db.Get(key)
	if len(data)!=0&&err==nil{
		return data,err
	}
	return nil,err
}

func (db *dbDatabaseImpl)Delete(key []byte) error{

	if err:=db.check(); err!=nil{
		return err
	}
	err:=db.redis_db.Delete(key)

	if err==nil{
		go db.ssdb_db.Delete(key)
	}
	return err
}

func (db *dbDatabaseImpl)NewIterator(prefix []byte) (Iterator){
	return db.ssdb_db.NewIterator(prefix)
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
	redis_db *RsDatabase
	ssdb_db *SSDatabase
	batch_status bool
	batch_map map[string] []byte
}

func (db *dbDatabaseImpl)NewBatch() Batch{
	if err:=db.check(); err!=nil{
		log.Notice("Bad operation:try to create a new batch with closed db")
		return nil
	}
	return &DB_Batch{
		redis_db:db.redis_db,
		ssdb_db:db.ssdb_db,
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

	err:=batch.Rdwrite(batch.redis_db.rd_pool)

	if err==nil{
		msp2:=batch.batch_map
		batch.batch_map = make(map[string][]byte)
		go batch.Sdwrite(batch.ssdb_db.rd_pool,msp2)
	}
	return err
}

func (batch *DB_Batch)Rdwrite(db *redis.Pool) error{
	batch.mutex.Lock()

	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	num:=0;
	var err error
	for {
		list := make([]string, 0, 20)

		for k, v := range batch.batch_map {
			list = append(list, string(k), string(v))
		}
		con :=db.Get()
		_, err:= con.Do("mset", list)
		defer con.Close()
		if err == nil {
			break
		} else {
			num++
			f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY|os.O_CREATE, 0644)
			if err1 != nil {
				fmt.Println("1.txt file create failed. err: " + err.Error())
			} else {
				n, _ := f.Seek(0, os.SEEK_END)
				currentTime := time.Now().Local()
				newFormat := currentTime.Format("2006-01-02 15:04:05.000")
				str := portDBPath + newFormat + `con.Do("mset",list) :` + err.Error() +" num:"+strconv.Itoa(num)+"\n"
				_, err1 = f.WriteAt([]byte(str), n)
				f.Close()
			}
		}
		if err.Error()!="ERR Connection timed out"||num>=3{
			break
		}
	}



	batch.mutex.Unlock()
	return err
}

func (batch *DB_Batch)Sdwrite(db *redis.Pool,map2 map[string] []byte) error{
	batch.mutex.Lock()
	fmt.Println("go ssdb start ")



	if batch.batch_status==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	num:=0;
	var err error
	for {
		list := make([]string, 0, 20)

		for k, v := range map2 {
			list = append(list, string(k), string(v))
		}
		con :=db.Get()
		_, err:= con.Do("mset", list)
		defer con.Close()
		if err == nil {
			break
		} else {
			num++
			f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY|os.O_CREATE, 0644)
			if err1 != nil {
				fmt.Println("1.txt file create failed. err: " + err.Error())
			} else {
				n, _ := f.Seek(0, os.SEEK_END)
				currentTime := time.Now().Local()
				newFormat := currentTime.Format("2006-01-02 15:04:05.000")
				str := portDBPath + newFormat + `con.Do("mset",list) :` + err.Error() +" num:"+strconv.Itoa(num)+"\n"
				_, err1 = f.WriteAt([]byte(str), n)
				f.Close()
			}
		}
		if err.Error()!="ERR Connection timed out"||num>=3{
			break
		}
	}

	batch.mutex.Unlock()
	return err
}
