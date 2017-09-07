//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	//"github.com/astaxie/beego"
	//"github.com/astaxie/beego/logs"
	"github.com/rs/cors"

	//"hyperchain/api/rest/routers"
	"hyperchain/namespace"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
	admin "hyperchain/api/admin"
	"errors"
	"hyperchain/common"
	"crypto/x509"
	"io/ioutil"
	"crypto/tls"
	"golang.org/x/net/http2"
)

const (
	maxHTTPRequestContentLength = 1024 * 256
	ReadTimeout	            = 3 * time.Second
)

var (
	hs           internalRPCServer
)

type httpServerImpl struct {
	stopHp			chan bool
	restartHp		chan bool
	nr			namespace.NamespaceManager
	port                    int

	httpListener 		net.Listener
	httpHandler  		*Server
	httpAllowedOrigins 	[]string
}

func GetHttpServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) internalRPCServer {
	if hs == nil {
		hs = &httpServerImpl{
			nr: 			nr,
			stopHp: 		stopHp,
			restartHp: 		restartHp,
			httpAllowedOrigins: 	[]string{"*"},
			port:                   nr.GlobalConfig().GetInt(common.JSON_RPC_PORT),
		}
	}
	return hs
}

// start starts the http RPC endpoint.
func (hi *httpServerImpl) start() error {

	var (
		listener net.Listener
		err      error
	)

	config := hi.nr.GlobalConfig()

	// start http listener
	handler := NewServer(hi.nr, hi.stopHp, hi.restartHp)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", admin.LoginServer)
	mux.Handle("/", newCorsHandler(handler, hi.httpAllowedOrigins))

	if config.GetBool(common.HTTP_SECURITY) {
		// enable https
		log.Noticef("starting http service at port %v ... , security connection is enabled.", hi.port)

		pool := x509.NewCertPool()

		caCrt, err := ioutil.ReadFile(config.GetString(common.P2P_TLS_CA))
		if err != nil {
			fmt.Println("ReadFile err:", err)
			return err
		}
		pool.AppendCertsFromPEM(caCrt)

		srv := newHTTPServer(mux, &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		})
		srv.Addr = ":"+config.GetString(common.JSON_RPC_PORT)
		http2.ConfigureServer(srv, &http2.Server{})
		go func() {
			err = srv.ListenAndServeTLS(config.GetString(common.P2P_TLS_CERT), config.GetString(common.P2P_TLS_CERT_PRIV))
			if err != nil {
				log.Errorf("ListenAndServeTLS error: %v", err)
			}
		}()
		//TODO 需要支持手动关闭安全http服务

		//serverCert, err := tls.LoadX509KeyPair(config.GetString(common.P2P_TLS_CERT), config.GetString(common.P2P_TLS_CERT_PRIV))
		//if err != nil {
		//	fmt.Println("Loadx509keypair err:", err)
		//	return err
		//}
		//listener, err = tls.Listen("tcp", ":"+config.GetString(common.JSON_RPC_PORT), &tls.Config{
		//	ClientCAs:  pool,
		//	ClientAuth: tls.RequireAndVerifyClientCert,
		//	Certificates: []tls.Certificate{serverCert},
		//})
		//if err != nil {
		//	fmt.Println(err)
		//	return err
		//}
		//go srv.Serve(listener)
	} else {
		// disable https
		log.Noticef("starting http service at port %v ... , security connection is disenabled.", hi.port)
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", hi.port))
		if err != nil {
			return err
		}
		go newHTTPServer(mux, nil).Serve(listener)
	}

	hi.httpListener = listener
	hi.httpHandler = handler

	return nil
}

// stop stops the http RPC endpoint.
func (hi *httpServerImpl) stop() error {
	log.Noticef("stopping http service at port %v ...", hi.port)
	if hi.httpListener != nil {
		hi.httpListener.Close()
		hi.httpListener = nil
	}

	if hi.httpHandler != nil {
		hi.httpHandler.Stop()
		hi.httpHandler = nil
		time.Sleep(4 * time.Second)
	}

	log.Notice("http service stopped")
	return nil
}

// restart restarts the http RPC endpoint.
func (hi *httpServerImpl) restart() error {
	log.Noticef("restarting http service at port %v ...", hi.port)
	if err := hi.stop(); err != nil {
		return err
	}
	if err := hi.start(); err != nil {
		return err
	}
	return nil
}

func(hi *httpServerImpl) getPort() int {
	return hi.port
}

func (hi *httpServerImpl) setPort(port int) error {
	if port == 0 {
		return errors.New("please offer http port")
	}
	hi.port = port
	return nil
}

type httpReadWrite struct {
	io.Reader
	io.Writer
}

func (hrw *httpReadWrite) Close() error {
	return nil
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

// newHTTPServer creates a new http RPC server around an API provider.
func newHTTPServer(mux *http.ServeMux, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Handler: mux,
		ReadTimeout:  ReadTimeout,
		TLSConfig: tlsConfig,
	}
}

func newCorsHandler(srv *Server, allowedOrigins []string) http.Handler {
	// disable CORS support if user has not specified a custom CORS configuration
	if len(allowedOrigins) == 0 {
		return srv
	}

	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"POST", "GET"},
		MaxAge:         600,
		AllowedHeaders: []string{"*"},
	})
	return c.Handler(srv)
}

// ServeHTTP serves JSON-RPC requests over HTTP.
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > maxHTTPRequestContentLength {
		http.Error(w,
			fmt.Sprintf("content length too large (%d>%d)", r.ContentLength, maxHTTPRequestContentLength),
			http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("content-type", "application/json")
	codec := NewJSONCodec(&httpReadWrite{r.Body, w}, r, srv.namespaceMgr, nil)
	defer codec.Close()
	srv.ServeSingleRequest(codec, OptionMethodInvocation)
}