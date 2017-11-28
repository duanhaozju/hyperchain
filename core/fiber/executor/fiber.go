package executor

import (
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
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/op/go-logging"
	"strconv"
	"sync/atomic"
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
	stop             int32
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
	fr.stop = 0
	return fr, nil
}

func (f *ExeFiber) recovery() error {
	if data, err := f.md.Get([]byte(consumeIndexPrefix)); err == nil {
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return err
		}
		atomic.StoreUint64(&f.lastConsumeIndex, i)
	} else {
		atomic.StoreUint64(&f.lastConsumeIndex, 0)
		f.logger.Warningf("no last consume index found, set lastConsumeIndex to 0")
	}

	if data, err := f.md.Get([]byte(commitIndexPrefix)); err == nil {
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return err
		}
		atomic.StoreUint64(&f.lastCommitIndex, i)
	} else {
		atomic.StoreUint64(&f.lastCommitIndex, 0)
		f.logger.Warningf("no last commit index found, set lastCommitIndex to 0")
	}
	return nil
}

func (f *ExeFiber) Start() error {
	var es service.Service
	atomic.StoreInt32(&f.stop, 0)

	for atomic.LoadInt32(&f.stop) == 0 {
		f.logger.Debugf("%v executor service == nil? %v", f.ol.GetLastCommit(), es == nil)

		nextConsumeIndex := atomic.LoadUint64(&f.lastConsumeIndex) + 1
		if nextConsumeIndex <= f.ol.GetLastCommit() {
			if e, err := f.ol.Fetch(nextConsumeIndex); err == nil {
				if payload, err := proto.Marshal(e); err != nil {
					f.logger.Errorf("unmarshal [%v] error %v", e, err)
				} else {
					logMsg := &pb.IMessage{
						Type:    pb.Type_OP_LOG,
						Payload: payload,
					}

					es = f.ns.Service(executorId)
					if es == nil {
						f.logger.Error("no executor service connection found, continue")
						time.Sleep(time.Second)
						continue
					}

					if err := es.Send(logMsg); err != nil {
						f.logger.Error(err)
						continue
					} else {
						atomic.StoreUint64(&f.lastConsumeIndex, nextConsumeIndex)
						if nextConsumeIndex%30 == 0 { //TODO(Xiaoyi Wang: how to store index effectively.
							buf := make([]byte, 0)
							b := strconv.AppendUint(buf, nextConsumeIndex, 10)
							f.md.Put([]byte(consumeIndexPrefix), b)
						}
					}
				}

			} else {
				time.Sleep(time.Second)
			}
		} else {
			time.Sleep(time.Second)
		}
	}

	return nil
}

func (f *ExeFiber) processExecutorRequest(exit chan bool) {
	var es service.Service
	atomic.StoreInt32(&f.stop, 0)
	for atomic.LoadInt32(&f.stop) == 0 {
		es = f.ns.Service(executorId)
		if es == nil {
			time.Sleep(time.Second)
		} else {
			select {
			case <-exit:
				return
			case req := <-es.Receive():
				f.handle(req)
			}
		}
	}
}

func (f *ExeFiber) handle(req *pb.IMessage) {
	switch req.Event {
	case pb.Event_OpLogAck:
		ack := &event.OpLogAck{}
		err := proto.Unmarshal(req.Payload, ack)
		if err != nil {
			f.logger.Error(err)
		} else {
			if ack.SeqNo > atomic.LoadUint64(&f.lastCommitIndex) {
				atomic.StoreUint64(&f.lastCommitIndex, ack.SeqNo)
				if ack.SeqNo%10 == 0 { // TODO(Xiaoyi Wang): make this more effectively
					buf := make([]byte, 0)
					b := strconv.AppendUint(buf, ack.SeqNo, 10)
					f.md.Put([]byte(commitIndexPrefix), b)
				}
			}

		}
	default:
		f.logger.Errorf("invalid request type, %v", req)
	}
}

func (f *ExeFiber) Stop() error {
	atomic.StoreInt32(&f.stop, 0)
	return nil
}
