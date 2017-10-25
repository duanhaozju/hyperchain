package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/op/go-logging"
	"hyperchain/common"
	hcom "hyperchain/hyperdb/common"
	dbutils "hyperchain/hyperdb/db"
	"strconv"
	"time"
)

const (
	module = "hyperdb/cassandra"
)

type CassandraClient struct {
	client    *gocql.Session
	logger    *logging.Logger
	namespace string
}

type CassandraBatch struct {
	batch  *gocql.Batch
	client *gocql.Session
	logger *logging.Logger
}

type CassandraIter struct {
	client *gocql.Session
	logger *logging.Logger
	iter   *gocql.Iter
	key    []byte
	value  []byte
}

func NewCassandra(conf *common.Config, namespace string) (*CassandraClient, error) {
	log := common.GetLogger(namespace, module)
	// connect to the cluster
	num := conf.GetInt(hcom.NODE_NUM)
	nodes := make([]string, num)
	for i := 0; i < num; i++ {
		nodes[i] = conf.GetString(hcom.NODE_prefix + strconv.Itoa(i+1)) //ip num from 1 so add 1
	}
	cluster := gocql.NewCluster(nodes...)
	cluster.Keyspace = "example"
	cluster.Consistency = gocql.One
	//set connect pool num  default is =2
	cluster.NumConns = 3

	session, err := cluster.CreateSession()
	if err != nil {
		log.Errorf("init cassandra session failed with err : %v", err.Error())
		return nil, err
	}
	time.Sleep(1 * time.Second) //Sleep so the fillPool can complete.

	return &CassandraClient{
		client:    session,
		logger:    log,
		namespace: namespace,
	}, nil
}
func (cc *CassandraClient) Close() {
	cc.client.Close()
}
func (cc *CassandraClient) Put(key []byte, value []byte) error {
	return cc.client.Query(`INSERT INTO hyperdb (key, value) VALUES (?,?)`, key, value).Exec()
}

func (cc *CassandraClient) Get(key []byte) ([]byte, error) {
	var value string
	err := cc.client.Query(`SELECT  value from hyperdb WHERE key = ?`, key).Consistency(gocql.One).Scan(&value)
	if err != nil && err.Error() == "not found" {
		return nil, dbutils.DB_NOT_FOUND
	}
	return []byte(value), err
}

func (cc *CassandraClient) Delete(key []byte) error {
	return cc.client.Query(`DELETE FROM hyperdb WHERE key = ?`, key).Exec()
}

func (cc *CassandraClient) NewBatch() dbutils.Batch {
	batch := cc.client.NewBatch(gocql.UnloggedBatch)
	return &CassandraBatch{
		batch:  batch,
		client: cc.client,
		logger: cc.logger,
	}
}

func (cc *CassandraClient) NewIterator(prefix []byte) dbutils.Iterator {
	iter := cc.client.Query(`SELECT * FROM hyperdb WHERE key >=? and key<? allow FILTERING ;`, prefix, dbutils.BytesPrefix(prefix)).Iter()
	return &CassandraIter{
		iter:   iter,
		client: cc.client,
		logger: cc.logger,
	}
}
func (cc *CassandraClient) Scan(begin, end []byte) dbutils.Iterator {

	iter := cc.client.Query(`SELECT * FROM hyperdb WHERE key >=? and key<? allow FILTERING ;`, begin, dbutils.BytesPrefix(end)).Iter()
	return &CassandraIter{
		iter:   iter,
		client: cc.client,
		logger: cc.logger,
	}
}
func (cc *CassandraClient) Namespace() string {
	return cc.namespace
}

func (cc *CassandraClient) MakeSnapshot(string, []string) error {
	return nil
}

func (cb *CassandraBatch) Put(key, value []byte) error {
	cb.batch.Query(`INSERT INTO hyperdb (key, value) VALUES (?,?)`, key, value)
	return nil
}

func (cb *CassandraBatch) Delete(key []byte) error {
	cb.batch.Query(`DELETE FROM hyperdb WHERE key = ?`, key)
	return nil
}

func (cb *CassandraBatch) Write() error {
	err := cb.client.ExecuteBatch(cb.batch)
	if err != nil {
		cb.logger.Error("commit failed", err.Error())
	}
	return err
}

func (cb *CassandraBatch) Len() int {
	return len(cb.batch.Entries)
}

func (ci *CassandraIter) Key() []byte {
	return ci.key
}

func (ci *CassandraIter) Value() []byte {
	return ci.value
}

func (ci *CassandraIter) Seek(key []byte) bool {
	ci.iter = ci.client.Query(`SELECT * FROM hyperdb WHERE key >=? and key<? allow FILTERING ;`, key, dbutils.BytesPrefix(key)).Iter()
	return true
}

func (ci *CassandraIter) Next() bool {
	return ci.iter.Scan(&ci.key, &ci.value)
}

func (ci *CassandraIter) Prev() bool {
	return false
}

func (ci *CassandraIter) Release() {
	ci.iter.Close()
}

func (ci *CassandraIter) Error() error {
	return nil
}
