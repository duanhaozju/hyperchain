package api

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"hyperchain/common"
	hrpc "hyperchain/rpc"
	hm "hyperchain/service/hypexec/controller"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	maxHTTPRequestContentLength = 1024 * 256
	ReadTimeout                 = 5 * time.Second
)

var (
	hs internalRPCServer
)

type httpServerImpl struct {
	er     hm.ExecutorController
	port   int
	config *common.Config

	httpListener net.Listener
	httpHandler  *ServerEx
}

// GetHttpServer creates and returns a new httpServerImpl instance implements internalRPCServer interface.
func GetHttpServer(er hm.ExecutorController, config *common.Config) internalRPCServer {
	if hs == nil {
		hs = &httpServerImpl{
			er:     er,
			port:   config.GetInt(common.JSON_RPC_PORT_EXECUTOR),
			config: config,
		}
	}
	return hs
}

// start starts the http RPC endpoint. It will start the appropriate server based on the parameters of the configuration file.
func (hsi *httpServerImpl) start() error {

	var (
		listener net.Listener
		srv      *http.Server
		err      error
	)

	handler := NewServerEx(hsi.er, hsi.config)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handler.admin.LoginServer)
	mux.Handle("/", newCorsHandler(handler, hsi.config.GetStringSlice(common.HTTP_ALLOWEDORIGINS)))

	isVersion2 := hsi.config.GetBool(common.HTTP_VERSION2)
	isHTTPS := hsi.config.GetBool(common.HTTP_SECURITY)

	// prepare tls.Config
	tlsConfig, err := hsi.secureConfig()
	if err != nil {
		return err
	}

	if isVersion2 && isHTTPS {

		// http2, https
		log.Noticef("starting http/2 service at port %v ... , secure connection is enabled.", hsi.port)

		// start http listener with secure connection and http/2
		tlsConfig.NextProtos = []string{"h2"}
		if listener, err = tls.Listen("tcp", fmt.Sprintf(":%d", hsi.port), tlsConfig); err != nil {
			return err
		}

		srv = newHTTPServer(mux, tlsConfig)
		http2.ConfigureServer(srv, &http2.Server{})

	} else if !isVersion2 && isHTTPS {

		// http1.1, https
		log.Noticef("starting http/1.1 service at port %v ... , secure connection is enabled.", hsi.port)

		// start http listener with secure connection and http/1.1
		if listener, err = tls.Listen("tcp", fmt.Sprintf(":%d", hsi.port), tlsConfig); err != nil {
			return err
		}

		srv = newHTTPServer(mux, tlsConfig)

	} else {

		// http1.1, disable https
		log.Noticef("starting http/1.1 service at port %v ... , secure connection is disenabled.", hsi.port)

		// start http listener with http/1.1
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", hsi.port))
		if err != nil {
			return err
		}

		srv = newHTTPServer(mux, nil)

	}

	go srv.Serve(listener)

	hsi.httpListener = listener
	hsi.httpHandler = handler

	return nil
}

// stop stops the http RPC endpoint.
func (hsi *httpServerImpl) stop() error {
	log.Noticef("stopping http service at port %v ...", hsi.port)
	if hsi.httpListener != nil {
		hsi.httpListener.Close()
		hsi.httpListener = nil
	}

	if hsi.httpHandler != nil {
		hsi.httpHandler.Stop()
		hsi.httpHandler = nil

		// wait for ReadTimeout to close all the established connection
		time.Sleep(ReadTimeout)
	}

	log.Notice("http service stopped")
	return nil
}

// restart restarts the http RPC endpoint.
func (hsi *httpServerImpl) restart() error {
	log.Noticef("restarting http service at port %v ...", hsi.port)
	if err := hsi.stop(); err != nil {
		return err
	}
	if err := hsi.start(); err != nil {
		return err
	}
	return nil
}

func (hsi *httpServerImpl) getPort() int {
	return hsi.port
}

func (hsi *httpServerImpl) setPort(port int) error {
	if port == 0 {
		return errors.New("please offer http port")
	}
	hsi.port = port
	return nil
}

func (hsi *httpServerImpl) secureConfig() (*tls.Config, error) {

	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(hsi.config.GetString(common.P2P_TLS_CA))
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return nil, err
	}
	pool.AppendCertsFromPEM(caCrt)

	serverCert, err := tls.LoadX509KeyPair(hsi.config.GetString(common.P2P_TLS_CERT), hsi.config.GetString(common.P2P_TLS_CERT_PRIV))
	if err != nil {
		log.Errorf("Loadx509keypair err: ", err)
		return nil, err
	}
	tlsConfig := &tls.Config{
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
	}

	return tlsConfig, nil
}

// newHTTPServer creates a new http RPC server around an API provider.
func newHTTPServer(mux *http.ServeMux, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Handler:     mux,
		ReadTimeout: ReadTimeout,
		TLSConfig:   tlsConfig,
	}
}

func newCorsHandler(srv *ServerEx, allowedOrigins []string) http.Handler {
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

type httpReadWrite struct {
	io.Reader
	io.Writer
}

func (hrw *httpReadWrite) Close() error { return nil }

// ServeHTTP serves JSON-RPC requests over HTTP.
func (srv *ServerEx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > maxHTTPRequestContentLength {
		http.Error(w,
			fmt.Sprintf("content length too large (%d>%d)", r.ContentLength, maxHTTPRequestContentLength),
			http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("content-type", "application/json")
	codec := NewJSONCodec(&httpReadWrite{r.Body, w}, r, srv.er, nil)
	defer codec.Close()
	srv.ServeSingleRequest(codec, hrpc.OptionMethodInvocation)
}
