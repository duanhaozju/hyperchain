package admin

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	pb "github.com/hyperchain/hyperchain/common/protos"
	"github.com/hyperchain/hyperchain/common/service/client"
	"github.com/hyperchain/hyperchain/service/hypexec/controller"
	"github.com/op/go-logging"
	"time"
)

type Administrator struct {
	exeCtl      controller.ExecutorController
	adminClient *client.ServiceClient
	conf        *common.Config
	logger      *logging.Logger
	stop        chan struct{}
}

func NewAdministrator(exeCtl controller.ExecutorController, conf *common.Config) *Administrator {
	return &Administrator{
		exeCtl: exeCtl,
		conf:   conf,
		logger: common.GetLogger(common.DEFAULT_LOG, "admin"),
		stop:   make(chan struct{}, 2),
	}
}

func (admin *Administrator) Start() error {
	//client address for mark this admin connect
	address := admin.conf.GetString(common.EXECUTOR_HOST_ADDR)

	adminClient, err := client.New(admin.conf.GetInt(common.INTERNAL_PORT),
		admin.conf.GetString(common.EXECUTOR_SERVER_IP), client.ADMINISTRATOR, address)

	if err != nil {
		return err
	}

	admin.logger.Info("try connect to hyperchain...")
	for {
		//TODO: add retry times limit
		err = adminClient.Connect()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	rsp, err := adminClient.Register(0, pb.FROM_ADMINISTRATOR, &pb.RegisterMessage{
		Address: address,
	})

	if err != nil {
		return err
	}

	if len(rsp.Payload) != 0 {
		if err := admin.exeCtl.StartAllExecutorSrvs(); err != nil {
			return err
		}
	}

	admin.logger.Info("connect to hyperchain successful! ")

	h := NewAdminHandler(admin.exeCtl)
	adminClient.AddHandler(h)
	admin.adminClient = adminClient
	go admin.listenSendResponse(h, adminClient)
	return nil
}

func (admin *Administrator) Stop() {
	admin.adminClient.Close()
	admin.stop <- struct{}{}
}

func (admin *Administrator) listenSendResponse(e *AdminHandler, adminConnect *client.ServiceClient) {
	for {
		select {
		case ev := <-e.Ch:
			payload, _ := proto.Marshal(ev.are)
			err := adminConnect.Send(&pb.IMessage{
				Id:      ev.rspId, //TODO: Fix it
				Type:    pb.Type_RESPONSE,
				Ok:      ev.are.Ok,
				Payload: payload,
			})
			if err != nil {
				// TODO : how to deal with the send failed?
			}
		case <-admin.stop:
			return
		}
	}
}
