package jsonrpc

import (
	"github.com/rs/cors"
	"hyperchain/event"
	"net/http"
	"strconv"
	"io"
	"fmt"
	"hyperchain/hpc"
)

const (
	maxHTTPRequestContentLength = 1024 * 128
)

type httpReadWrite struct{
	io.Reader
	io.Writer
}

func (hrw *httpReadWrite) Close() error{
	return nil
}

func Start(httpPort int,eventMux *event.TypeMux) error{
	//log.Info("=============enter Start()=================")
	eventMux = eventMux

	server := NewServer()

	// 得到API，注册服务
	apis := hpc.GetAPIs()

	// api.Namespace 是API的命名空间，api.Service 是一个拥有命名空间对应对象的所有方法的对象
	for _, api := range apis {
		if err := server.RegisterName(api.Namespace, api.Service);err != nil {
			return err
		}
	}
	//go func() {
	//	log.Info("start the simple httpserver")
	//	cmd := exec.Command("pyhton $GOPATH/src/hyperchain/jsonrpc/Dashboard/simpleHttpServer.py")
	//	//cmd := exec.Command("pwd")
	//	cmd.Stdout = os.Stdout
	//	cmd.Stderr = os.Stderr
	//	log.Info(cmd.Run())
	//}()

	startHttp(httpPort, server)

	return nil
}

func startHttp(httpPort int, srv *Server) {
	//log.Info("=============enter startHttp()=================")
	// TODO AllowedOrigins should be a parameter
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET"},
	})

	// Insert the middleware
	handler := c.Handler(newJSONHTTPHandler(srv))

	http.ListenAndServe(":"+strconv.Itoa(httpPort),handler)
}

func newJSONHTTPHandler(srv *Server) http.HandlerFunc{
	//log.Info("=============enter newJSONHTTPHandler()=================")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxHTTPRequestContentLength {
			http.Error(w,
				fmt.Sprintf("content length too large (%d>%d)", r.ContentLength, maxHTTPRequestContentLength),
				http.StatusRequestEntityTooLarge)
			return
		}

		w.Header().Set("content-type", "application/json")

		// TODO NewJSONCodec
		codec := NewJSONCodec(&httpReadWrite{r.Body, w})
		defer codec.Close()
		srv.ServeSingleRequest(codec, OptionMethodInvocation)
	}
}


