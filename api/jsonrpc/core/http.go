//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/rs/cors"
	"hyperchain/api/rest/routers"
	"hyperchain/common"
	"hyperchain/namespace"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	maxHTTPRequestContentLength = 1024 * 256
)

var server *Server

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

func Start(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) error {
	server = NewServer(nr, stopHp, restartHp)
	startHttp(server)
	return nil
}

func Stop() {
	server.Stop()
}

func startHttp(srv *Server) {
	config := srv.namespaceMgr.GlobalConfig()

	httpPort := config.GetInt(common.C_HTTP_PORT)
	restPort := config.GetInt(common.C_REST_PORT)
	logsPath := config.GetString(common.LOG_DUMP_FILE_DIR)

	// TODO AllowedOrigins should be a parameter
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET"},
	})

	// Insert the middleware
	handler := c.Handler(newJSONHTTPHandler(srv))

	log.Debugf("start to listen http port: %d", httpPort)
	go http.ListenAndServe(":"+strconv.Itoa(httpPort), handler)

	// rest service
	routers.NewRouter()
	beego.BConfig.CopyRequestBody = true
	beego.SetLogFuncCall(true)

	logs.SetLogger(logs.AdapterFile, `{"filename": "`+logsPath+"/RESTful-API-"+strconv.Itoa(restPort)+"-"+time.Now().Format("2006-01-02 15:04:05")+`"}`)
	beego.BeeLogger.DelLogger("console")

	beego.Run("0.0.0.0:" + strconv.Itoa(restPort))

}

func newJSONHTTPHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxHTTPRequestContentLength {
			http.Error(w,
				fmt.Sprintf("content length too large (%d>%d)", r.ContentLength, maxHTTPRequestContentLength),
				http.StatusRequestEntityTooLarge)
			return
		}
		w.Header().Set("content-type", "application/json")
		codec := NewJSONCodec(&httpReadWrite{r.Body, w}, r, srv.namespaceMgr)
		defer codec.Close()
		srv.ServeSingleRequest(codec, OptionMethodInvocation)
	}
}
