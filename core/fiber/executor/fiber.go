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
	"github.com/hyperchain/hyperchain/manager/protos"
	"github.com/op/go-logging"
	"strconv"
	"sync/atomic"
	"time"
	"github.com/hyperchain/hyperchain/consensus/rbft"
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
	stop             chan struct{}
	eventMux         *event.TypeMux
}

func NewFiber(conf *common.Config, ns *service.NamespaceServices, ol oplog.OpLog, eventMux *event.TypeMux) (fiber.Fiber, error) {
	namespace := conf.GetString(common.NAMESPACE)
	if len(namespace) == 0 {
		return nil, fmt.Errorf("no namespace field found in config")
	}

	fr := &ExeFiber{
		conf:     conf,
		ns:       ns,
		ol:       ol,
		logger:   common.GetLogger(namespace, "fiber"),
		stop:     make(chan struct{}),
		eventMux: eventMux,
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
	go f.processExecutorRequest()
	for {
		select {
		case <-f.stop:
			f.logger.Notice("fiber exit")
			return nil
		default:
			f.sendNextConsume()
		}
	}

	return nil
}

func (f *ExeFiber) sendNextConsume() {
	var es service.Service
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
					return
				}

				if err := es.Send(logMsg); err != nil {
					f.logger.Error(err)
					return
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

func (f *ExeFiber) processExecutorRequest() {
	var es service.Service
	for {
		es = f.ns.Service(executorId)
		if es == nil {
			time.Sleep(time.Second)
		} else {
			select {
			case <-f.stop:
				return
			case req := <-es.Receive():
				f.handle(req)
			}
		}
	}
}

func (f *ExeFiber) handle(req *pb.IMessage) {
	switch req.Event {
	case pb.Event_OpLogFetch:
		fetch := &event.OpLogFetch{}
		err := proto.Unmarshal(req.Payload, fetch)
		if err != nil {
			f.logger.Error(err)
		} else {
			le, err := f.ol.Fetch(fetch.LogID)
			f.logger.Infof("fiber fetch log entry id with %v", le.Lid)
			if err != nil {
				f.logger.Errorf("fetch log entry with id %v, error %v", fetch.LogID, err)
			}
			payload, err := proto.Marshal(le)
			if err != nil {
				f.logger.Errorf("marshal for log entry error: %v", err)
			}
			msg := &pb.IMessage{
				Id:      req.Id,
				Type:    pb.Type_RESPONSE,
				Payload: payload,
			}
			f.ns.Service(executorId).Send(msg)
		}
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
	case pb.Event_Checkpoint:
		blki := &protos.BlockchainInfo{}
		if err := proto.Unmarshal(req.Payload, blki); err != nil {
			f.logger.Error(err)
			return
		}
		payload, err := proto.Marshal(blki)
		if err != nil {
			f.logger.Errorf("Marshal AddNode Error!")
			return
		}
		chkptEvent := &rbft.LocalEvent{
			Service:   rbft.CORE_RBFT_SERVICE,
			EventType: rbft.CORE_EXECUTOR_CHECKPOINT_EVENT,
			Event:     payload,
		}
		go f.eventMux.Post(chkptEvent)
	default:
		f.logger.Errorf("invalid request type, %v", req)
	}
}

func (f *ExeFiber) Stop() {
	if f.stop != nil {
		close(f.stop)
	}
	f.stop = nil
	f.logger.Notice("stop fiber successfully")
}

//Send info to the remote peer
func (f *ExeFiber) Send(interface{}) error {
	//TODO(Xiaoyi Wang): send to remote peer
	return nil
}
