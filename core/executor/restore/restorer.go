package restore

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"hyperchain/common"
	cm "hyperchain/core/common"
	"hyperchain/core/ledger/chain"
	"hyperchain/core/ledger/state"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
)

var (
	StateInvalidErr = errors.New("state invalid")
	FileNotExsitErr = errors.New("file not exist")
)

type Handler struct {
	conf *common.Config
	db   db.Database
	ns   string
}

func NewRestorer(conf *common.Config, db db.Database, ns string) *Handler {
	return &Handler{
		conf: conf,
		db:   db,
		ns:   ns,
	}
}

func (handler *Handler) Restore(sid string) error {
	sid = convertId(sid)
	if exist, _ := checkExist(handler.ns, sid, handler.conf); exist == false {
		return FileNotExsitErr
	}
	p := path.Join("snapshots", "SNAPSHOT_"+sid)
	sdb, err := hyperdb.NewDatabase(handler.conf, p, hyperdb.GetDatabaseType(handler.conf), handler.ns)
	if err != nil {
		return err
	}
	defer sdb.Close()

	meta, err := chain.GetSnapshotMetaFunc(sdb)
	if err != nil {
		return err
	}

	// check
	if res := checkIntegrity(common.Hex2Bytes(meta.MerkleRoot), meta.Height, sdb, handler.conf); !res {
		return StateInvalidErr
	}

	// clear up online db
	iter := handler.db.NewIterator([]byte(""))
	for iter.Next() {
		handler.db.Delete(iter.Key())
	}
	iter.Release()

	// apply ws
	entries := cm.RetrieveSnapshotFileds()
	for _, entry := range entries {
		iter := sdb.NewIterator([]byte(entry))
		for iter.Next() {
			handler.db.Put(iter.Key(), iter.Value())
		}
		iter.Release()
	}

	blk, err := chain.GetBlockFunc(sdb, common.Hex2Bytes(meta.BlockHash))
	if err != nil {
		return err
	}

	// recheck
	if res := checkIntegrity(blk.MerkleRoot, blk.Number, handler.db, handler.conf); !res {
		return StateInvalidErr
	}

	// apply blk and chain
	writeBatch := handler.db.NewBatch()
	if _, err := chain.PersistBlock(writeBatch, blk, false, false); err != nil {
		return err
	}
	if err := chain.UpdateChain(meta.Namespace, writeBatch, blk, false, false, false); err != nil {
		return err
	}
	if err := chain.UpdateGenesisTag(meta.Namespace, meta.Height, writeBatch, false, false); err != nil {
		return err
	}
	if err := writeBatch.Write(); err != nil {
		return err
	}
	_, tag := chain.GetGenesisTag(meta.Namespace)
	fmt.Println("current genesis tag, ", tag)
	return nil
}

func checkExist(ns, id string, conf *common.Config) (bool, error) {
	p := path.Join(cm.GetDatabaseHome(ns, conf), "snapshots", "SNAPSHOT_"+id)
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func checkIntegrity(merkleRoot []byte, height uint64, db db.Database, conf *common.Config) bool {
	stateDb, err := state.New(common.BytesToHash(merkleRoot), db, db, conf, height)
	if err != nil {
		return false
	}
	rehash, err := stateDb.RecomputeCryptoHash()
	if err != nil || rehash != common.BytesToHash(merkleRoot) {
		return false
	}
	return true
}

func convertId(id string) string {
	if strings.HasPrefix(id, "0x") {
		return id
	} else {
		id = "0x" + id
		return id
	}
}
