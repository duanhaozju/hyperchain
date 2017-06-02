//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	//"github.com/astaxie/beego"
	//"github.com/astaxie/beego/logs"
	"github.com/rs/cors"

	//"hyperchain/api/rest/routers"
	"hyperchain/common"
	"hyperchain/namespace"

	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	//"strconv"
	"time"
)

const (
	maxHTTPRequestContentLength = 1024 * 256
)

var (
	StoppedError = errors.New("Listener stopped")
	hs           RPCServer
)

type httpServerImpl struct {
	nsMgr        namespace.NamespaceManager
	rpcServer    *Server
	stop         chan bool
	stopListener *StoppableListener
}

func GetHttpServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) RPCServer {
	if hs == nil {
		hs = newHttpServer(nr, stopHp, restartHp)
	}
	return hs
}

func newHttpServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) *httpServerImpl {
	hi := &httpServerImpl{
		nsMgr: nr,
		stop:  make(chan bool),
	}
	hi.rpcServer = NewServer(nr, stopHp, restartHp)
	return hi
}

//Start start the http service, this method will block if the http
//service start successful.
func (hi *httpServerImpl) Start() error {
	log.Notice("start http service ...")
	config := hi.rpcServer.namespaceMgr.GlobalConfig()
	httpPort := config.GetInt(common.C_HTTP_PORT)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET"},
	})
	hi.rpcServer.Start()
	handler := c.Handler(newHttpHandler(hi.rpcServer))

	//t1 := time.Now()
	var err error
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		return err
	}
	hi.stopListener, err = NewListener(listener)
	if err != nil {
		return err
	}

	go http.Serve(hi.stopListener, handler)
	// TODO 可能要换成下面的代码
	//err = http.Serve(hi.stopListener, handler)

	//t2 := time.Now()
	//if err != nil && t2.Sub(t1) < 2*time.Second {
	//	log.Errorf("start http service error %v", err)
	//	return err
	//}

	log.Critical("http service closed and release the binding port")
	return nil
}

//Stop stop the http service, this method will wait for 3 seconds before stop.
func (hi *httpServerImpl) Stop() error {
	//TODO:Fix it, this stop method will not stop...
	log.Notice("stop http service ...")
	if hi.stopListener != nil {
		hi.stopListener.Stop()
		hi.stopListener.Close()
	}
	hi.rpcServer.Stop()
	time.Sleep(4 * time.Second)
	log.Notice("stopped http service")
	return nil
}

//Restart restart the http service, this method will retry 5 times
//until it start successful
func (hi *httpServerImpl) Restart() error {
	hi.Stop()
	go func() {
		maxRetry := uint(5)
		var i uint
		for i = 1; i <= maxRetry; i++ {
			log.Noticef("%d retry start http service ...", i)
			beforeStart := time.Now()
			err := hi.Start()
			if err != nil {
				afterStart := time.Now()
				if afterStart.Sub(beforeStart) < 2*time.Second {
					log.Errorf("Restart http error %v", err)
					time.Sleep((1 << i) * time.Second)
				} else {
					return
				}
			}
		}
		if i > maxRetry {
			log.Errorf("Restart times overflow!")
		}
	}()
	return nil
}

type httpReadWrite struct {
	io.Reader
	io.Writer
}

func (hrw *httpReadWrite) Close() error {
	return nil
}

func newHttpHandler(srv *Server) http.HandlerFunc {
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

//func startRestService(srv *Server) {
//	config := srv.namespaceMgr.GlobalConfig()
//	restPort := config.GetInt(common.C_REST_PORT)
//	logsPath := config.GetString(common.LOG_DUMP_FILE_DIR)
//
//	// rest service
//	routers.NewRouter()
//	beego.BConfig.CopyRequestBody = true
//	beego.SetLogFuncCall(true)
//
//	logs.SetLogger(logs.AdapterFile, `{"filename": "`+logsPath+"/RESTful-API-"+strconv.Itoa(restPort)+"-"+time.Now().Format("2006-01-02 15:04:05")+`"}`)
//	beego.BeeLogger.DelLogger("console")
//
//	beego.Run("0.0.0.0:" + strconv.Itoa(restPort))
//}

type StoppableListener struct {
	*net.TCPListener          //Wrapped listener
	stop             chan int //Channel used only to indicate listener should shutdown
}

func NewListener(l net.Listener) (*StoppableListener, error) {
	tcpL, ok := l.(*net.TCPListener)

	if !ok {
		return nil, errors.New("Cannot wrap listener")
	}
	retval := &StoppableListener{}
	retval.TCPListener = tcpL
	retval.stop = make(chan int)

	return retval, nil
}

func (sl *StoppableListener) Accept() (net.Conn, error) {
	for {
		//Wait up to one second for a new connection
		sl.SetDeadline(time.Now().Add(time.Second))
		newConn, err := sl.TCPListener.Accept()

		//Check for the channel being closed
		select {
		case <-sl.stop:
			log.Errorf("try close %v", err)
			if err == nil {
				log.Errorf("close")
				err = newConn.Close()
				log.Errorf("closed %v", err)
			}
			return nil, StoppedError
		default:
			//If the channel is still open, continue as normal
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		return newConn, err
	}
}

func (sl *StoppableListener) Stop() {
	close(sl.stop)
}
