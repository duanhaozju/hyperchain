package utils
import(
	"github.com/syndtr/goleveldb/leveldb"
)
type DBPath struct {
	Base string
	Path string
}
//var DBPATH  =  "../../db"

var DBPATH = DBPath{
	//Base: dir,_ := os.Getwd(); dir,  //-- 初始化为执行程序的
	Base: GetBasePath()+"/db/",
}


func (dbpath DBPath)String() string {
	return dbpath.Base + dbpath.Path
}


func GetConnection()(*leveldb.DB,error){
	db, err := leveldb.OpenFile(DBPATH.String(), nil )
	return db, err
}

