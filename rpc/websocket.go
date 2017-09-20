package jsonrpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"hyperchain/common"
	"hyperchain/namespace"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	ReadBufferSize  = 1024 * 256
	WriteBufferSize = 1024 * 256
)

var (
	wsS internalRPCServer
)

type wsServerImpl struct {
	stopHp    chan bool
	restartHp chan bool
	nr        namespace.NamespaceManager
	port      int

	wsConns          map[*websocket.Conn]*Notifier
	wsConnsMux       sync.Mutex
	wsHandler        *Server
	wsListener       net.Listener
	wsAllowedOrigins []string // allowedOrigins should be a comma-separated list of allowed origin URLs.
	// To allow connections with any origin, pass "*".
}

type httpReadWriteCloser struct {
	io.Reader
	io.WriteCloser
}

func GetWSServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) internalRPCServer {
	if wsS == nil {
		wsS = &wsServerImpl{
			stopHp:           stopHp,
			restartHp:        restartHp,
			nr:               nr,
			wsAllowedOrigins: []string{"*"},
			wsConns:          make(map[*websocket.Conn]*Notifier),
			port:             nr.GlobalConfig().GetInt(common.WEBSOCKET_PORT),
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
	handler := NewServer(wssi.nr, wssi.stopHp, wssi.restartHp)
	if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", wssi.port)); err != nil {
		log.Errorf("%v", err)
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

		time.Sleep(4 * time.Second)
	}

	// todo this loop may be wrapped by a lock
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
func (wssi *wsServerImpl) newWSServer(srv *Server) *http.Server {
	return &http.Server{Handler: wssi.newWebsocketHandler(srv)}
}

// newWebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
func (wssi *wsServerImpl) newWebsocketHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		//if options&OptionSubscriptions == OptionSubscriptions {
		notifier := NewNotifier()
		ctx = context.WithValue(ctx, NotifierKey{}, notifier)
		notifier.subChs = common.GetSubChs(ctx)
		//}

		wssi.wsConnsMux.Lock()
		wssi.wsConns[conn] = notifier
		wssi.wsConnsMux.Unlock()

		defer func() {
			wssi.closeConnection(notifier, conn)
			cancel()
		}()

		for {
			_, nr, err := conn.NextReader()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Error(err)
				}
				break
			}

			nw, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				// TODO need write error
				log.Error(err)
				break
			}

			codec := NewJSONCodec(&httpReadWriteCloser{nr, nw}, r, srv.namespaceMgr, conn)
			notifier.codec = codec
			srv.ServeCodec(codec, OptionMethodInvocation|OptionSubscriptions, ctx)
		}
	}
}

func (wssi *wsServerImpl) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	if u, err := url.Parse(origin); err != nil {
		return false
	} else {
		for _, o := range wssi.wsAllowedOrigins {
			if o == "*" {
				return true
			} else if o == u.Host {
				return true
			}
		}
	}

	return false
}

func (wssi *wsServerImpl) closeConnection(notifier *Notifier, conn *websocket.Conn) {
	log.Debugf("cancel the context and close websocket connection %p, release resource", conn)

	notifier.Close()
	conn.Close()

	wssi.wsConnsMux.Lock()
	delete(wssi.wsConns, conn)
	wssi.wsConnsMux.Unlock()
}
