//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/rs/cors"
	"hyperchain/api"
	"hyperchain/api/rest_api/routers"
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/membersrvc"
	"io"
	"net/http"
	"strconv"
	"time"
	"hyperchain/common"
)

const (
	maxHTTPRequestContentLength = 1024 * 256
)

type RateLimitConfig struct {
	Enable           bool
	TxFillRate       time.Duration
	TxRatePeak       int64
	ContractFillRate time.Duration
	ContractRatePeak int64
}
type httpReadWrite struct {
	io.Reader
	io.Writer
}

func (hrw *httpReadWrite) Close() error {
	return nil
}

func Start(httpPort int, restPort int, logsPath string, eventMux *event.TypeMux, pm *manager.ProtocolManager, cm *membersrvc.CAManager, config *common.Config) error {
	eventMux = eventMux

	server := NewServer()

	// 得到API，注册服务
	apis := hpc.GetAPIs(eventMux, pm, cm, config)

	// api.Namespace 是API的命名空间，api.Service 是一个拥有命名空间对应对象的所有方法的对象
	for _, api := range apis {
		if err := server.RegisterName(api.Namespace, api.Service); err != nil {
			log.Errorf("registerName error: %v ", err)
			return err
		}
	}

	startHttp(httpPort, restPort, logsPath, server, cm)

	return nil
}

func startHttp(httpPort int, restPort int, logsPath string, srv *Server, cm *membersrvc.CAManager) {
	// TODO AllowedOrigins should be a parameter
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET"},
	})

	// Insert the middleware
	handler := c.Handler(newJSONHTTPHandler(srv, cm))

	go http.ListenAndServe(":"+strconv.Itoa(httpPort), handler)

	// ===================================== 2016.11.15 START ================================ //
	routers.NewRouter()
	beego.BConfig.CopyRequestBody = true
	beego.SetLogFuncCall(true)

	logs.SetLogger(logs.AdapterFile, `{"filename": "`+logsPath+"/RESTful-API-"+strconv.Itoa(restPort)+"-"+time.Now().Format("2006-01-02 15:04:05")+`"}`)
	beego.BeeLogger.DelLogger("console")

	// todo 读取　app.conf　配置文件
	// the first param adapterName is ini/json/xml/yaml.
	//beego.LoadAppConfig("ini", "jsonrpc/RESTful_api/conf/app.conf")
	beego.Run("127.0.0.1:" + strconv.Itoa(restPort))
	// ===================================== 2016.11.15 END  ================================ //
}

func newJSONHTTPHandler(srv *Server, cm *membersrvc.CAManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Critical("has request")
		if r.ContentLength > maxHTTPRequestContentLength {
			http.Error(w,
				fmt.Sprintf("content length too large (%d>%d)", r.ContentLength, maxHTTPRequestContentLength),
				http.StatusRequestEntityTooLarge)
			return
		}
		//header interceptor
		//headerHandler(w,r)

		w.Header().Set("content-type", "application/json")

		// TODO NewJSONCodec
		codec := NewJSONCodec(&httpReadWrite{r.Body, w}, r.Header, cm)
		defer codec.Close()
		srv.ServeSingleRequest(codec, OptionMethodInvocation)
	}
}
