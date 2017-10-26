package admin

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/common/client"
	pb "hyperchain/common/protos"
	"hyperchain/service/executor/manager"
	"time"
)

type Administrator struct {
	ecMgr       manager.ExecutorManager
	adminClient *client.ServiceClient
	conf        *common.Config
	logger      *logging.Logger
}

func NewAdministrator(ecMgr manager.ExecutorManager, conf *common.Config) *Administrator {
	return &Administrator{
		ecMgr: ecMgr,
		conf:  conf,
		//logger: common.GetLogger(common.DEFAULT_LOG, "admin"),
		logger: logging.MustGetLogger("executorAdmin"),
	}
}

func (admin *Administrator) Start() error {
	//TODO : wait for "namespace" param delete
	address := admin.conf.GetString(common.EXECUTOR_HOST_ADDR)

	adminClient, err := client.New(admin.conf.GetInt(common.INTERNAL_PORT), "127.0.0.1", client.EXECUTOR, "global")
	if err != nil {
		return err
	}

	admin.logger.Info("connecting hyperchain...")
	for {
		err = adminClient.Connect()
		if err == nil {
			break
		}
		d, _ := time.ParseDuration(fmt.Sprintf("%ds", 1))
		time.Sleep(d)
	}
	admin.logger.Info("connected hyperchain.")

	admin.logger.Info("registering admin client...")
	err = adminClient.Register(pb.FROM_ADMINISTRATOR, &pb.RegisterMessage{
		Address: address,
	})
	if err != nil {
		return err
	}
	admin.logger.Info("registered admin client.")

	h := NewAdminHandler(admin.ecMgr)
	adminClient.AddHandler(h)
	admin.adminClient = adminClient
	go listenSendResponse(h, adminClient)
	return nil
}

func (admin *Administrator) Stop() {

}

func listenSendResponse(e *AdminHandler, adminConnect *client.ServiceClient) {
	for {
		select {
		case ev := <-e.Ch:
			payload, _ := proto.Marshal(ev)
			err := adminConnect.Send(&pb.IMessage{
				Type:    pb.Type_RESPONSE,
				Ok:      true,
				Payload: payload, //TODO: refactor payload
			})
			if err != nil {
				//logger.Errorf("adminclient %s Send message to hyperchain filed", "IP")
				// check log
				// TODO : how to deal with the send failed?
			}
		}
	}
}
