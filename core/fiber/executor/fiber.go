package executor

import (
	"encoding/binary"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/common/service"
	pb "github.com/hyperchain/hyperchain/common/service/protos"
	"github.com/hyperchain/hyperchain/core/fiber"
	"github.com/hyperchain/hyperchain/core/oplog"
	"github.com/hyperchain/hyperchain/hyperdb"
	hc "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"strconv"
	"time"
)

const (
	consumeIndexPrefix = "last.consume.index."
	commitIndexPrefix  = "last.commit.index."
	executorId         = "EXECUTOR-0"
)

//Fiber response for log data transfer.
type ExeFiber struct {
	md               db.Database //meta data
	lastConsumeIndex uint64
	lastCommitIndex  uint64
	ns               *service.NamespaceServices
	ol               oplog.OpLog
	conf             *common.Config
	logger           *logging.Logger
}

func NewFiber(conf *common.Config, ns *service.NamespaceServices, ol oplog.OpLog) (fiber.Fiber, error) {
	namespace := conf.GetString(common.NAMESPACE)
	if len(namespace) == 0 {
		return nil, fmt.Errorf("no namespace field found in config")
	}

	fr := &ExeFiber{
		conf:   conf,
		ns:     ns,
		ol:     ol,
		logger: common.GetLogger(namespace, "fiber"),
	}
	var err error
	fr.md, err = hyperdb.GetOrCreateDatabase(conf, namespace, hc.DBNAME_META)
	if err != nil {
		return nil, err
	}
	if err = fr.recovery(); err != nil {
		return nil, err
	}
	return fr, nil
}

func (f *ExeFiber) recovery() error {
	if data, err := f.md.Get([]byte(consumeIndexPrefix)); err == nil {
		binary.Varint(data)
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return err
		}
		f.lastConsumeIndex = i
	} else {
		f.lastConsumeIndex = 0
		f.logger.Warningf("no last consume index found, set lastConsumeIndex to 0")
	}

	if data, err := f.md.Get([]byte(commitIndexPrefix)); err == nil {
		binary.Varint(data)
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return err
		}
		f.lastCommitIndex = i
	} else {
		f.lastCommitIndex = 0
		f.logger.Warningf("no last commit index found, set lastCommitIndex to 0")
	}
	return nil
}

func (f *ExeFiber) Start() error {
	var es service.Service
	for { //TODO(Xiaoyi Wang): add close related control
		es = f.ns.Service(executorId)
		f.logger.Debugf("%v executor service == nil? %v", f.ol.GetLastCommit(), es == nil)
		if es != nil && (f.lastConsumeIndex+1 <= f.ol.GetLastCommit()) {
			if e, err := f.ol.Fetch(f.lastConsumeIndex + 1); err == nil { //如果fetch不到应该阻塞该线程
				if payload, err := proto.Marshal(e); err != nil {
					f.logger.Errorf("unmarshal [%v] error %v", e, err)
				} else {
					logMsg := &pb.IMessage{
						Type:    pb.Type_OP_LOG,
						Payload: payload,
					}
					if err := es.Send(logMsg); err != nil {
						f.logger.Error(err)
						continue
					} else {
						f.lastConsumeIndex++ //TODO(Xiaoyi Wang): make this run in atomic way
					}
				}

			} else {
				time.Sleep(time.Second)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (f *ExeFiber) Stop() error {
	//TODO
	return nil
}
