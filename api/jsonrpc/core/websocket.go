package jsonrpc

import (
	"hyperchain/namespace"
	"hyperchain/common"
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
	"net"
	"io"
	"net/url"
)

const (
	ReadBufferSize = 1024 * 256
	WriteBufferSize = 1024 * 256
)

var (
	wsS          RPCServer

)

type wsServerImpl struct {
	nsMgr          namespace.NamespaceManager
	rpcServer      *Server
	allowedOrigins []string
}

type httpReadWriteCloser struct {
	io.Reader
	io.WriteCloser
}

func GetWSServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) RPCServer {
	if wsS == nil {
		wsS = newWSServer(nr, stopHp, restartHp)
	}
	return wsS
}

func newWSServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) *wsServerImpl {
	wssi := &wsServerImpl{
		nsMgr: nr,
		allowedOrigins: []string{"*"},
	}
	wssi.rpcServer = NewServer(nr, stopHp, restartHp)
	return wssi
}

func (wssi *wsServerImpl) Start() error{
	log.Notice("start websocket service ...")
	config := wssi.rpcServer.namespaceMgr.GlobalConfig()
	wsPort := config.GetInt(common.C_WEBSOCKET_PORT)

	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", wsPort)); err != nil {
		log.Errorf("%v",err)
		return err
	}
	go wssi.newServer(wssi.rpcServer).Serve(listener)
	log.Notice(fmt.Sprintf("WebSocket endpoint opened: ws://%s", fmt.Sprintf("%d", wsPort)))

	// All listeners booted successfully
	//n.wsEndpoint = endpoint
	//n.wsListener = listener
	//n.wsHandler = handler
	return nil
}

func (wssi *wsServerImpl) Stop() error {
	return nil
}

func (wssi *wsServerImpl) Restart() error {
	return nil
}

func (wssi *wsServerImpl) newServer(srv *Server) *http.Server {
	return &http.Server{Handler: wssi.newWebsocketHandler(srv)}
}

func (wssi *wsServerImpl) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	if u, err := url.Parse(origin); err != nil {
		return false
	} else {
		for _, o := range wssi.allowedOrigins {
			if o == "*" {
				return true
			} else if o == u.Host {
				return true
			}
		}
	}


	return  false
}

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

		defer conn.Close()

		for {
			_, nr, err := conn.NextReader()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Error(err)
				}
				// TODO 是否需要做错误返回？
				break
			}

			nw, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				// TODO 同上
				log.Error(err)
				break
			}

			codec := NewJSONCodec(&httpReadWriteCloser{nr, nw}, r, srv.namespaceMgr)

			srv.ServeCodec(codec, OptionMethodInvocation|OptionSubscriptions)
		}
	}
}
