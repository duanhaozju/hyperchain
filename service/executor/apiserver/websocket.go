package apiserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"hyperchain/common"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	hm "hyperchain/service/executor/manager"
	hrpc "hyperchain/rpc"
)

const (
	ReadBufferSize  = 1024 * 256
	WriteBufferSize = 1024 * 256
)

var (
	wsS internalRPCServer
)

type wsServerImpl struct {
	er     hm.ExecutorManager
	port   int
	config *common.Config

	wsConns    map[*websocket.Conn]*hrpc.Notifier
	wsConnsMux sync.Mutex
	wsHandler  *ServerEx
	wsListener net.Listener
}

type httpReadWriteCloser struct {
	io.Reader
	io.WriteCloser
}

// GetWSServer creates and returns a new wsServerImpl instance implements internalRPCServer interface.
func GetWSServer(er hm.ExecutorManager, config *common.Config) internalRPCServer {
	if wsS == nil {
		wsS = &wsServerImpl{
			er:      er,
			wsConns: make(map[*websocket.Conn]*hrpc.Notifier),
			port:    config.GetInt(common.WEBSOCKET_PORT),
			config:  config,
		}
	}
	return wsS
}

// start starts the websocket RPC endpoint.
func (wssi *wsServerImpl) start() error {
	log.Noticef("starting websocket service at port %v ...", wssi.port)

	var (
		listener net.Listener
		err      error
	)

	// start websocket listener
	handler := NewServerEx(wssi.er, wssi.config)
	if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", wssi.port)); err != nil {
		return err
	}
	go wssi.newWSServer(handler).Serve(listener)

	wssi.wsListener = listener
	wssi.wsHandler = handler

	return nil
}

// stop terminates the websocket RPC endpoint.
func (wssi *wsServerImpl) stop() error {
	log.Noticef("stopping websocket service at port %v ...", wssi.port)
	if wssi.wsListener != nil {
		wssi.wsListener.Close()
		wssi.wsListener = nil
	}

	if wssi.wsHandler != nil {
		wssi.wsHandler.Stop()
		wssi.wsHandler = nil
	}

	// todo this loop may be wrapped by a lock?
	// close all the opened connection, and release its resource
	for c, n := range wssi.wsConns {
		wssi.closeConnection(n, c)
	}

	log.Notice("websocket service stopped")

	return nil
}

// restart restarts the websocket RPC endpoint.
func (wssi *wsServerImpl) restart() error {
	log.Noticef("restarting websocket service at port %v ...", wssi.port)
	if err := wssi.stop(); err != nil {
		return err
	}
	if err := wssi.start(); err != nil {
		return err
	}
	return nil
}

func (wssi *wsServerImpl) getPort() int {
	return wssi.port
}

func (wssi *wsServerImpl) setPort(port int) error {
	if port == 0 {
		return errors.New("please offer websocket port")
	}
	wssi.port = port
	return nil
}

// newWSServer creates a new websocket RPC server around an API provider.
func (wssi *wsServerImpl) newWSServer(srv *ServerEx) *http.Server {
	return &http.Server{Handler: wssi.newWebsocketHandler(srv)}
}

// newWebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
func (wssi *wsServerImpl) newWebsocketHandler(srv *ServerEx) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// upgrading an HTTP connection to a WebSocket connection
		upgrader := websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return wssi.isOriginAllowed(origin)
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("new websocket connection %p", conn)

		ctx, cancel := context.WithCancel(context.Background())

		// create a notifier to push message for the connection
		notifier := hrpc.NewNotifier()
		ctx = context.WithValue(ctx, hrpc.NotifierKey{}, notifier)
		notifier.SetSubChs(common.GetSubChs(ctx))

		wssi.wsConnsMux.Lock()
		wssi.wsConns[conn] = notifier
		wssi.wsConnsMux.Unlock()

		defer func() {
			wssi.closeConnection(notifier, conn)
			cancel()
		}()

		for {
			// waitting for new message from client
			_, nr, err := conn.NextReader()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Error(err)
				}
				break
			}

			nw, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Error(err)
				break
			}

			codec := NewJSONCodec(&httpReadWriteCloser{nr, nw}, r, srv.er, conn)
			notifier.SetCodec(codec)
			srv.ServeCodec(codec, hrpc.OptionMethodInvocation|hrpc.OptionSubscriptions, ctx)
		}
	}
}

func (wssi *wsServerImpl) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	allowedOrigins := wssi.config.GetStringSlice(common.HTTP_ALLOWEDORIGINS)

	if u, err := url.Parse(origin); err != nil {
		return false
	} else {
		for _, o := range allowedOrigins {
			if o == "*" {
				return true
			} else if o == u.Host {
				return true
			}
		}
	}

	return false
}

func (wssi *wsServerImpl) closeConnection(notifier *hrpc.Notifier, conn *websocket.Conn) {
	log.Debugf("cancel the context and close websocket connection %p, release resource", conn)

	notifier.Close()
	conn.Close()

	wssi.wsConnsMux.Lock()
	delete(wssi.wsConns, conn)
	wssi.wsConnsMux.Unlock()
}

